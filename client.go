package websocket

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/iautre/auth/model"
	"github.com/iautre/gowk"
)

// 客户端信息
type Client struct {
	AppKey string
	Auid   uint
	Socket *websocket.Conn
	Send   chan []byte
	Ctx    *gin.Context
	App    *model.App
}

type ClientGroup struct {
	Clients map[uint]*Client
}

// 读取消息
func (c *Client) Read() {
	defer func() {
		Manager.Unregister <- c
		c.Socket.Close()
	}()

	for {
		c.Socket.PongHandler()
		_, message, err := c.Socket.ReadMessage()
		if err != nil {
			Manager.Unregister <- c
			c.Socket.Close()
			break
		}
		gowk.Log().Info(c.Ctx, fmt.Sprintf("读取到客户端的信息:%s", string(message)))
		c.MessageFromWS(message)
		Manager.Broadcast <- message
	}
}

//写入消息
func (c *Client) Write() {
	defer func() {
		c.Socket.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			gowk.Log().Info(c.Ctx, fmt.Sprintf("发送到到客户端的信息:%s", string(message)))
			c.Socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}

// 接收消息
func (c *Client) MessageFromWS(data []byte) {
	if messageHandler != nil {
		messageInfo := &MessageInfo{
			AppKey:  c.AppKey,
			Auid:    c.Auid,
			Message: data,
		}
		go messageHandler.Message(c.Ctx, messageInfo)
	}
}
