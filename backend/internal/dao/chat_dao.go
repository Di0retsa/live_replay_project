package dao

import (
	"gorm.io/gorm"
	"live_replay_project/backend/common/retcode"
	"live_replay_project/backend/internal/model"
	"net/http"
)

type ChatDao struct {
	db *gorm.DB
}

func NewChatDao(db *gorm.DB) *ChatDao {
	return &ChatDao{db: db}
}

func (c *ChatDao) SaveChatMessage(message *model.ChatMessage) error {
	err := c.db.Create(message).Error
	if err != nil {
		return retcode.NewError(http.StatusInternalServerError, "保存消息失败")
	}
	return nil
}

func (c *ChatDao) GetChatHistory(replayId uint32, limit int) ([]model.ChatMessage, error) {
	var chatMessages []model.ChatMessage
	err := c.db.Where("replay_id = ?", replayId).Find(&chatMessages).Order("timestamp desc").Limit(limit).Error
	if err != nil {
		return nil, retcode.NewError(http.StatusInternalServerError, "获取历史消息失败")
	}
	return chatMessages, nil
}
