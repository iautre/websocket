package websocket

import (
	"encoding/json"
	"net/http"

	"github.com/autrec/auth"
	"github.com/autrec/gowk"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func Routers(r *gin.RouterGroup) {
	//ws连接
	r.GET("", WsHandler)
	r.POST("/:app/:auid", SendToWS)
}

//socket 连接 中间件 作用:升级协议,用户验证,自定义信息等
func WsHandler(c *gin.Context) {
	ws := &websocket.Upgrader{
		//设置允许跨域
		CheckOrigin: func(r *http.Request) bool { return true },
		//Subprotocols: []string{c.GetHeader("Sec-WebSocket-Protocol")},
	}
	conn, err := ws.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		http.NotFound(c.Writer, c.Request)
		return
	}
	appkey := c.Query("appkey")
	app, err := auth.CheckApp(appkey)
	if err != nil {
		conn.WriteJSON(gowk.Response().Message(gowk.ERR_NOAPP, nil))
		conn.Close()
		return
	}
	//通过token解析出auid
	claims, err := auth.GetClaims(c.Query("token"))
	//校验token失败 断开连接
	if err != nil {
		conn.WriteJSON(gowk.Response().Message(gowk.ERR_TOKEN, nil))
		conn.Close()
		return
	}

	client := &Client{
		AppKey: appkey,
		Auid:   claims.Auid,
		Socket: conn,
		Send:   make(chan []byte),
		Ctx:    c,
		App:    app,
	}
	go client.Read()
	go client.Write()
	Manager.Register <- client
}

//发送消息
func SendToWS(c *gin.Context) {
	appkey := c.Param("app")
	auid := c.Param("auid")
	var content map[string]interface{}
	err := c.ShouldBind(&content)
	if err != nil {
		gowk.Response().Fail(c, gowk.ERR_PARAM, err)
		return
	}
	content["fromApp"] = c.GetString("APPKEY")
	content["fromAuid"] = c.GetUint("AUID")
	message, _ := json.Marshal(content)
	messageInfo := &MessageInfo{
		AppKey:  appkey,
		Auid:    auid,
		Message: message,
	}
	err = SendMessage(c, messageInfo)
	if err != nil {
		gowk.Response().Fail(c, gowk.NewError(500, "目标不在线"), nil)
		return
	}
	gowk.Response().Success(c, nil)
}
