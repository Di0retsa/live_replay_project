package main

import (
	"github.com/gin-gonic/gin"
	"live_replay_project/backend/global"
	"live_replay_project/backend/initialize"
)

func main() {
	router := initialize.GlobalInit()
	gin.SetMode(global.Config.Server.Level)
	router.Run(":" + global.Config.Server.Port)
}
