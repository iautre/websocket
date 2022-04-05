package websocket

import (
	"errors"

	"github.com/gin-gonic/gin"
)

type MessageInfo struct {
	AppKey  string
	Auid    string
	Message []byte
}

type ClientInfo struct {
	AppKey string
	Auid   string
}

type MessageHandler interface {
	Message(*gin.Context, *MessageInfo)
}

var messageHandler MessageHandler

func SetMessagehandler(m MessageHandler) {
	messageHandler = m
}

// 获取所有在线客户端
func GetClients() []*ClientInfo {
	arr := make([]*ClientInfo, 0)
	for _, v := range Manager.getAllClient() {
		cinfo := &ClientInfo{
			AppKey: v.AppKey,
			Auid:   v.Auid,
		}
		arr = append(arr, cinfo)
	}
	return arr
}

// 发送消息
func SendMessage(ctx *gin.Context, msg *MessageInfo) error {
	client := Manager.getClient(msg.AppKey, msg.Auid)
	if client == nil {
		// gowk.Response().Fail(ctx, gowk.NewError(500, "目标不在线"), nil)
		return errors.New("目标客户端不在线")
	}
	client.Ctx = ctx
	client.Send <- msg.Message
	// gowk.Response().Success(ctx, nil)
	return nil
}
