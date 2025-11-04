# 管理后台文章API文档

## 概述

管理后台文章API提供了完整的文章管理功能，包括创建、编辑、删除、列表查询等操作。所有接口都需要管理员权限（Bearer Token认证）。

## 基础路径

所有接口的基础路径：`/api/admin/posts`

## 认证

所有接口都需要在请求头中携带管理员Token：

```
Authorization: Bearer {your-admin-token}
```

## API列表

### 1. 获取文章列表

**接口**：`GET /api/admin/posts`

**功能**：获取文章列表，支持分页、搜索、筛选

**请求参数**（Query参数）：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认为1 |
| page_size | int | 否 | 每页数量，默认10，最大100 |
| size | int | 否 | 兼容旧参数名，等同于page_size |
| q | string | 否 | 搜索关键词（标题或内容） |
| sort | string | 否 | 排序方式，ASC或DESC，默认DESC |
| category | string | 否 | 分类slug筛选 |
| tag | string | 否 | 标签slug筛选 |
| status | string | 否 | 状态筛选：published/draft/pending/trash，不传则显示所有 |

**请求示例**：

```bash
GET /api/admin/posts?page=1&page_size=20&status=published&q=测试
```

**响应示例**：

```json
{
  "posts": [
    {
      "id": 1,
      "title": "测试文章",
      "slug": "test-post",
      "content": "文章内容...",
      "excerpt": "摘要",
      "cover_image": "https://example.com/image.jpg",
      "status": "published",
      "visibility": "public",
      "view_count": 100,
      "like_count": 10,
      "comment_count": 5,
      "author_id": 1,
      "published_at": "2024-01-01T10:00:00Z",
      "created_at": "2024-01-01T10:00:00Z",
      "updated_at": "2024-01-01T10:00:00Z",
      "category_names": ["技术", "编程"],
      "category_ids": [1, 2],
      "tag_names": ["Go", "Golang"],
      "tag_ids": [1, 2]
    }
  ],
  "total": 100,
  "page": 1,
  "page_size": 20,
  "total_pages": 5
}
```

---

### 2. 创建文章

**接口**：`POST /api/admin/posts`

**功能**：创建新文章

**请求体**：

```json
{
  "title": "文章标题",              // 必填
  "slug": "article-slug",          // 可选，不传则自动生成
  "content": "文章内容",            // 必填
  "excerpt": "文章摘要",            // 可选
  "cover_image": "封面图URL",       // 可选
  "status": "draft",               // 可选：draft/published/pending，默认draft
  "visibility": "public",          // 可选：public/private/password，默认public
  "categories": [1, 2],           // 可选：分类ID数组
  "tags": [1, 2, 3],              // 可选：标签ID数组
  "author_id": 1                  // 可选：作者ID，不传则使用当前登录用户
}
```

**请求示例**：

```bash
curl -X POST http://localhost:8080/api/admin/posts \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "我的新文章",
    "content": "这是文章内容",
    "excerpt": "这是摘要",
    "status": "draft",
    "categories": [1, 2],
    "tags": [1, 2]
  }'
```

**响应示例**：

```json
{
  "post": {
    "id": 1,
    "title": "我的新文章",
    "slug": "wo-de-xin-wen-zhang",  // 自动生成的slug
    "content": "这是文章内容",
    "excerpt": "这是摘要",
    "cover_image": "",
    "status": "draft",
    "visibility": "public",
    "author_id": 1,
    "published_at": null,
    "created_at": "2024-01-01T10:00:00Z",
    "updated_at": "2024-01-01T10:00:00Z"
  }
}
```

**说明**：

- `slug`如果未提供，会自动根据标题生成（中文转拼音或使用时间戳）
- `slug`会自动确保唯一性（如果已存在会添加数字后缀）
- 如果`status`为`published`，会自动设置`published_at`时间
- 如果不指定`author_id`，会使用当前登录用户的ID

---

### 3. 获取文章详情

**接口**：`GET /api/admin/posts/:id`

**功能**：获取指定文章的详细信息（包含分类和标签ID）

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | uint64 | 文章ID |

**请求示例**：

```bash
GET /api/admin/posts/1
```

**响应示例**：

```json
{
  "post": {
    "id": 1,
    "title": "我的新文章",
    "slug": "wo-de-xin-wen-zhang",
    "content": "这是文章内容",
    "excerpt": "这是摘要",
    "cover_image": "",
    "status": "published",
    "visibility": "public",
    "author_id": 1,
    "view_count": 100,
    "like_count": 10,
    "comment_count": 5,
    "published_at": "2024-01-01T10:00:00Z",
    "created_at": "2024-01-01T10:00:00Z",
    "updated_at": "2024-01-01T10:00:00Z"
  },
  "category_ids": [1, 2],
  "tag_ids": [1, 2, 3]
}
```

