package dao

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"live_replay_project/backend/common/retcode"
	"live_replay_project/backend/global"
	"live_replay_project/backend/internal/model"
	"net/http"
)

type ReplayDao struct {
	db *gorm.DB
}

func NewReplayDao(db *gorm.DB) *ReplayDao {
	return &ReplayDao{db: db}
}

func (r *ReplayDao) ListAllReplays(ctx *gin.Context) ([]model.Replay, error) {
	var replays []model.Replay
	err := r.db.WithContext(ctx).Find(&replays).Order("create_time desc").Error
	if err != nil {
		return nil, retcode.NewError(http.StatusInternalServerError, "List Replays Failed")
	}
	return replays, nil
}

func (r *ReplayDao) SaveReplay(ctx *gin.Context, replay model.Replay) error {
	err := r.db.WithContext(ctx).Create(&replay).Error
	if err != nil {
		return retcode.NewError(http.StatusInternalServerError, "Save Replay Failed")
	}
	return nil
}

func (r *ReplayDao) GetReplayById(ctx *gin.Context, replayId uint32) (*model.Replay, error) {
	var replay model.Replay
	err := r.db.WithContext(ctx).Find(&replay, replayId).Error

	if err != nil {
		return nil, retcode.NewError(http.StatusInternalServerError, "Get Replay Failed")
	}
	return &replay, nil
}

func (r *ReplayDao) UpdateBatch(replays *[]model.Replay) {
	for _, replay := range *replays {
		err := r.db.Model(&replay).Where("replay_id", replay.ReplayId).Updates(&replay).Error
		if err != nil {
			global.Logger.Error("Update Replays Failed")
			return
		}
	}
}
