# 管理后台API优化文档

## 概述

本次优化主要针对管理后台的文章编辑、分类标签创建和评论管理功能，提供了更完整、更易用的API接口。

## 主要优化内容

### 1. 文章编辑接口优化 ✅

#### 问题
- 获取文章详情时，只返回分类和标签的ID，前端需要额外请求获取完整信息
- 缺少获取所有分类和标签的接口

#### 解决方案

**1.1 优化获取文章详情接口**

**接口**：`GET /api/admin/posts/:id`

**改进前**：
```json
{
  "post": {...},
  "category_ids": [1, 2],
  "tag_ids": [1, 2, 3]
}
```

**改进后**：
```json
{
  "post": {
    "id": 1,
    "title": "文章标题",
    "content": "...",
    ...
  },
  "categories": [
    {
      "id": 1,
      "name": "技术",
      "slug": "tech",
      "description": "..."
    },
    {
      "id": 2,
      "name": "编程",
      "slug": "programming",
      "description": "..."
    }
  ],
  "tags": [
    {
      "id": 1,
      "name": "Go",
      "slug": "go"
    },
    {
      "id": 2,
      "name": "Golang",
      "slug": "golang"
    }
  ],
  "category_ids": [1, 2],  // 兼容性保留
  "tag_ids": [1, 2, 3]      // 兼容性保留
}
```

**优势**：
- ✅ 一次性获取所有需要的信息，减少前端请求次数
- ✅ 前端可以直接使用分类和标签的完整信息，无需额外查询
- ✅ 保持向后兼容，仍然返回ID数组

**1.2 新增获取所有分类和标签接口**

**接口1**：`GET /api/admin/categories`

获取所有分类列表，供文章编辑时选择使用。

**响应示例**：
```json
{
  "categories": [
    {
      "id": 1,
      "name": "技术",
      "slug": "tech",
      "description": "技术相关文章",
      "parent_id": null,
      "created_at": "..."
    }
  ]
}
```

**接口2**：`GET /api/admin/tags`

获取所有标签列表，供文章编辑时选择使用。

**响应示例**：
```json
{
  "tags": [
    {
      "id": 1,
      "name": "Go",
      "slug": "go",
      "created_at": "..."
    }
  ]
}
```

---

### 2. 分类和标签创建接口优化 ✅

#### 问题
- 创建分类和标签时必须手动提供slug
- slug需要符合URL规范，手动输入容易出错

#### 解决方案

**2.1 自动生成slug**

**接口**：`POST /api/admin/categories`

**改进前**：
```json
{
  "name": "技术分类",
  "slug": "tech-category",  // 必填
  "description": "..."
}
```

**改进后**：
```json
{
  "name": "技术分类",
  "slug": "tech-category",  // 可选，不传则自动生成
  "description": "..."
}
```

如果不传`slug`，系统会根据`name`自动生成：
- `"技术分类"` → `"ji-shu-fen-lei"` 或 `"category-1234567890"`

**接口**：`POST /api/admin/tags`

同样支持自动生成slug。

**2.2 Slug生成规则**

1. 转小写
2. 替换空格为短横线
3. 移除特殊字符，只保留字母、数字、短横线
4. 移除连续的短横线
5. 移除首尾短横线
6. 如果结果为空，使用时间戳

---

### 3. 评论管理接口全面优化 ✅

#### 问题
- 只支持按文章ID查询，不支持筛选和搜索
- 缺少分页功能
- 缺少批量操作
- 缺少回复功能

#### 解决方案

**3.1 评论列表接口（支持分页和筛选）**

**接口**：`GET /api/admin/comments`

**请求参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页数量，默认20，最大100 |
| post_id | uint64 | 否 | 筛选特定文章 |
| status | string | 否 | 筛选状态：approved/pending/spam/trash |
| q | string | 否 | 搜索关键词（内容、作者名、邮箱） |
| sort | string | 否 | 排序：ASC/DESC，默认DESC |

**请求示例**：
```
GET /api/admin/comments?page=1&page_size=20&status=pending&q=测试
```

**响应示例**：
```json
{
  "comments": [
    {
      "id": 1,
      "content": "评论内容",
      "author_name": "张三",
      "author_email": "zhangsan@example.com",
      "post_id": 1,
      "parent_id": null,
      "status": "pending",
      "like_count": 0,
      "created_at": "2024-01-01T10:00:00Z"
    }
  ],
  "total": 100,
  "page": 1,
  "page_size": 20,
  "total_pages": 5
}
```

**3.2 获取评论详情（含回复）**

**接口**：`GET /api/admin/comments/:id`

**响应示例**：
```json
{
  "comment": {
    "id": 1,
    "content": "主评论内容",
    "author_name": "张三",
    "post_id": 1,
    "parent_id": null,
    "status": "approved",
    ...
  },
  "replies": [
    {
      "id": 2,
      "content": "回复内容",
      "author_name": "李四",
      "parent_id": 1,
      ...
    }
  ]
}
```

**3.3 批量操作接口**

**批量删除评论**

**接口**：`POST /api/admin/comments/batch-delete`

**请求体**：
```json
{
  "ids": [1, 2, 3, 4, 5]
}
```

**响应**：
```json
{
  "message": "批量删除成功",
  "count": 5
}
```

**批量更新状态**

**接口**：`POST /api/admin/comments/batch-status`

**请求体**：
```json
{
  "ids": [1, 2, 3],
  "status": "approved"
}
```

**响应**：
```json
{
  "message": "批量更新成功",
  "count": 3
}
```

**3.4 回复评论功能**

