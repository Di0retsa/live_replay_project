package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"live_replay_project/backend/common/enum"
	"live_replay_project/backend/common/retcode"
	"live_replay_project/backend/global"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
)

var imageFiles []string

func LoadImagePaths() {
	path := filepath.Join(global.CWD, "/static/images")
	files, err := os.ReadDir(path)
	if err != nil {
		global.Logger.Fatal(err, "loadImagePaths Failed")
		return
	}
	for _, file := range files {
		if !file.IsDir() {
			ext := filepath.Ext(file.Name())
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
				imageFiles = append(imageFiles, filepath.Join(path, file.Name()))
			}
		}
	}
	if len(imageFiles) == 0 {
		global.Logger.Fatal(err, "目录为空")
	}
}

func GenerateCaptcha() (string, string, string, int, error) {
	if len(imageFiles) == 0 {
		return "", "", "", 0, fmt.Errorf("没有可用的背景图片")
	}

	randIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(imageFiles))))
	imgPath := imageFiles[randIdx.Int64()]

	file, err := os.Open(imgPath)
	if err != nil {
		return "", "", "", 0, fmt.Errorf("打开图片失败: %v", err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return "", "", "", 0, fmt.Errorf("解码图片失败: %v", err)
	}
	global.Logger.Info(fmt.Sprintf("使用图片: %s, 格式: %s, 尺寸: %dx%d", imgPath, format, img.Bounds().Dx(), img.Bounds().Dy()))

	// 确保图片尺寸足够大
	if img.Bounds().Dx() < enum.PuzzleWidth*2 || img.Bounds().Dy() < enum.PuzzleHeight {
		return "", "", "", 0, fmt.Errorf("图片太小，无法创建拼图")
	}

	// 随机生成拼图块位置 (确保拼图块在图片内，并且右侧有足够空间显示挖孔)
	maxPuzzleX := img.Bounds().Dx() - enum.PuzzleWidth - 10 // 减去拼图宽度并留一些边距
	if maxPuzzleX <= enum.PuzzleWidth {                     // 确保挖孔位置和拼图本身不重叠太多
		maxPuzzleX = img.Bounds().Dx() / 2
	}
	randXBig, _ := rand.Int(rand.Reader, big.NewInt(int64(maxPuzzleX-enum.PuzzleWidth))) // X轴起始点至少是puzzleWidth
	puzzleX := int(randXBig.Int64()) + enum.PuzzleWidth/2                                // 保证左边有一定空间

	randYBig, _ := rand.Int(rand.Reader, big.NewInt(int64(img.Bounds().Dy()-enum.PuzzleHeight)))
	puzzleY := int(randYBig.Int64())

	// 裁剪拼图块
	puzzleRect := image.Rect(puzzleX, puzzleY, puzzleX+enum.PuzzleWidth, puzzleY+enum.PuzzleHeight)
	puzzlePiece := image.NewRGBA(puzzleRect) // 创建一个新的 RGBA 图片来存储裁剪的区域
	draw.Draw(puzzlePiece, puzzlePiece.Bounds(), img, image.Point{X: puzzleX, Y: puzzleY}, draw.Src)

	// 创建带挖孔的背景图
	backgroundWithHole := image.NewRGBA(img.Bounds())
	draw.Draw(backgroundWithHole, img.Bounds(), img, image.Point{}, draw.Src)

	// 在背景图上挖孔
	fillColor := color.RGBA{R: 0, G: 0, B: 0, A: 180}
	holeColor := image.NewUniform(fillColor) // 半透明黑色
	draw.Draw(backgroundWithHole, puzzleRect, holeColor, image.Point{}, draw.Src)

	// 将图片转为 Base64
	var puzzleBuf bytes.Buffer
	if err := jpeg.Encode(&puzzleBuf, puzzlePiece, nil); err != nil {
		return "", "", "", 0, fmt.Errorf("编码拼图块失败: %v", err)
	}
	puzzleBase64 := base64.StdEncoding.EncodeToString(puzzleBuf.Bytes())

	var bgBuf bytes.Buffer
	if err := jpeg.Encode(&bgBuf, backgroundWithHole, nil); err != nil {
		return "", "", "", 0, fmt.Errorf("编码背景图失败: %v", err)
	}
	bgBase64 := base64.StdEncoding.EncodeToString(bgBuf.Bytes())

	// 生成 CAPTCHA ID 并存储, 设置过期时间为一分钟
	captchaID := uuid.New().String()
	conn := global.RedisClient.Get()
	defer conn.Close()
	_, err = conn.Do("SET", "captcha:"+captchaID, puzzleX, "EX", 60)
	if err != nil {
		global.Logger.Error(err.Error())
		return "", "", "", 0, retcode.NewError(http.StatusInternalServerError, "保存CaptchaID失败")
	}

	return captchaID, "data:image/jpeg;base64," + bgBase64, "data:image/jpeg;base64," + puzzleBase64, puzzleY, nil // puzzleY 用于前端定位滑块垂直位置
}
