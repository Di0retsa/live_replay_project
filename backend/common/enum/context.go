package enum

import "time"

const (
	CurrentId   = "currentId"
	CurrentName = "currentName"
)

type PageNum = int

const (
	MaxPageSize = 100
	MinPageSize = 10
)

const (
	Md5vIteration = 3
)

const (
	UploadVideoPath = "./static/video/"
	UploadCoverPath = "./static/cover/"
)

const (
	PuzzleWidth  = 300
	PuzzleHeight = 300
	Tolerance    = 25 // 容差
)

const (
	// 允许写入消息到对端的时间。
	WriteWait = 10 * time.Second

	// 允许从对端读取下一个 pong 消息的时间。
	PongWait = 60 * time.Second

	// 向对端发送 ping 消息的周期。必须小于 pongWait。
	PingPeriod = (PongWait * 9) / 10

	// 允许从对端接收的最大消息尺寸。
	MaxMessageSize = 1024 // 增加到 1KB 示例
)
