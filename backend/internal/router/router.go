package router

import "live_replay_project/backend/internal/router/api"

type Router struct {
	api.UserRouter
	api.ReplayRouter
	api.ChatRouter
}

var AllRouter = new(Router)
