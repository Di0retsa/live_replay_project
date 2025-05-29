package hub

import (
	"github.com/gorilla/websocket"
	"sync"
)

type Client struct {
	Hub      *Hub            // 指向 Hub 的指针
	Conn     *websocket.Conn // WebSocket 连接
	Send     chan []byte     // 发送消息的缓冲通道
	ReplayId uint32          // 客户端所在的房间ID
	UserId   uint32          // 认证后的用户ID
	Username string          // 认证后的用户名
}

type Hub struct {
	Clients       map[*Client]bool            // 所有已连接客户端
	Rooms         map[uint32]map[*Client]bool // 房间ID -> 该房间客户端集合
	Register      chan *Client                // 客户端注册通道
	Unregister    chan *Client                // 客户端注销通道
	Broadcast     chan HubMessage             // 从客户端接收到的待广播消息
	RoomBroadcast chan HubMessage             // 用于将消息定向到特定房间的通道
	Mu            sync.RWMutex                // 保护 rooms 和 clients 的并发访问
}

type HubMessage struct {
	ReplayId uint32
	Message  []byte
	Sender   *Client // 标记消息发送者，用于避免给自己回发消息
}

func NewHub() *Hub {
	return &Hub{
		Clients:       make(map[*Client]bool),
		Rooms:         make(map[uint32]map[*Client]bool),
		Register:      make(chan *Client),
		Unregister:    make(chan *Client),
		Broadcast:     make(chan HubMessage),
		RoomBroadcast: make(chan HubMessage, 256),
	}
}
