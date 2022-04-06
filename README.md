一个基于websocket的库

## gin 使用
```
r := gin.New()
group := r.Group("/ws")
websocket.Routers(group)

```