package initialize

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/websocket"
	"live_replay_project/backend/common/utils"
	"live_replay_project/backend/config"
	"live_replay_project/backend/global"
	"live_replay_project/backend/internal/hub"
	r "live_replay_project/backend/internal/router"
	"live_replay_project/backend/logger"
	"live_replay_project/backend/persistence"
	"net/http"
	"os"
	"time"
)

func GlobalInit() *gin.Engine {
	// 配置文件初始化
	global.Config = config.InitLoadConfig()
	println(global.Config)
	// Log
	global.Logger = logger.NewMySLog(global.Config.Log.Level, global.Config.Log.FilePath)
	// Gorm
	global.DB = persistence.InitDatabase(global.Config.DataSource.Dsn())
	// CWD
	global.CWD, _ = os.Getwd()
	// Redis Pool
	global.RedisClient = redisPoolInit()
	// Captcha
	utils.LoadImagePaths()
	// webSocket upgrader
	global.Upgrader = upgraderInit()
	// Hub
	global.Hub = hub.NewHub()
	//router
	router := routerInit()
	return router
}

func redisPoolInit() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     global.Config.Redis.MaxIdle,
		IdleTimeout: time.Second * time.Duration(global.Config.Redis.IdleTimeout),
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(
				"tcp",
				global.Config.Redis.Host+":"+global.Config.Redis.Port,
				redis.DialPassword(global.Config.Redis.Password),
				// redis.DialUsername(global.Config.Redis.Username),
			)
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
		MaxActive:       global.Config.Redis.MaxActive,
		Wait:            global.Config.Redis.Wait,
		MaxConnLifetime: time.Second * time.Duration(global.Config.Redis.MaxConnLifeTime),
	}
}

func routerInit() *gin.Engine {
	router := gin.Default()
	allRouter := r.AllRouter

	// 跨域
	corsConfig := cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // 允许的前端源
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "user"},
		ExposeHeaders:    []string{"Content-Length", "Content-Range"}, // 允许前端访问 Content-Range
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(corsConfig))

	// user
	user := router.Group("")
	{
		allRouter.UserRouter.InitUserRouter(user)
	}
	// replay
	replay := router.Group("")
	{
		allRouter.ReplayRouter.InitReplayRouter(replay)
	}
	// chat
	chat := router.Group("")
	{
		allRouter.ChatRouter.InitChatRouter(chat)
	}

	return router
}

func upgraderInit() *websocket.Upgrader {
	return &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			//allowOrigins := []string{"http://localhost:3000", "http://localhost:5173"}
			//origin := r.Header.Get("Origin")
			//for _, allowOrigin := range allowOrigins {
			//	if origin == allowOrigin {
			//		return true
			//	}
			//}
			//return false
			return true
		},
	}
}