---

### 4. 更新文章

**接口**：`PUT /api/admin/posts/:id`

**功能**：更新文章信息

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | uint64 | 文章ID |

**请求体**（所有字段都是可选的，只传需要更新的字段）：

```json
{
  "title": "更新后的标题",
  "slug": "updated-slug",
  "content": "更新后的内容",
  "excerpt": "更新后的摘要",
  "cover_image": "新的封面图URL",
  "status": "published",
  "visibility": "private",
  "categories": [1, 3],      // 传入新的分类ID数组，会替换所有旧关联
  "tags": [2, 4]             // 传入新的标签ID数组，会替换所有旧关联
}
```

**请求示例**：

```bash
curl -X PUT http://localhost:8080/api/admin/posts/1 \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "更新后的标题",
    "status": "published",
    "categories": [1, 3],
    "tags": [2, 4]
  }'
```

**响应示例**：

```json
{
  "post": {
    "id": 1,
    "title": "更新后的标题",
    "slug": "geng-xin-hou-de-biao-ti",
    "content": "更新后的内容",
    "status": "published",
    ...
  }
}
```

**说明**：

- 如果只更新`title`而不传`slug`，会自动根据新标题生成新的slug
- 如果更新`slug`，会自动验证唯一性并确保不重复
- 如果更新`status`为`published`且之前不是published，会自动设置`published_at`
- 如果传入`categories`或`tags`数组（即使是空数组），会替换所有旧关联
- 如果不传`categories`或`tags`，则不会更新关联关系

---

### 5. 删除文章

**接口**：`DELETE /api/admin/posts/:id`

**功能**：删除指定文章（会同时删除文章的分类和标签关联）

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | uint64 | 文章ID |

**请求示例**：

```bash
curl -X DELETE http://localhost:8080/api/admin/posts/1 \
  -H "Authorization: Bearer your-token"
```

**响应示例**：

```json
{
  "message": "删除成功"
}
```

---

## Slug自动生成规则

1. 将标题转为小写
2. 替换空格为短横线`-`
3. 移除特殊字符，只保留字母、数字、短横线
4. 移除连续的短横线
5. 移除首尾短横线
6. 如果结果为空，使用时间戳：`post-{timestamp}`
7. 如果slug已存在，添加数字后缀：`slug-1`, `slug-2`等

**示例**：

- `"我的新文章"` → `"wo-de-xin-wen-zhang"` 或 `"post-1234567890"`
- `"Hello World!"` → `"hello-world"`
- `"Test Article"` → `"test-article"`

---

## 状态说明

### 文章状态（status）

| 值 | 说明 |
|----|------|
| draft | 草稿（默认） |
| published | 已发布 |
| pending | 待审核 |
| trash | 已删除（回收站） |

### 可见性（visibility）

| 值 | 说明 |
|----|------|
| public | 公开（默认） |
| private | 私有 |
| password | 密码保护 |

---

## 错误处理

所有接口的错误响应格式：

```json
{
  "error": "错误描述信息"
}
```

常见错误码：

- `400 Bad Request` - 请求参数错误
- `401 Unauthorized` - 未认证或token无效
- `403 Forbidden` - 权限不足（非管理员）
- `404 Not Found` - 文章不存在
- `500 Internal Server Error` - 服务器内部错误

---

## 最佳实践

1. **创建文章时**：
   - 可以不传`slug`，让系统自动生成
   - 创建时建议使用`draft`状态，编辑完成后再改为`published`
   - 创建后立即保存返回的`id`，用于后续更新

2. **更新文章时**：
   - 使用PATCH语义：只传需要更新的字段
   - 更新分类和标签时，传入完整的新数组（会替换所有旧关联）
   - 如果只想清空分类或标签，传入空数组`[]`

3. **列表查询时**：
   - 使用`page_size`控制每页数量，建议10-20
   - 使用`status`筛选不同状态的文章
   - 使用`q`进行全文搜索

4. **性能优化**：
   - 列表接口支持分页，避免一次查询过多数据
   - 可以结合缓存减少数据库查询

---

## 更新日志

### v1.0.0 (2024-01-01)
- ✅ 支持文章列表分页查询
- ✅ 支持文章创建（自动生成slug）
- ✅ 支持文章详情获取（包含分类标签ID）
- ✅ 支持文章更新（支持分类标签更新）
- ✅ 支持文章删除
- ✅ 支持状态筛选和搜索

