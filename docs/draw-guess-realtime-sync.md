# 你画我猜实时同步

## 目标

多人联机作画时允许轻微延迟，但远端必须看到完整、连续、不断生长的一笔。

## 当前方案

采用“单房间 WebSocket + 整笔快照同步”：

1. 前端在 `pointerdown` 生成稳定 `stroke_id`。
2. 前端维护当前整笔的完整点集，每次发送这笔的最新快照。
3. 后端按 `player_id + stroke_id` 覆盖同一笔旧快照。
4. 远端收到更新后替换本地同一笔并重绘画布。

这个方案比纯增量点更耗带宽，但对最多 5 人的小房间更稳定，也更容易保证画线连续。

## 后端入口

- `internal/modules/content/routes/drawguess.go`：注册 `/api/drawguess` 路由。
- `internal/modules/drawguess/controller.go`：HTTP 与 WebSocket 控制器。
- `internal/modules/drawguess/manager.go`：房间、玩家、聊天、轮次、画布状态。

关键数据结构：

```go
type CanvasAction struct {
    ID       string  `json:"id"`
    StrokeID string  `json:"stroke_id,omitempty"`
    PlayerID string  `json:"player_id"`
    Kind     string  `json:"kind"`
    Color    string  `json:"color,omitempty"`
    Width    int     `json:"width,omitempty"`
    Tool     string  `json:"tool,omitempty"`
    Final    bool    `json:"final,omitempty"`
    Points   []Point `json:"points,omitempty"`
}
```

## 运行限制

- 最大房间数：`60`
- 每房间最大玩家数：`5`
- 每房间最大画布动作：`320`
- 单笔最大点数：`4096`
- 回合时长：`80s`
- 房间空闲清理：`20min`

## 本地与线上代理

本地 Vite 代理需要启用 WebSocket：

```js
proxy: {
  '/api': {
    target: 'http://localhost:8080',
    changeOrigin: true,
    ws: true,
  },
}
```

Nginx 反代需要保留 Upgrade：

```nginx
location /api/ {
    proxy_pass http://127.0.0.1:8080;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_read_timeout 600s;
}
```

## 后续优化

- 对远端重绘做 80-120ms 节流。
- 房间量增长后再考虑二进制点协议。
- 开发模式增加 `stroke_id`、点数、发送频率等调试信息。
