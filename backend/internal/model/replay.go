package model

import (
	"gorm.io/gorm"
	"live_replay_project/backend/common/enum"
	"time"
)

type Replay struct {
	ReplayId    uint32    `json:"replay_id" redis:"replayId"`
	Title       string    `json:"title" redis:"title"`
	Description string    `json:"description" redis:"description"`
	Duration    int64     `json:"duration" redis:"duration"`
	StoragePath string    `json:"storage_path" redis:"storagePath"`
	CoverPath   string    `json:"cover_path" redis:"coverPath"`
	UserId      uint32    `json:"user_id" redis:"userId"`
	CreateTime  time.Time `json:"create_time" redis:"createTime"`
	UpdateTime  time.Time `json:"update_time" redis:"updateTime"`
	Views       uint32    `json:"views" redis:"views"`
	Comments    uint32    `json:"comments" redis:"comments"`
}

func (r *Replay) BeforeCreate(db *gorm.DB) error {
	r.CreateTime = time.Now()
	r.UpdateTime = r.CreateTime
	r.Views = 0
	r.Comments = 0
	value := db.Statement.Context.Value(enum.CurrentId)
	if uid, ok := value.(uint32); ok {
		r.UserId = uid
	}
	return nil
}

func (r *Replay) BeforeUpdate(db *gorm.DB) error {
	r.UpdateTime = time.Now()
	return nil
}
