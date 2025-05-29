package controller

import (
	"github.com/gin-gonic/gin"
	"live_replay_project/backend/common/retcode"
	"live_replay_project/backend/global"
	"live_replay_project/backend/internal/request"
	"live_replay_project/backend/internal/service"
	"net/http"
	"strconv"
	"strings"
)

type ReplayController struct {
	service service.ReplayService
}

func NewReplayController(service service.ReplayService) *ReplayController {
	return &ReplayController{service: service}
}

func (rc *ReplayController) ListReplays(ctx *gin.Context) {
	replays, err := rc.service.ListReplays(ctx)
	if err != nil {
		global.Logger.Debug("ReplayController ListReplays Failed:", err)
		retcode.Fatal(ctx, err, "")
		return
	}
	retcode.OK(ctx, replays)
}

// UploadReplay 暂时只是将文件路径等放到本地，待接入OSS
func (rc *ReplayController) UploadReplay(ctx *gin.Context) {
	video, err := ctx.FormFile("uploadVideo")
	if err != nil || !strings.HasSuffix(video.Filename, ".mp4") {
		global.Logger.Debug("ReplayController Upload Video Failed:", err)
		retcode.Fatal(ctx, err, "请上传后缀为mp4的视频！")
		return
	}

	cover, err := ctx.FormFile("uploadCover")
	if err != nil || !(strings.HasSuffix(cover.Filename, ".jpg") || strings.HasSuffix(cover.Filename, ".png")) {
		global.Logger.Debug("ReplayController Upload Cover Failed:", err)
		retcode.Fatal(ctx, err, "请上传后缀为jpg或png的图片！")
		return
	}
	dto := request.UploadReplayDTO{}

	err = ctx.Bind(&dto)
	if len([]rune(dto.Description)) > 200 {
		retcode.Fatal(ctx, err, "简介长度不能大于200！")
		return
	}

	if err != nil {
		global.Logger.Debug("ReplayController Upload Replay Binding Failed:", err)
		retcode.Fatal(ctx, err, "")
		return
	}

	err = rc.service.SaveReplay(ctx, dto, video, cover)
	if err != nil {
		global.Logger.Debug("ReplayController Save Replay Failed:", err)
		retcode.Fatal(ctx, err, "")
		return
	}
	retcode.OK(ctx, "")
}

func (rc *ReplayController) GetReplayById(ctx *gin.Context) {
	replayId, err := strconv.ParseUint(ctx.Param("replayId"), 10, 32)
	if err != nil {
		global.Logger.Debug("ReplayController GetReplayById Get Path Param Failed:", err)
		retcode.Fatal(ctx, err, "路径参数获取失败")
		return
	}

	replay, err := rc.service.GetReplayById(ctx, uint32(replayId))
	if err != nil {
		global.Logger.Debug("ReplayController GetReplayById Failed:", err)
		retcode.Fatal(ctx, err, "")
		return
	}
	retcode.OK(ctx, replay)
}

func (rc *ReplayController) GetCover(ctx *gin.Context) {
	coverName := ctx.Param("coverName")
	if coverName == "" {
		global.Logger.Debug("ReplayController Get Cover Name Failed")
		retcode.Fatal(ctx, retcode.NewError(http.StatusBadRequest, ""), "获取路径参数失败")
		return
	}

	path, err := rc.service.GetReplayFile(ctx, "cover", coverName)
	if err != nil || path == "" {
		retcode.Fatal(ctx, err, "")
	}
	ctx.File(path)
}

func (rc *ReplayController) GetVideo(ctx *gin.Context) {
	videoName := ctx.Param("videoName")
	if videoName == "" {
		global.Logger.Debug("ReplayController Get Video Name Failed")
		retcode.Fatal(ctx, retcode.NewError(http.StatusBadRequest, ""), "获取路径参数失败")
		return
	}

	file, err := rc.service.GetReplayFile(ctx, "video", videoName)
	if err != nil || file == "" {
		retcode.Fatal(ctx, err, "")
	}
	ctx.File(file)
}
