# API 接口文档

## 基础信息

- **Base URL**: `http://localhost:8080/api`
- **认证方式**: Bearer Token
- **Content-Type**: `application/json; charset=utf-8`

## 认证说明

### 获取Token

```bash
POST /api/users/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password"
}
```

**响应**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "admin",
    "role": "admin"
  }
}
```

### 使用Token

在请求头中添加：
```
Authorization: Bearer {token}
```

---

## 前端用户接口

### 1. 文章接口

#### 获取文章列表

```bash
GET /api/posts?page=1&size=10&q=搜索关键词&category=分类slug&tag=标签slug&sort=latest
```

**参数**:
- `page`: 页码（必需）
- `size`: 每页数量（必需）
- `q`: 搜索关键词（可选）
- `category`: 分类slug（可选）
- `tag`: 标签slug（可选）
- `sort`: 排序方式（可选：latest, popular）

**响应**:
```json
{
  "posts": [
    {
      "id": 1,
      "title": "文章标题",
      "slug": "article-slug",
      "excerpt": "文章摘要",
      "cover_image": "http://localhost:8080/uploads/images/xxx.jpg",
      "published_at": "2024-01-01T00:00:00Z",
      "view_count": 100,
      "category_names": ["技术", "编程"],
      "tag_names": ["Go", "Golang"]
    }
  ]
}
```

#### 获取文章详情

```bash
GET /api/posts/:id
```

**响应**:
```json
{
  "post": {
    "id": 1,
    "title": "文章标题",
    "slug": "article-slug",
    "content": "<p>文章内容...</p>",
    "excerpt": "文章摘要",
    "cover_image": "http://localhost:8080/uploads/images/xxx.jpg",
    "view_count": 100,
    "like_count": 10,
    "comment_count": 5,
    "published_at": "2024-01-01T00:00:00Z",
    "categories": [
      {
        "id": 1,
        "name": "技术",
        "slug": "tech"
      }
    ],
    "tags": [
      {
        "id": 1,
        "name": "Go",
        "slug": "go"
      }
    ],
    "category_ids": [1],
    "tag_ids": [1]
  },
  "categories": [...],
  "tags": [...]
}
```

### 2. 分类接口

```bash
GET /api/categories
```

**响应**:
```json
{
  "categories": [
    {
      "id": 1,
      "name": "技术",
      "slug": "tech",
      "description": "技术相关文章"
    }
  ]
}
```

### 3. 标签接口

```bash
GET /api/tags
```

**响应**:
```json
{
  "tags": [
    {
      "id": 1,
      "name": "Go",
      "slug": "go",
      "description": "Go语言"
    }
  ]
}
```

### 4. 评论接口

#### 获取文章评论

```bash
GET /api/comments?post_id=1
```

#### 创建评论

```bash
POST /api/comments
Content-Type: application/json

{
  "post_id": 1,
  "content": "评论内容",
  "author_name": "访客",
  "author_email": "guest@example.com",
  "parent_id": null  // 可选，回复评论时使用
}
```

### 5. 点赞接口

#### 切换点赞状态

```bash
POST /api/like/toggle
Content-Type: application/json

{
  "user_id": 1,
  "post_id": 1
}
```

#### 获取点赞数

```bash
GET /api/like/count?post_id=1
```

### 6. 页面接口

#### 获取页面（通过slug）

```bash
GET /api/pages/:slug_or_id
```

**示例**:
```bash
GET /api/pages/about-me
```

**响应**:
```json
{
  "page": {
    "id": 1,
    "title": "关于我",
    "slug": "about-me",
    "content": "<p>页面内容...</p>",
    "status": "published"
  }
}
```

### 7. 热点数据接口

```bash
GET /api/hotdata?data_type=trending_posts&period=all_time&limit=10
```

**参数**:
- `data_type`: 数据类型（可选：trending_posts, popular_tags, active_users）
- `period`: 统计周期（可选：daily, weekly, monthly, all_time）
- `limit`: 返回数量（可选，默认10，最多20）

**响应**:
```json
{
  "list": [
    {
      "id": 1,
      "data_type": "trending_posts",
      "data_key": "posts",
      "data_value": "[{\"id\":1,\"title\":\"文章1\"}]",
      "score": 100.5,
      "period": "all_time"
    }
  ]
}
```

---

## 后台管理接口

所有后台管理接口都需要管理员权限（Bearer Token）。

**Base URL**: `/api/admin`

### 1. 文章管理

#### 文章列表

```bash
GET /api/admin/posts?page=1&page_size=10&q=搜索&status=published&sort=latest
```

#### 创建文章

```bash
POST /api/admin/posts
Authorization: Bearer {token}
Content-Type: application/json

