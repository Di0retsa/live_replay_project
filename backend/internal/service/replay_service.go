package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"live_replay_project/backend/common/enum"
	"live_replay_project/backend/common/retcode"
	"live_replay_project/backend/common/utils"
	"live_replay_project/backend/global"
	"live_replay_project/backend/internal/dao"
	"live_replay_project/backend/internal/model"
	"live_replay_project/backend/internal/request"
	"live_replay_project/backend/internal/response"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type ReplayService interface {
	ListReplays(ctx *gin.Context) (response.ListReplaysVO, error)
	SaveReplay(ctx *gin.Context, dto request.UploadReplayDTO, video *multipart.FileHeader, cover *multipart.FileHeader) error
	GetReplayById(ctx *gin.Context, replayID uint32) (*model.Replay, error)
	GetReplayFile(ctx *gin.Context, fileType string, fileName string) (string, error)
}

type ReplayServiceImpl struct {
	repo *dao.ReplayDao
}

func NewReplayService(repo *dao.ReplayDao) ReplayService {
	r := &ReplayServiceImpl{repo: repo}
	go r.syncToDBEveryMinute()
	return r
}

func (r *ReplayServiceImpl) ListReplays(ctx *gin.Context) (response.ListReplaysVO, error) {
	listReplaysVO := response.ListReplaysVO{}

	// 尝试从redis获取
	conn := global.RedisClient.Get()
	defer conn.Close()

	//replaysJson, err := redis.ByteSlices(conn.Do("HVALS", "replays"))
	//if !errors.Is(err, redis.ErrNil) && len(replaysJson) != 0 {
	//	for _, replayJson := range replaysJson {
	//		replay := model.Replay{}
	//		json.Unmarshal(replayJson, &replay)
	//		listReplaysVO.List = append(listReplaysVO.List, replay)
	//	}
	//	listReplaysVO.Total = int64(len(listReplaysVO.List))
	//	return listReplaysVO, nil
	//}

	// 视频多了或许应该考虑分页查询
	ids, err := redis.Strings(conn.Do("SMEMBERS", "replays:ids"))
	if !errors.Is(err, redis.ErrNil) && len(ids) != 0 {
		//for _, id := range ids {
		//	replayJson, _ := redis.Bytes(conn.Do("JSON.GET", id))
		//	var replay model.Replay
		//	_ = json.Unmarshal(replayJson, &replay)
		//	listReplaysVO.List = append(listReplaysVO.List, replay)
		//}

		// Pipe优化
		for _, id := range ids {
			conn.Send("JSON.GET", id)
		}
		conn.Flush()
		for range ids {
			replayJson, _ := redis.Bytes(conn.Receive())
			var replay model.Replay
			_ = json.Unmarshal(replayJson, &replay)
			listReplaysVO.List = append(listReplaysVO.List, replay)
		}
		listReplaysVO.Total = int64(len(listReplaysVO.List))
		return listReplaysVO, nil
	}

	replays, err := r.repo.ListAllReplays(ctx)
	if err != nil {
		return listReplaysVO, err
	}
	listReplaysVO.List = replays
	listReplaysVO.Total = int64(len(replays))

	//for _, replay := range replays {
	//	key := replay.ReplayId
	//	replayJson, _ := json.Marshal(replay)
	//	conn.Send("HMSET", "replays", key, replayJson)
	//}

	// 异步保存
	go saveReplaysToRedis(&replays)
	//for _, replay := range replays {
	//	conn.Send("MULTI")
	//	key := fmt.Sprintf("replays:%d", replay.ReplayId)
	//	replayJson, _ := json.Marshal(replay)
	//	_, err := conn.Do("JSON.SET", key, ".", replayJson)
	//	if err != nil {
	//		fmt.Println("redis set err:", err)
	//	}
	//	_, err = conn.Do("SADD", "replays:ids", key)
	//	if err != nil {
	//		fmt.Println("redis sadd err:", err)
	//	}
	//	_, err = conn.Do("EXEC")
	//	if err != nil {
	//		return listReplaysVO, retcode.NewError(http.StatusInternalServerError, "Redis EXEC Failed")
	//	}
	//}
	return listReplaysVO, nil
}

