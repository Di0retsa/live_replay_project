package service

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"live_replay_project/backend/common/enum"
	"live_replay_project/backend/global"
	"live_replay_project/backend/internal/dao"
	"live_replay_project/backend/internal/hub"
	"live_replay_project/backend/internal/model"
	"log"
	"time"
)

type ChatService interface {
	WritePump(ctx *gin.Context, c *hub.Client)
	ReadPump(ctx *gin.Context, c *hub.Client)
	LoadAndSendHistory(client *hub.Client)
}

type ChatServiceImpl struct {
	repo *dao.ChatDao
}

func NewChatService(repo *dao.ChatDao) ChatService {
	return &ChatServiceImpl{repo: repo}
}

func (ci *ChatServiceImpl) WritePump(ctx *gin.Context, c *hub.Client) {
	ticker := time.NewTicker(enum.PingPeriod)
	defer func() {
		ticker.Stop()
		if err := c.Conn.Close(); err != nil {
			log.Printf("关闭客户端 %s (用户: %s) 连接时发生错误: %v", c.Conn.RemoteAddr(), c.Username, err)
		}
		log.Printf("客户端 %s (用户: %s) 的 writePump 退出", c.Conn.RemoteAddr(), c.Username)
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(enum.WriteWait)); err != nil {
				log.Printf("设置写超时失败 for client %s: %v", c.Conn.RemoteAddr(), err)
				return // 返回以关闭连接
			}
			if !ok {
				// Hub 关闭了 c.send 通道，说明客户端需要关闭
				log.Printf("客户端 %s (用户: %s) 的发送通道已关闭，发送 CloseMessage", c.Conn.RemoteAddr(), c.Username)
				if err := c.Conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					log.Printf("发送 CloseMessage 失败 for client %s: %v", c.Conn.RemoteAddr(), err)
				}
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("获取 NextWriter 失败 for client %s: %v", c.Conn.RemoteAddr(), err)
				return
			}
			if _, err := w.Write(message); err != nil {
				log.Printf("写入消息失败 for client %s: %v", c.Conn.RemoteAddr(), err)
				// 不需要关闭 w，因为 NextWriter 上的错误意味着连接已损坏
				return
			}

			// 如果通道中还有更多消息，则将它们添加到当前 WebSocket 消息中以提高效率
			n := len(c.Send)
			for i := 0; i < n; i++ {
				// w.Write([]byte{'\n'}) // 如果需要分隔符
				if _, err := w.Write(<-c.Send); err != nil {
					log.Printf("写入附加消息失败 for client %s: %v", c.Conn.RemoteAddr(), err)
					return
				}
			}

			if err := w.Close(); err != nil {
				log.Printf("关闭写入器失败 for client %s: %v", c.Conn.RemoteAddr(), err)
				return
			}

		case <-ticker.C:
			// 定期发送 ping 消息以保持连接活跃或检测死连接
			if err := c.Conn.SetWriteDeadline(time.Now().Add(enum.WriteWait)); err != nil {
				log.Printf("设置 Ping 写超时失败 for client %s: %v", c.Conn.RemoteAddr(), err)
				return
			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("发送 PingMessage 失败 for client %s: %v", c.Conn.RemoteAddr(), err)
				return
			}
		}
	}
}

