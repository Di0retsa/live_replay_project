package common

import (
	"gorm.io/gorm"
	"live_replay_project/backend/common/enum"
)

type Result struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

type PageResult struct {
	Total   int64       `json:"total"`
	Records interface{} `json:"records"`
}

func PageVerify(page *int, pageSize *int) {
	if *page < 1 {
		*page = 1
	}
	switch {
	case *pageSize > 100:
		*pageSize = enum.MaxPageSize
	case *pageSize <= 0:
		*pageSize = enum.MinPageSize
	}
}

func (p *PageResult) Paginate(page *int, pageSize *int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		PageVerify(page, pageSize)
		db.Offset((*page - 1) * *pageSize).Limit((*pageSize))
		return db
	}
}
