package api

import (
	"github.com/gin-gonic/gin"
	"live_replay_project/backend/global"
	"live_replay_project/backend/internal/controller"
	"live_replay_project/backend/internal/dao"
	"live_replay_project/backend/internal/service"
	"live_replay_project/backend/middle"
)

type ChatRouter struct{}

func (c *ChatRouter) InitChatRouter(group *gin.RouterGroup) {
	privateRouter := group.Group("ws/chat")
	privateRouter.Use(middle.VerifyToken())
	chatController := controller.NewChatController(
		service.NewChatService(
			dao.NewChatDao(global.DB),
		),
	)
	go chatController.Run(global.Hub)

	{
		privateRouter.GET("/:replayId", chatController.GetConnection)
	}
}