func (ci *ChatServiceImpl) ReadPump(_ *gin.Context, c *hub.Client) {
	defer func() {
		c.Hub.Unregister <- c // 从 Hub 注销客户端
		if err := c.Conn.Close(); err != nil {
			log.Printf("关闭客户端 %s (用户: %s) 连接时发生错误: %v", c.Conn.RemoteAddr(), c.Username, err)
		}
		log.Printf("客户端 %s (用户: %s) 的 readPump 退出", c.Conn.RemoteAddr(), c.Username)
	}()

	c.Conn.SetReadLimit(enum.MaxMessageSize)
	if err := c.Conn.SetReadDeadline(time.Now().Add(enum.PongWait)); err != nil {
		log.Printf("设置初始读超时失败 for client %s: %v", c.Conn.RemoteAddr(), err)
		return // 提前退出，将触发 defer 中的清理
	}
	c.Conn.SetPongHandler(func(string) error {
		log.Printf("收到来自客户端 %s (用户: %s) 的 Pong 消息", c.Conn.RemoteAddr(), c.Username)
		return c.Conn.SetReadDeadline(time.Now().Add(enum.PongWait))
	})

	for {
		messageType, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
				log.Printf("客户端 %s (用户: %s) 连接非预期关闭: %v", c.Conn.RemoteAddr(), c.Username, err)
			} else if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("客户端 %s (用户: %s) 正常关闭连接", c.Conn.RemoteAddr(), c.Username)
			} else {
				log.Printf("读取消息错误 from client %s (用户: %s): %v", c.Conn.RemoteAddr(), c.Username, err)
			}
			break // 退出循环，将触发 defer 中的注销和关闭
		}

		if messageType != websocket.TextMessage {
			log.Printf("收到非文本消息类型 %d from client %s", messageType, c.Conn.RemoteAddr())
			continue // 忽略非文本消息
		}

		log.Printf("收到来自客户端 %s (用户: %s, 房间: %s) 的消息: %s",
			c.Conn.RemoteAddr(), c.Username, c.ReplayId, string(messageBytes))

		var receivedMsg model.ChatMessage
		if err := json.Unmarshal(messageBytes, &receivedMsg); err != nil {
			log.Printf("解析客户端 %s (用户: %s) 消息 JSON 错误: %v, 原始消息: %s",
				c.Conn.RemoteAddr(), c.Username, err, string(messageBytes))
			// 发送错误消息回客户端
			errMsg := model.ChatMessage{Content: "消息格式错误", Type: "error", Timestamp: time.Now().UTC()}
			errBytes, _ := json.Marshal(errMsg)
			c.Send <- errBytes
			continue
		}

		// 填充/覆盖服务器端信息
		receivedMsg.Timestamp = time.Now().UTC()
		receivedMsg.ReplayId = c.ReplayId // 确保使用服务器端的房间ID
		receivedMsg.UserId = c.UserId     // 使用服务器端认证的UserID
		receivedMsg.Username = c.Username // 使用服务器端认证的Username

		processedMsgBytes, err := json.Marshal(receivedMsg)
		if err != nil {
			log.Printf("序列化处理后消息错误 for client %s (用户: %s): %v",
				c.Conn.RemoteAddr(), c.Username, err)
			continue
		}

		// (可选) 持久化消息到数据库
		if receivedMsg.Type != "join" {
			go ci.repo.SaveChatMessage(&receivedMsg)
			go increaseComments(c.ReplayId)
		}

		// 将处理后的消息发送到 Hub 的 roomBroadcast 通道，由 Hub 负责分发
		c.Hub.RoomBroadcast <- hub.HubMessage{
			ReplayId: c.ReplayId,
			Message:  processedMsgBytes,
			Sender:   c, // 传递发送者信息
		}
	}
}

// 加载并发送历史消息
func (ci *ChatServiceImpl) LoadAndSendHistory(client *hub.Client) {
	log.Printf("为客户端 %s (用户: %s, _房间: %s) 加载历史消息...", client.Conn.RemoteAddr(), client.Username, client.ReplayId)
	// 假设 db.GetChatHistory(replayID, limit) 返回 []ChatMessage
	history, err := ci.repo.GetChatHistory(client.ReplayId, 20) // 获取最近20条
	if err != nil {
		log.Printf("加载房间 %s 历史消息失败: %v", client.ReplayId, err)
		return
	}

	for i := 0; i < len(history); i++ { // 通常按时间倒序获取，发送时再反转或直接发送
		msg := history[i]
		msg.Type = "history_message" // 标记为历史消息
		historyMsgBytes, err := json.Marshal(msg)
		if err != nil {
			log.Printf("序列化历史消息失败: %v", err)
			continue
		}
		// 直接发送给该客户端，不通过 Hub 广播
		select {
		case client.Send <- historyMsgBytes:
		default:
			log.Printf("发送历史消息到客户端 %s (用户: %s) 失败 (通道已满或关闭)", client.Conn.RemoteAddr(), client.Username)
			return // 如果通道关闭，则停止发送
		}
	}
	log.Printf("为客户端 %s (用户: %s, 房间: %s) 发送完 %d 条历史消息", client.Conn.RemoteAddr(), client.Username, client.ReplayId, len(history))
}

func increaseComments(replayId uint32) error {
	conn := global.RedisClient.Get()
	defer conn.Close()
	key := fmt.Sprintf("replays:%d", replayId)
	_, err := conn.Do("JSON.NUMINCRBY", key, ".comments", 1)
	return err
}
