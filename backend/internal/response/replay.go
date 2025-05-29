package response

import "live_replay_project/backend/internal/model"

type ListReplaysVO struct {
	Total int64
	List  []model.Replay
}
