package persistence

import (
	"errors"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"live_replay_project/backend/global"
	"time"
)

var (
	GormToManyRequestError = errors.New("gorm: too many request")
)

func InitDatabase(dsn string) *gorm.DB {
	var ormLogger logger.Interface
	if gin.Mode() == gin.DebugMode {
		ormLogger = logger.Default.LogMode(logger.Info)
	} else {
		ormLogger = logger.Default
	}
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,   // DSN data source name
		DefaultStringSize:         256,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false, // 根据版本自动配置
	}), &gorm.Config{
		Logger: ormLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic(err)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Second * 30)

	SlowQueryLog(db)

	GormRateLimiter(db, rate.NewLimiter(500, 1000))

	return db
}

func SlowQueryLog(db *gorm.DB) {
	err := db.Callback().Query().Before("*").Register("slow_query_start", func(db *gorm.DB) {
		now := time.Now()
		db.Set("start_time", now)
	})
	if err != nil {
		panic(err)
	}

	err = db.Callback().Query().After("*").Register("slow_query_end", func(db *gorm.DB) {
		now := time.Now()
		start, ok := db.Get("slow_query_start")
		if ok {
			duration := now.Sub(start.(time.Time))
			if duration > time.Millisecond*200 {
				global.Logger.Error("慢查询", "SQL:", db.Statement.SQL.String())
			}
		}
	})
	if err != nil {
		panic(err)
	}
}

func GormRateLimiter(db *gorm.DB, r *rate.Limiter) {
	err := db.Callback().Query().Before("*").Register("RateLimitGormMiddleware", func(db *gorm.DB) {
		if !r.Allow() {
			db.AddError(GormToManyRequestError)
			global.Logger.Error(GormToManyRequestError.Error())
			return
		}
	})
	if err != nil {
		panic(err)
	}
}