{
  "title": "文章标题",
  "content": "文章内容",
  "excerpt": "文章摘要",
  "cover_image": "http://localhost:8080/uploads/images/xxx.jpg",
  "image": "data:image/jpeg;base64,...",  // 可选，base64图片
  "category_ids": [1, 2],
  "tag_ids": [1, 2],
  "status": "published"
}
```

**支持FormData格式**（有图片文件时）:
```bash
POST /api/admin/posts
Authorization: Bearer {token}
Content-Type: multipart/form-data

image: [文件]
title: 文章标题
content: 文章内容
category_ids[]: 1
category_ids[]: 2
tag_ids[]: 1
status: published
```

#### 获取文章详情

```bash
GET /api/admin/posts/:id
```

#### 更新文章

```bash
PUT /api/admin/posts/:id
Authorization: Bearer {token}
Content-Type: application/json

{
  "title": "更新标题",
  "content": "更新内容",
  "category_ids": [1, 3],
  "tag_ids": [2],
  "status": "published"
}
```

#### 删除文章

```bash
DELETE /api/admin/posts/:id
```

### 2. 分类管理

#### 分类列表

```bash
GET /api/admin/categories
```

#### 创建分类

```bash
POST /api/admin/categories
Authorization: Bearer {token}
Content-Type: application/json

{
  "name": "技术",
  "slug": "tech",
  "description": "技术相关文章"
}
```

#### 更新分类

```bash
PUT /api/admin/categories/:id
```

#### 删除分类

```bash
DELETE /api/admin/categories/:id
```

### 3. 标签管理

#### 标签列表

```bash
GET /api/admin/tags
```

#### 创建标签

```bash
POST /api/admin/tags
Authorization: Bearer {token}
Content-Type: application/json

{
  "name": "Go",
  "slug": "go",
  "description": "Go语言"
}
```

#### 更新标签

```bash
PUT /api/admin/tags/:id
```

#### 删除标签

```bash
DELETE /api/admin/tags/:id
```

### 4. 评论管理

#### 评论列表

```bash
GET /api/admin/comments?page=1&page_size=10&post_id=1&status=approved&q=搜索
```

#### 获取评论详情

```bash
GET /api/admin/comments/:id
```

#### 更新评论状态

```bash
PUT /api/admin/comments/:id/status
Content-Type: application/json

{
  "status": "approved"  // pending, approved, rejected
}
```

#### 批量更新评论状态

```bash
PUT /api/admin/comments/status
Content-Type: application/json

{
  "ids": [1, 2, 3],
  "status": "approved"
}
```

#### 删除评论

```bash
DELETE /api/admin/comments/:id
```

#### 批量删除评论

```bash
DELETE /api/admin/comments
Content-Type: application/json

{
  "ids": [1, 2, 3]
}
```

#### 回复评论

```bash
POST /api/admin/comments/:id/reply
Content-Type: application/json

{
  "content": "回复内容"
}
```

### 5. 用户管理

#### 用户列表

```bash
GET /api/admin/users?page=1&page_size=10
```

#### 删除用户

```bash
DELETE /api/admin/users/:id
```

#### 更新用户状态

```bash
PUT /api/admin/users/:id/status
Content-Type: application/json

{
  "status": "active"  // active, inactive
}
```

#### 更新用户角色

```bash
PUT /api/admin/users/:id/role
Content-Type: application/json

{
  "role": "admin"  // user, admin
}
```

### 6. 页面管理

#### 页面列表

```bash
GET /api/admin/pages
```

#### 创建页面

```bash
POST /api/admin/pages
Authorization: Bearer {token}
Content-Type: application/json

{
  "title": "关于我",
  "slug": "about-me",
  "content": "<p>页面内容...</p>",
  "excerpt": "页面摘要",
  "status": "published"
}
```

#### 获取页面详情

```bash
GET /api/admin/pages/:id
```

#### 更新页面

```bash
PUT /api/admin/pages/:id
Authorization: Bearer {token}
Content-Type: application/json

