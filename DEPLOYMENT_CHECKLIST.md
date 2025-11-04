# 部署检查清单

## 部署前准备

### 服务器环境
- [ ] 服务器已安装Docker和Docker Compose
- [ ] 服务器已开放必要端口（80, 443, 8080, 22）
- [ ] 服务器有足够的资源（CPU、内存、磁盘）

### 代码准备
- [ ] 代码已推送到GitHub仓库
- [ ] 已创建必要的Docker文件（Dockerfile, docker-compose.yml）
- [ ] 已创建部署脚本（scripts/deploy.sh）
- [ ] 已创建环境变量模板（.env.example）

### 配置准备
- [ ] 已配置数据库连接信息
- [ ] 已配置API基础URL
- [ ] 已配置前端API地址
- [ ] 已配置域名和DNS解析

## 部署步骤

### 第一次部署
1. [ ] 在服务器上克隆代码
2. [ ] 复制.env.example为.env并配置
3. [ ] 修改database/db.go支持环境变量
4. [ ] 修改前端API地址配置
5. [ ] 运行部署脚本
6. [ ] 验证服务是否正常启动
7. [ ] 测试API接口
8. [ ] 测试前端页面

### 自动化部署
1. [ ] 配置GitHub Secrets（SERVER_HOST, SERVER_USER, SERVER_SSH_KEY）
2. [ ] 在服务器上配置SSH密钥
3. [ ] 测试GitHub Actions workflow
4. [ ] 验证自动部署功能

## 部署后验证

### 功能验证
- [ ] 前端页面可以正常访问
- [ ] API接口可以正常调用
- [ ] 数据库连接正常
- [ ] 文件上传功能正常
- [ ] 图片可以正常显示
- [ ] 用户登录功能正常
- [ ] 后台管理功能正常

### 性能验证
- [ ] 页面加载速度正常
- [ ] API响应时间正常
- [ ] 数据库查询性能正常

### 安全验证
- [ ] 已配置HTTPS（如果有域名）
- [ ] 已修改所有默认密码
- [ ] 防火墙已正确配置
- [ ] 敏感信息已从代码中移除

## 监控和维护

### 日志监控
- [ ] 已配置日志查看方式
- [ ] 已设置日志轮转（可选）

### 备份策略
- [ ] 已配置数据库备份脚本
- [ ] 已测试备份恢复流程
- [ ] 已设置定期备份计划（可选）

### 更新流程
- [ ] 已了解如何更新代码
- [ ] 已了解如何重启服务
- [ ] 已了解如何回滚（如有需要）

## 常见问题

### 服务无法启动
- [ ] 检查Docker是否运行：`sudo systemctl status docker`
- [ ] 检查端口是否被占用：`netstat -tulpn | grep -E '8080|80|3306'`
- [ ] 查看容器日志：`docker-compose logs`

### 数据库连接失败
- [ ] 检查MySQL容器是否运行：`docker ps | grep mysql`
- [ ] 检查环境变量配置：`cat .env`
- [ ] 检查数据库日志：`docker logs blog-mysql`

### 前端无法访问API
- [ ] 检查API服务是否运行：`docker ps | grep api`
- [ ] 检查API日志：`docker logs blog-api`
- [ ] 检查网络连接：`curl http://localhost:8080/api/posts?page=1&size=1`

## 联系支持

如遇到无法解决的问题，请提供：
- 服务器系统信息
- Docker版本信息
- 错误日志
- 部署步骤

