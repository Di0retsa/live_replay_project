package api

import (
	"github.com/gin-gonic/gin"
	"live_replay_project/backend/global"
	"live_replay_project/backend/internal/controller"
	"live_replay_project/backend/internal/dao"
	"live_replay_project/backend/internal/service"
	"live_replay_project/backend/middle"
)

type UserRouter struct{}

func (ur *UserRouter) InitUserRouter(group *gin.RouterGroup) {
	publicRouter := group.Group("user")
	privateRouter := group.Group("user")
	privateRouter.Use(middle.VerifyToken())
	userController := controller.NewUserController(
		service.NewUserService(
			dao.NewUserDao(global.DB),
		),
	)

	{
		publicRouter.POST("/register", userController.Register)
		publicRouter.POST("/login/password", userController.LoginByPassword)
		publicRouter.POST("/code", userController.GetCode)
		publicRouter.POST("/login/code", userController.LoginByCode)
		publicRouter.GET("/login/captcha", userController.GetCaptcha)
		publicRouter.POST("/login/captcha", userController.VerifyCaptcha)
	}

	{
		// TODO privateRoute
	}
}
