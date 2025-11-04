# 文章创建和更新接口修复说明

## 问题描述

错误：`保存失败: 参数格式错误: invalid character '-' in numeric literal`

原因：后端接口字段名与前端不匹配，导致JSON解析失败。

## 修复内容

### 1. 修复UpdatePost函数

**问题**：
- 使用了错误的字段名：`categories` 和 `tags`
- 前端使用的是：`category_ids` 和 `tag_ids`
- 有测试代码导致提前return

**修复**：
- ✅ 字段名改为 `category_ids` 和 `tag_ids`
- ✅ 支持FormData和JSON两种格式
- ✅ 移除测试代码
- ✅ 添加cover_image字段处理

### 2. 前端数据格式

前端发送的数据格式：

**JSON格式**（无文件时）：
```json
{
  "title": "标题",
  "content": "内容",
  "excerpt": "摘要",
  "cover_image": "URL或空字符串",
  "category_ids": [1, 2],  // ✅ 使用category_ids
  "tag_ids": [1, 2],       // ✅ 使用tag_ids
  "status": "published"
}
```

**FormData格式**（有文件时）：
```
image: [文件对象]
title: "标题"
content: "内容"
excerpt: "摘要"
category_ids[]: "1"
category_ids[]: "2"
tag_ids[]: "1"
tag_ids[]: "2"
status: "published"
```

## 接口说明

### 创建文章 POST /api/admin/posts

**支持两种格式**：

1. **FormData**（有image文件时）
   - Content-Type: `multipart/form-data`
   - 字段：`image`（文件），`title`, `content`, `excerpt`, `status`
   - 数组字段：`category_ids[]` 或 `category_ids`, `tag_ids[]` 或 `tag_ids`

2. **JSON**（无文件时）
   - Content-Type: `application/json; charset=utf-8`
   - 字段：`title`, `content`, `excerpt`, `cover_image`, `status`
   - 数组字段：`category_ids`, `tag_ids`

### 更新文章 PUT /api/admin/posts/:id

**支持两种格式**（与创建接口相同）：

1. **FormData**（有image文件时）
2. **JSON**（无文件时）

## 字段对照表

| 前端字段 | 后端字段（JSON） | 后端字段（FormData） | 说明 |
|---------|----------------|-------------------|------|
| category_ids | category_ids | category_ids[] 或 category_ids | 分类ID数组 |
| tag_ids | tag_ids | tag_ids[] 或 tag_ids | 标签ID数组 |
| image | - | image | 图片文件（仅FormData） |
| cover_image | cover_image | cover_image | 封面图URL（仅JSON） |

## 测试验证

### 创建文章（无文件）

```bash
curl -X POST http://localhost:8080/api/admin/posts \
  -H "Authorization: Bearer token" \
  -H "Content-Type: application/json; charset=utf-8" \
  -d '{
    "title": "测试文章",
    "content": "内容",
    "category_ids": [1, 2],
    "tag_ids": [1],
    "status": "draft"
  }'
```

### 创建文章（有文件）

```bash
curl -X POST http://localhost:8080/api/admin/posts \
  -H "Authorization: Bearer token" \
  -F "title=测试文章" \
  -F "content=内容" \
  -F "category_ids[]=1" \
  -F "category_ids[]=2" \
  -F "tag_ids[]=1" \
  -F "image=@/path/to/image.jpg"
```

### 更新文章（无文件）

```bash
curl -X PUT http://localhost:8080/api/admin/posts/1 \
  -H "Authorization: Bearer token" \
  -H "Content-Type: application/json; charset=utf-8" \
  -d '{
    "title": "更新标题",
    "category_ids": [1, 3],
    "tag_ids": [2]
  }'
```

## 修复总结

✅ **字段名匹配**：后端字段名与前端完全一致  
✅ **格式支持**：支持FormData和JSON两种格式  
✅ **代码清理**：移除测试代码和未使用的逻辑  
✅ **功能完整**：文件上传、分类标签更新都正常工作  

现在接口应该可以正常工作了！🎉

