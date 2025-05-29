package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"live_replay_project/backend/common/enum"
	"live_replay_project/backend/common/retcode"
	"live_replay_project/backend/global"
	"live_replay_project/backend/internal/hub"
	"live_replay_project/backend/internal/model"
	"live_replay_project/backend/internal/service"
	"log"
	"net/http"
	"strconv"
	"time"
)

type ChatController struct {
	service service.ChatService
	hub     *hub.Hub
}

func NewChatController(service service.ChatService) *ChatController {
	return &ChatController{service: service, hub: global.Hub}
}

func (cc *ChatController) GetConnection(ctx *gin.Context) {
	id := ctx.Param("replayId")
	if id == "" {
		global.Logger.Debug("ChatController GetConnection Get Param Failed")
		retcode.Fatal(ctx, retcode.NewError(http.StatusBadRequest, "房间号不能为空"), "")
		return
	}
	replayId, _ := strconv.ParseUint(id, 10, 32)

	var username string
	var userId uint64
	if value, exists := ctx.Get(enum.CurrentName); exists {
		username = value.(string)
	}
	if value, exists := ctx.Get(enum.CurrentId); exists {
		userId = value.(uint64)
	}

	conn, err := global.Upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		global.Logger.Debug("ChatController GetConnection Upgrade Failed")
		return
	}
	global.Logger.Info(fmt.Sprintf("WebSocket 连接已建立 for replay %s from %s\n", replayId, conn.RemoteAddr()))

	client := &hub.Client{
		Hub:      global.Hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		ReplayId: uint32(replayId),
		UserId:   uint32(userId),
		Username: username,
	}
	client.Hub.Register <- client

	// TODO
	go cc.service.WritePump(ctx, client)
	go cc.service.ReadPump(ctx, client)
}

func (cc *ChatController) Run(h *hub.Hub) {
	log.Println("Hub 开始运行...")
	for {
		select {
		case client := <-h.Register:
			h.Mu.Lock()
			// 初始化房间（如果不存在）
			if _, ok := h.Rooms[client.ReplayId]; !ok {
				h.Rooms[client.ReplayId] = make(map[*hub.Client]bool)
				log.Println("房间 %s 已创建", client.ReplayId)
			}
			// 将客户端添加到房间
			h.Rooms[client.ReplayId][client] = true
			h.Clients[client] = true
			log.Println("客户端 %s (用户: %s) 注册到房间 %s。房间内客户端数量: %d",
				client.Conn.RemoteAddr(), client.Username, client.ReplayId, len(h.Rooms[client.ReplayId]))
			h.Mu.Unlock()

			// 发送系统消息通知房间内其他用户有人加入
			joinMsg := model.ChatMessage{Username: "系统", Content: client.Username + " 加入了聊天室", Timestamp: time.Now().UTC(), ReplayId: client.ReplayId, Type: "system_message"}
			joinMsgBytes, _ := json.Marshal(joinMsg)
			h.RoomBroadcast <- hub.HubMessage{ReplayId: client.ReplayId, Message: joinMsgBytes, Sender: client} // sender 用于避免给自己发

			// 加载并发送历史消息给新连接的客户端
			go cc.service.LoadAndSendHistory(client)

		case client := <-h.Unregister:
			h.Mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client) // 从全局列表移除
				if roomClients, roomExists := h.Rooms[client.ReplayId]; roomExists {
					delete(roomClients, client) // 从房间列表移除
					log.Printf("客户端 %s (用户: %s) 从房间 %s 注销\n", client.Conn.RemoteAddr(), client.Username, client.ReplayId)
					if len(roomClients) == 0 {
						delete(h.Rooms, client.ReplayId) // 如果房间为空，则删除房间
						log.Printf("房间 %s 已空并移除", client.ReplayId)
					}
				}
				close(client.Send) // 关闭客户端的发送通道
			}
			h.Mu.Unlock()

		case hubMsg := <-h.RoomBroadcast: // 处理需要广播到特定房间的消息
			h.Mu.RLock()
			if roomClients, ok := h.Rooms[hubMsg.ReplayId]; ok {
				for clientInRoom := range roomClients {
					if hubMsg.Sender != nil && clientInRoom == hubMsg.Sender {
						continue
					}
					select {
					case clientInRoom.Send <- hubMsg.Message:
					default: // 发送通道已满或关闭，清理该客户端
						log.Printf("客户端 %s (用户: %s) 的发送通道已满或关闭，在广播时注销该客户端\n",
							clientInRoom.Conn.RemoteAddr(), clientInRoom.Username)
					}
				}
			}
			h.Mu.RUnlock()
		}
	}
}