func (r *ReplayServiceImpl) SaveReplay(ctx *gin.Context, dto request.UploadReplayDTO, video *multipart.FileHeader, cover *multipart.FileHeader) error {
	videoName := video.Filename
	videoFilePath := filepath.Join(enum.UploadVideoPath, videoName)
	err := ctx.SaveUploadedFile(video, videoFilePath)
	if err != nil {
		return retcode.NewError(http.StatusInternalServerError, "上传文件出错")
	}

	coverName := cover.Filename
	coverFilePath := filepath.Join(enum.UploadCoverPath, coverName)
	err = ctx.SaveUploadedFile(cover, coverFilePath)
	if err != nil {
		return retcode.NewError(http.StatusInternalServerError, "上传文件出错")
	}

	duration, err := utils.GetVideoDuration(videoFilePath)
	if err != nil {
		return err
	}

	replay := model.Replay{
		Title:       dto.Title,
		Description: dto.Description,
		Duration:    int64(duration),
		StoragePath: videoFilePath,
		CoverPath:   coverFilePath,
	}

	id, exists := ctx.Get(enum.CurrentId)
	if !exists {
		return retcode.NewError(http.StatusInternalServerError, "获取用户id失败")
	}
	if uid, ok := id.(uint64); ok {
		replay.UserId = uint32(uid)
	}

	err = r.repo.SaveReplay(ctx, replay)

	// 异步删除缓存
	go r.removeReplaysFromRedis()
	return err
}

func (r *ReplayServiceImpl) GetReplayById(ctx *gin.Context, replayID uint32) (*model.Replay, error) {
	replay := &model.Replay{}
	conn := global.RedisClient.Get()
	defer conn.Close()

	// replayJson, err := redis.Bytes(conn.Do("HGET", "replays", replayID))
	key := fmt.Sprintf("replays:%d", replayID)
	replayJson, err := redis.Bytes(conn.Do("JSON.GET", key))
	if !errors.Is(err, redis.ErrNil) && len(replayJson) != 0 {
		err := json.Unmarshal(replayJson, replay)
		if err == nil {
			go increaseViews(replayID)
			return replay, nil
		}
	}

	replay, err = r.repo.GetReplayById(ctx, replayID)
	if err != nil {
		return nil, err
	}
	// conn.Do("HSET", "replays", replayID, replay)
	replayJson, err = json.Marshal(replay)
	_, err = conn.Do("JSON.SET", key, ".", replayJson)
	if err != nil {
		fmt.Println(err)
	}

	go increaseViews(replayID)
	return replay, nil
}

func (r *ReplayServiceImpl) GetReplayFile(_ *gin.Context, fileType string, fileName string) (string, error) {
	filePath := filepath.Join(global.CWD, "static", fileType, fileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		global.Logger.Debug("文件未找到")
		return "", retcode.NewError(http.StatusNotFound, "文件未找到")
	}
	return filePath, nil
}

func saveReplaysToRedis(replays *[]model.Replay) error {
	conn := global.RedisClient.Get()
	conn.Send("MULTI")
	for _, replay := range *replays {
		key := fmt.Sprintf("replays:%d", replay.ReplayId)
		replayJson, _ := json.Marshal(replay)
		_, _ = conn.Do("JSON.SET", key, ".", replayJson)
		_, _ = conn.Do("SADD", "replays:ids", key)
	}
	_, err := conn.Do("EXEC")
	return err
}

func (r *ReplayServiceImpl) removeReplaysFromRedis() error {
	conn := global.RedisClient.Get()
	defer conn.Close()
	// 真正删除前把views和comments更新回数据库
	ids, _ := redis.Values(conn.Do("SMEMBERS", "replays:ids"))
	var replays []model.Replay
	// Pipe优化
	for _, id := range ids {
		conn.Send("JSON.GET", id)
	}
	conn.Flush()
	for range ids {
		replayJson, _ := redis.Bytes(conn.Receive())
		var replay model.Replay
		_ = json.Unmarshal(replayJson, &replay)
		replays = append(replays, replay)
	}
	r.repo.UpdateBatch(&replays)

	// 真正开始删除
	conn.Send("MULTI")
	for _, id := range ids {
		conn.Send("JSON.DEL", id)
		conn.Send("SREM", "replays:ids", id)
	}
	conn.Send("EXEC")
	return conn.Flush()
}

func increaseViews(replayId uint32) error {
	conn := global.RedisClient.Get()
	defer conn.Close()
	key := fmt.Sprintf("replays:%d", replayId)
	_, err := conn.Do("JSON.NUMINCRBY", key, ".views", 1)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

func (r *ReplayServiceImpl) syncToDBEveryMinute() {
	ticker := time.NewTicker(60 * time.Second)
	for {
		select {
		case <-ticker.C:
			global.Logger.Info("定时任务触发：将缓存同步至DB")
			r.syncToDB()
		}
	}
}

func (r *ReplayServiceImpl) syncToDB() {
	conn := global.RedisClient.Get()
	defer conn.Close()
	ids, err := redis.Strings(conn.Do("SMEMBERS", "replays:ids"))
	if !errors.Is(err, redis.ErrNil) && len(ids) != 0 {
		var replays []model.Replay
		// Pipe优化
		for _, id := range ids {
			conn.Send("JSON.GET", id)
		}
		conn.Flush()
		for range ids {
			replayJson, _ := redis.Bytes(conn.Receive())
			var replay model.Replay
			_ = json.Unmarshal(replayJson, &replay)
			replays = append(replays, replay)
		}
		r.repo.UpdateBatch(&replays)
	}
	return
}
