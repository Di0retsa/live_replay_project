package api

import (
	"github.com/gin-gonic/gin"
	"live_replay_project/backend/global"
	"live_replay_project/backend/internal/controller"
	"live_replay_project/backend/internal/dao"
	"live_replay_project/backend/internal/service"
	"live_replay_project/backend/middle"
)

type ReplayRouter struct{}

func (rr *ReplayRouter) InitReplayRouter(group *gin.RouterGroup) {
	publicRouter := group.Group("replay")
	privateRouter := group.Group("replay")
	privateRouter.Use(middle.VerifyToken())
	replayController := controller.NewReplayController(
		service.NewReplayService(
			dao.NewReplayDao(global.DB),
		),
	)

	{
		publicRouter.GET("/list", replayController.ListReplays)
		publicRouter.GET("/static/cover/:coverName", replayController.GetCover)
	}

	{
		privateRouter.POST("/upload", replayController.UploadReplay)
		privateRouter.GET("/:replayId", replayController.GetReplayById)
		privateRouter.GET("/static/video/:videoName", replayController.GetVideo)
	}
}
