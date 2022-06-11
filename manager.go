package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/iautre/gowk"
)

func init() {
	go Manager.Start()
}

// 客户端管理
type ClientManager struct {
	ClientGroup          map[string]*ClientGroup
	Broadcast            chan []byte
	Register, Unregister chan *Client
	Lock                 sync.Mutex
}

// 全局 在线客户端管理
var Manager = &ClientManager{
	Broadcast:   make(chan []byte),
	Register:    make(chan *Client),
	Unregister:  make(chan *Client),
	ClientGroup: make(map[string]*ClientGroup),
}

// websocker运行, 用协程开启start -> go Manager.Start()
func (manager *ClientManager) Start() {
	for {
		log.Println("<---管道通信--->")
		select {
		case conn := <-manager.Register:
			manager.register(conn)
		case conn := <-manager.Unregister:
			manager.unregister(conn)
		case message := <-manager.Broadcast:
			manager.broadcast(message)
		}
	}
}

//注册
func (manager *ClientManager) register(client *Client) {
	log.Printf("新用户加入, appkey: %v, auid: %v", client.AppKey, client.Auid)
	manager.addClient(client)
	jsonMessage, _ := json.Marshal(gowk.Response().Message(gowk.ERR_WS_CONTENT, gowk.M{"appName": client.App.Name}))
	client.Send <- jsonMessage
}

//离开
func (manager *ClientManager) unregister(client *Client) {
	log.Printf("用户离开, appkey: %v, auid: %v", client.AppKey, client.Auid)
	if old := manager.getClient(client.AppKey, client.Auid); old != nil {
		jsonMessage, _ := json.Marshal(gowk.Response().Message(gowk.ERR_WS_CLOSE, nil))
		old.Send <- jsonMessage
		close(old.Send)
		manager.removeClient(client.AppKey, client.Auid)
	}
}

//处理消息
func (manager *ClientManager) broadcast(message []byte) {
	clientStruct := Client{}
	json.Unmarshal(message, &clientStruct)
	if client := manager.getClient(clientStruct.AppKey, clientStruct.Auid); client != nil {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			manager.removeClient(clientStruct.AppKey, clientStruct.Auid)
		}
	}
}

// 移除客户端
func (manager *ClientManager) removeClient(appkey string, auid uint) {
	manager.Lock.Lock()
	defer manager.Lock.Unlock()
	if group, ok := manager.ClientGroup[appkey]; ok {
		if _, ok := group.Clients[auid]; ok {
			delete(group.Clients, auid)
		}
	}
}

// 新增客户端
func (manager *ClientManager) addClient(client *Client) {
	manager.Lock.Lock()
	defer manager.Lock.Unlock()
	_, ok := manager.ClientGroup[client.AppKey]
	if ok == false {
		manager.ClientGroup[client.AppKey] = &ClientGroup{Clients: make(map[uint]*Client)}
	}
	manager.ClientGroup[client.AppKey].Clients[client.Auid] = client
}

// 获取客户端
func (manager *ClientManager) getClient(appkey string, auid uint) *Client {
	manager.Lock.Lock()
	defer manager.Lock.Unlock()
	if group, ok := manager.ClientGroup[appkey]; ok {
		if client, ok := group.Clients[auid]; ok {
			return client
		}
	}
	return nil
}

// 获取所有客户端
func (manager *ClientManager) getAllClient() []*Client {
	manager.Lock.Lock()
	defer manager.Lock.Unlock()
	arr := make([]*Client, 0)
	for _, v1 := range manager.ClientGroup {
		for _, v := range v1.Clients {
			arr = append(arr, v)
		}
	}
	return arr
}
