# API 精简索引

Base URL：`/api`

认证：登录后使用 `Authorization: Bearer <token>`。后台接口统一需要管理员权限。

## 认证与用户

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/users` | 注册 |
| `POST` | `/users/login` | 登录并返回 token |
| `GET` | `/users/:id` | 用户详情 |
| `PUT` | `/users/:id` | 更新用户 |
| `DELETE` | `/users/:id` | 删除用户 |

## 公开内容接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/posts` | 文章列表，支持分页、搜索、分类、标签、排序 |
| `GET` | `/posts/:id` | 文章详情，并记录浏览量 |
| `GET` | `/categories` | 分类列表 |
| `GET` | `/categories/:id` | 分类详情 |
| `GET` | `/categories/:id/full` | 分类详情与关联内容 |
| `GET` | `/tags` | 标签列表 |
| `GET` | `/tags/:id` | 标签详情 |
| `GET` | `/comments?post_id=<id>` | 文章评论 |
| `POST` | `/comments` | 创建评论 |
| `POST` | `/like/toggle` | 点赞/取消点赞 |
| `GET` | `/like/count` | 点赞数 |
| `GET` | `/pages` | 页面列表 |
| `GET` | `/pages/:id` | 页面详情 |
| `GET` | `/moments` | 动态列表 |
| `GET` | `/guestbook` | 已审核留言 |
| `POST` | `/guestbook` | 创建留言 |
| `GET` | `/hotdata` | 热点数据 |
| `GET` | `/stats` | 访问统计 |

## 工具接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/tools/image-compress/start` | 启动图片压缩 |
| `GET` | `/tools/image-compress/stream` | 压缩进度 SSE |
| `GET` | `/tools/image-compress/stats` | 压缩统计 |
| `GET` | `/tools/image-compress/download` | 下载压缩结果 |
| `POST` | `/drawguess/rooms` | 创建你画我猜房间 |
| `POST` | `/drawguess/rooms/:roomId/join` | 加入房间 |
| `GET` | `/drawguess/rooms/:roomId` | 房间状态 |
| `GET` | `/drawguess/rooms/:roomId/ws` | 房间 WebSocket |
| `POST` | `/drawguess/rooms/:roomId/start` | 开始游戏 |
| `POST` | `/drawguess/rooms/:roomId/guess` | 提交猜词 |
| `POST` | `/drawguess/rooms/:roomId/strokes` | 兼容 HTTP 画线提交 |
| `POST` | `/drawguess/rooms/:roomId/clear` | 清空画布 |
| `POST` | `/drawguess/rooms/:roomId/leave` | 离开房间 |

## 后台接口

后台前缀：`/api/admin`

| 模块 | 路径 |
| --- | --- |
| 用户 | `/users`、`/users/:id/status`、`/users/:id/role`、`/users/:id/password` |
| 文章 | `/posts`、`/posts/:id`、`/posts/suggest-taxonomy` |
| 分类 | `/categories`、`/categories/:id` |
| 标签 | `/tags`、`/tags/:id` |
| 评论 | `/comments`、`/comments/:id`、`/comments/:id/status`、`/comments/batch-delete`、`/comments/batch-status`、`/comments/:id/reply` |
| 动态 | `/moments`、`/moments/:id` |
| 留言 | `/guestbook`、`/guestbook/:id/status` |
| 页面 | `/pages`、`/pages/:id` |
| 上传 | `POST /upload/file`、`POST /upload/image`、`POST /upload/files`、`GET /upload/files`、`DELETE /upload/file` |
| 图片压缩 | `/upload/compress/start`、`/upload/compress/stream`、`/upload/compress/stats` |

## 注意

当前路由中仍存在若干公开写接口，如文章、分类、标签、页面和热点数据的 `POST/PUT/DELETE`。若前端公开站点不需要这些能力，应尽快加鉴权或只保留后台版本。
