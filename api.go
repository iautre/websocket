package websocket

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/iautre/auth"
	"github.com/iautre/gowk"
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
	// defer func() {
	// 	if err := gowk.Recover(); err != nil {
	// 		conn.WriteJSON(gowk.Response().Message(err, nil))
	// 		conn.Close()
	// 	}
	// }
	defer func() {
		if err := recover(); err != nil {
			var errMsg *gowk.ErrorCode
			err := json.Unmarshal([]byte(string(err.(string))), &errMsg)
			if err != nil {
				errMsg = gowk.ERR_UN
			}
			conn.WriteJSON(gowk.Response().Message(errMsg, nil))
			conn.Close()
		}
	}()
	appkey := c.Query("appkey")
	token := c.Query("token")
	app := auth.CheckApp(appkey)
	//通过token解析出auid
	claims := auth.GetClaims(token)
	//校验token失败 断开连接
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
	auid, _ := strconv.Atoi(c.Param("auid"))
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
		Auid:    uint(auid),
		Message: message,
	}
	err = SendMessage(c, messageInfo)
	if err != nil {
		gowk.Response().Fail(c, gowk.NewErrorCode(500, "目标不在线"), nil)
		return
	}
	gowk.Response().Success(c, nil)
}