{
  "title": "关于我",
  "slug": "about-me",
  "content": "<p>更新后的内容...</p>",
  "status": "published"
}
```

#### 删除页面

```bash
DELETE /api/admin/pages/:id
```

### 7. 文件上传

#### 上传图片

```bash
POST /api/admin/upload/image
Authorization: Bearer {token}
Content-Type: multipart/form-data

image: [文件]
```

**响应**:
```json
{
  "url": "http://localhost:8080/uploads/images/20251103-xxx.jpg",
  "path": "uploads/images/20251103-xxx.jpg",
  "filename": "20251103-xxx.jpg",
  "size": 12345,
  "type": "image/jpeg"
}
```

#### 上传文件

```bash
POST /api/admin/upload/file
Authorization: Bearer {token}
Content-Type: multipart/form-data

file: [文件]
```

#### 批量上传

```bash
POST /api/admin/upload/files
Authorization: Bearer {token}
Content-Type: multipart/form-data

files: [文件1]
files: [文件2]
```

#### 获取文件列表

```bash
GET /api/admin/upload/files?file_type=image&page=1&page_size=10
```

#### 删除文件

```bash
DELETE /api/admin/upload/file
Content-Type: application/json

{
  "path": "uploads/images/xxx.jpg"
}
```

---

## 数据结构

### 文章对象 (Post)

```typescript
interface Post {
  id: number;
  title: string;
  slug: string;
  content: string;
  excerpt: string;
  cover_image: string;        // 完整URL
  author_id: number;
  status: 'published' | 'draft' | 'pending' | 'trash';
  visibility: 'public' | 'private' | 'password';
  view_count: number;
  like_count: number;
  comment_count: number;
  published_at: string | null;
  created_at: string;
  updated_at: string;
  categories?: Category[];     // 完整的分类对象数组
  tags?: Tag[];              // 完整的标签对象数组
  category_ids?: number[];   // 分类ID数组
  tag_ids?: number[];        // 标签ID数组
}
```

### 分类对象 (Category)

```typescript
interface Category {
  id: number;
  name: string;
  slug: string;              // URL标识
  description?: string;
  parent_id?: number | null;
  sort_order?: number;
  post_count?: number;
  created_at: string;
  updated_at: string;
}
```

### 标签对象 (Tag)

```typescript
interface Tag {
  id: number;
  name: string;
  slug: string;              // URL标识
  description?: string;
  post_count?: number;
  created_at: string;
}
```

### 评论对象 (Comment)

```typescript
interface Comment {
  id: number;
  post_id: number;
  content: string;
  author_name: string;
  author_email: string;
  parent_id?: number | null;
  status: 'pending' | 'approved' | 'rejected';
  created_at: string;
  replies?: Comment[];       // 回复列表
}
```

### 页面对象 (Page)

```typescript
interface Page {
  id: number;
  title: string;
  slug: string;
  content: string;           // HTML格式
  excerpt?: string;
  status: 'published' | 'draft';
  created_at: string;
  updated_at: string;
}
```

---

## 图片URL说明

所有返回的图片URL都是完整的可访问地址：
- 开发环境: `http://localhost:8080/uploads/images/xxx.jpg`
- 生产环境: `https://api.example.com/uploads/images/xxx.jpg`

配置方式见 [README.md](./README.md) 中的配置说明。

---

## 错误码说明

- `200`: 成功
- `400`: 请求参数错误
- `401`: 未授权（需要登录）
- `403`: 权限不足（需要管理员权限）
- `404`: 资源不存在
- `500`: 服务器内部错误

---

## 注意事项

1. **认证**: 后台管理接口都需要Bearer Token认证
2. **权限**: 管理员接口需要 `role: admin` 的用户
3. **图片URL**: 所有图片URL自动转换为完整URL
4. **分页**: 列表接口支持分页，默认每页10条
5. **排序**: 支持按时间、热度等排序
6. **搜索**: 支持关键词搜索（标题、内容）

---

## 更新日志

- 2024-11-04: 优化文章详情接口，返回完整的分类和标签数据
- 2024-11-04: 添加页面管理功能
- 2024-11-04: 优化图片URL返回，支持完整URL
- 2024-11-04: 优化热点数据接口，默认返回前10条