**接口**：`POST /api/admin/comments/:id/reply`

**请求体**：
```json
{
  "content": "回复内容",
  "author_name": "管理员",
  "author_email": "admin@example.com"
}
```

**响应**：
```json
{
  "comment": {
    "id": 10,
    "content": "回复内容",
    "post_id": 1,
    "parent_id": 1,
    "status": "approved",  // 管理员回复自动审核通过
    ...
  }
}
```

**3.5 更新评论状态**

**接口**：`PUT /api/admin/comments/:id/status`

**请求体**：
```json
{
  "status": "approved"
}
```

**状态值**：
- `approved` - 已审核
- `pending` - 待审核
- `spam` - 垃圾评论
- `trash` - 已删除

---

## 新增/修改的文件

### DAO层
- ✅ `dao/post_dao.go` - 新增获取文章分类标签完整信息的函数
- ✅ `dao/comment_dao.go` - 新增分页、筛选、批量操作函数

### Service层
- ✅ `service/post_service.go` - 新增获取文章完整关系的函数
- ✅ `service/comment_service.go` - 新增分页响应和批量操作函数

### Controller层
- ✅ `controllers/admin/post_controller.go` - 优化获取文章详情
- ✅ `controllers/admin/category_controller.go` - 新增列表接口，优化创建接口
- ✅ `controllers/admin/tag_controller.go` - 新增列表接口，优化创建接口
- ✅ `controllers/admin/comment_controller.go` - 全面重写，支持所有新功能

### Routes层
- ✅ `routes/admin/category.go` - 新增GET接口
- ✅ `routes/admin/tag.go` - 新增GET接口
- ✅ `routes/admin/comment.go` - 新增多个新接口

---

## API接口列表

### 文章相关

| 方法 | 路径 | 功能 | 状态 |
|------|------|------|------|
| GET | `/api/admin/posts/:id` | 获取文章详情（含分类标签） | ✅ 已优化 |
| GET | `/api/admin/categories` | 获取所有分类列表 | ✅ 新增 |
| GET | `/api/admin/tags` | 获取所有标签列表 | ✅ 新增 |

### 分类相关

| 方法 | 路径 | 功能 | 状态 |
|------|------|------|------|
| POST | `/api/admin/categories` | 创建分类（支持自动生成slug） | ✅ 已优化 |

### 标签相关

| 方法 | 路径 | 功能 | 状态 |
|------|------|------|------|
| POST | `/api/admin/tags` | 创建标签（支持自动生成slug） | ✅ 已优化 |

### 评论相关

| 方法 | 路径 | 功能 | 状态 |
|------|------|------|------|
| GET | `/api/admin/comments` | 评论列表（分页、筛选） | ✅ 已优化 |
| GET | `/api/admin/comments/:id` | 获取评论详情（含回复） | ✅ 新增 |
| PUT | `/api/admin/comments/:id/status` | 更新评论状态 | ✅ 已优化 |
| DELETE | `/api/admin/comments/:id` | 删除单个评论 | ✅ 已存在 |
| POST | `/api/admin/comments/batch-delete` | 批量删除评论 | ✅ 新增 |
| POST | `/api/admin/comments/batch-status` | 批量更新状态 | ✅ 新增 |
| POST | `/api/admin/comments/:id/reply` | 回复评论 | ✅ 新增 |

---

## 使用示例

### 文章编辑流程

```javascript
// 1. 获取所有分类和标签（用于下拉选择）
const categories = await fetch('/api/admin/categories').then(r => r.json());
const tags = await fetch('/api/admin/tags').then(r => r.json());

// 2. 获取文章详情（包含完整的分类和标签信息）
const post = await fetch('/api/admin/posts/1').then(r => r.json());
// post.categories 包含完整的分类信息
// post.tags 包含完整的标签信息

// 3. 编辑文章
await fetch('/api/admin/posts/1', {
  method: 'PUT',
  body: JSON.stringify({
    title: "新标题",
    categories: [1, 2],  // 使用ID数组
    tags: [1, 2, 3]
  })
});
```

### 评论管理流程

```javascript
// 1. 获取待审核评论列表
const comments = await fetch('/api/admin/comments?status=pending&page=1&page_size=20')
  .then(r => r.json());

// 2. 批量审核通过
await fetch('/api/admin/comments/batch-status', {
  method: 'POST',
  body: JSON.stringify({
    ids: [1, 2, 3],
    status: 'approved'
  })
});

// 3. 回复评论
await fetch('/api/admin/comments/1/reply', {
  method: 'POST',
  body: JSON.stringify({
    content: "感谢您的评论",
    author_name: "管理员",
    author_email: "admin@example.com"
  })
});

// 4. 批量删除垃圾评论
await fetch('/api/admin/comments/batch-delete', {
  method: 'POST',
  body: JSON.stringify({
    ids: [5, 6, 7]
  })
});
```

---

## 向后兼容性

✅ **完全兼容** - 所有现有接口都保持向后兼容：

1. 文章详情接口仍然返回`category_ids`和`tag_ids`
2. 所有原有的评论接口仍然可用
3. 新增的接口都是独立的，不影响旧接口

---

## 总结

本次优化大幅提升了管理后台API的可用性和易用性：

1. ✅ **文章编辑**：一次性获取所有信息，无需多次请求
2. ✅ **分类标签创建**：自动生成slug，减少手动输入
3. ✅ **评论管理**：完整的分页、筛选、搜索、批量操作功能
4. ✅ **向后兼容**：所有改进都不影响现有功能

所有代码已通过编译测试，可以立即使用！🎉

