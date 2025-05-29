package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"live_replay_project/backend/global"
	"os/exec"
	"path/filepath"
	"strconv"
)

type FFProbeFormat struct {
	Duration string `json:"duration"`
}

type FFProbeOutput struct {
	Format FFProbeFormat `json:"format"`
}

func GetVideoDuration(filePath string) (float64, error) {
	path := filepath.Join(global.CWD, filePath)
	// ffprobe -v quiet -print_format json -show_format -show_streams filePath
	cmd := exec.Command("ffprobe",
		"-v", "quiet", // 安静模式，不输出不必要的日志
		"-print_format", "json", // 输出格式为 JSON
		"-show_format", // 显示文件格式信息 (包含时长)
		path,
	)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return 0, fmt.Errorf("ffprobe 执行失败: %w, Stderr: %s", err, stderr.String())
	}

	var ffProbeData FFProbeOutput
	if err := json.Unmarshal(out.Bytes(), &ffProbeData); err != nil {
		return 0, fmt.Errorf("解析 ffprobe JSON 输出失败: %w, Output: %s", err, out.String())
	}

	if ffProbeData.Format.Duration == "" {
		return 0, fmt.Errorf("无法从 ffprobe 输出中找到时长信息")
	}

	durationSeconds, err := strconv.ParseFloat(ffProbeData.Format.Duration, 64)
	if err != nil {
		return 0, fmt.Errorf("转换时长字符串 '%s' 为浮点数失败: %w", ffProbeData.Format.Duration, err)
	}

	return durationSeconds, nil
}
