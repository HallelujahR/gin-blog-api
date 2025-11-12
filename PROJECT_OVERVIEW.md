## 项目简介

一个前后端分离的博客系统：后端基于 Golang（Gin + GORM + MySQL），前端为 Vue 构建并由 Nginx 提供静态资源与反向代理。

### 技术栈
- 后端：Go 1.25、Gin、GORM、MySQL 8.0.44、bcrypt（用户密码）
- 前端：Vue（构建产物 dist），Nginx 提供静态资源与 /api 反代
- 容器/编排：Docker、Docker Compose

### 目录要点（后端）
- `controllers/`：HTTP 控制器（含后台与公开接口）
- `service/`：业务逻辑（用户、文章、分类、标签、评论、上传等）
- `dao/`：数据访问（GORM）
- `models/`：数据模型（User、Post、Tag、Category、Comment 等）
- `middleware/`：中间件（鉴权、CORS）
- `initializer/`：启动初始化（配置、数据库、路由）
- `database/db.go`：数据库连接（读取环境变量构造 DSN）
- `configs/config.go`：基础配置（API_BASE_URL）
- `routes/`：路由定义（公开与后台）
- `uploads/`：上传文件目录挂载

### 运行与部署（生产）
- Compose 服务：
  - `mysql`（3306:3306）
  - `api`（8080:8080）
  - `frontend`（80/443）挂载 `/opt/blog/gin-blog-vue-font/dist`
- Nginx 反向代理：
  - 静态：`/` -> `/usr/share/nginx/html`
  - API：`/api/` -> `api:8080`
- 脚本：`scripts/deploy.sh` 一键部署，包含前端构建（在 `/opt/blog/gin-blog-vue-font` 使用 Node 镜像构建）与默认 Nginx 配置生成。

### 配置与环境
- `.env`（与 compose 同目录）：`BLOG_ENV=prod`、`DB_HOST=mysql`、`DB_*`、`API_BASE_URL`、`MYSQL_ROOT_PASSWORD`
- 数据库账号建议使用非 root 的业务账号，并授予 `blog` 库权限

### 开发建议
- 后端本地使用 `air` 开发热重载；前端使用 `npm run dev` + Vite 代理 `/api` 到 `http://localhost:8080`

### 安全与运维
- 生产禁用默认弱口令，限制 3306 外网访问或使用安全组控制
- 定期备份数据库与上传目录
- 镜像与依赖定期更新

### 访问入口
- 前端：`http://<服务器IP>`
- API：`http://<服务器IP>:8080`（生产前端通过 `/api` 反代访问）




