package global

import (
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
	"live_replay_project/backend/config"
	"live_replay_project/backend/internal/hub"
	"live_replay_project/backend/logger"
)

var (
	Config      *config.AllConfig
	Logger      *logger.MySLog
	DB          *gorm.DB
	CWD         string
	RedisClient *redis.Pool
	Upgrader    *websocket.Upgrader
	Hub         *hub.Hub
)
