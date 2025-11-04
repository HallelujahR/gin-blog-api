# 文章更新接口优化文档

## 问题描述

后台文章编辑页面发送的参数格式可能与当前接口不匹配，导致更新失败。

## 优化内容

### 1. 支持多种字段名格式

**优化前**：只支持固定的字段名
```json
{
  "categories": [1, 2],
  "tags": [1, 2]
}
```

**优化后**：支持多种字段名格式，兼容不同前端框架的命名习惯
- `categories` / `category_ids` / `categoryIds`
- `tags` / `tag_ids` / `tagIds`

### 2. 支持多种数据类型

**优化前**：只支持 `[]uint64`

**优化后**：自动转换多种数据类型
- `[]uint64` - 直接使用
- `[]float64` - JSON数字类型（自动转换）
- `[]string` - 字符串数组（自动解析）
- `[]interface{}` - 混合类型数组（自动转换）
- `[]int` / `[]int64` - 整数数组（自动转换）

### 3. 改进字段更新逻辑

**优化前**：只更新非空字段，无法清空字段值

**优化后**：
- 所有字段都可以更新（包括清空）
- 使用 `map[string]interface{}` 灵活接收参数
- 只在字段存在时才更新，不存在的字段保持原值

### 4. 改进响应格式

**优化后**：返回更新后的完整信息
```json
{
  "post": {
    // 更新后的文章信息
  },
  "categories": [
    // 更新后的分类信息
  ],
  "tags": [
    // 更新后的标签信息
  ]
}
```

## 支持的参数格式

### 格式1：标准格式
```json
{
  "title": "新标题",
  "content": "新内容",
  "categories": [1, 2, 3],
  "tags": [1, 2]
}
```

### 格式2：下划线格式
```json
{
  "title": "新标题",
  "category_ids": [1, 2, 3],
  "tag_ids": [1, 2]
}
```

### 格式3：驼峰格式
```json
{
  "title": "新标题",
  "categoryIds": [1, 2, 3],
  "tagIds": [1, 2]
}
```

### 格式4：字符串数组
```json
{
  "categories": ["1", "2", "3"],
  "tags": ["1", "2"]
}
```

### 格式5：混合类型
```json
{
  "categories": [1, "2", 3.0],
  "tags": [1, 2]
}
```

## 字段说明

### 基本字段

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| title | string | 否 | 文章标题 |
| slug | string | 否 | URL标识（不传则自动生成） |
| content | string | 否 | 文章内容（可清空） |
| excerpt | string | 否 | 文章摘要（可清空） |
| cover_image | string | 否 | 封面图URL（可清空） |
| status | string | 否 | 状态：draft/published/pending/trash |
| visibility | string | 否 | 可见性：public/private/password |

### 关联字段

| 字段名 | 类型 | 必填 | 说明 | 支持的格式 |
|--------|------|------|------|-----------|
| categories | array | 否 | 分类ID数组 | `categories`, `category_ids`, `categoryIds` |
| tags | array | 否 | 标签ID数组 | `tags`, `tag_ids`, `tagIds` |

**注意**：
- 传入空数组 `[]` 会清空所有关联
- 不传字段则保持原有关联不变
- 传入新数组会替换所有旧关联

## 使用示例

### 示例1：更新基本字段
```bash
curl -X PUT http://localhost:8080/api/admin/posts/1 \
  -H "Authorization: Bearer token" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "更新的标题",
    "content": "更新的内容",
    "status": "published"
  }'
```

### 示例2：更新分类和标签（标准格式）
```bash
curl -X PUT http://localhost:8080/api/admin/posts/1 \
  -H "Authorization: Bearer token" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "新标题",
    "categories": [1, 2, 3],
    "tags": [1, 2]
  }'
```

### 示例3：更新分类和标签（下划线格式）
```bash
curl -X PUT http://localhost:8080/api/admin/posts/1 \
  -H "Authorization: Bearer token" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "新标题",
    "category_ids": [1, 2],
    "tag_ids": [1, 2, 3]
  }'
```

### 示例4：更新分类和标签（驼峰格式）
```bash
curl -X PUT http://localhost:8080/api/admin/posts/1 \
  -H "Authorization: Bearer token" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "新标题",
    "categoryIds": [1, 2],
    "tagIds": [1, 2]
  }'
```

### 示例5：清空分类和标签
```bash
curl -X PUT http://localhost:8080/api/admin/posts/1 \
  -H "Authorization: Bearer token" \
  -H "Content-Type: application/json" \
  -d '{
    "categories": [],
    "tags": []
  }'
```

### 示例6：清空摘要和封面
```bash
curl -X PUT http://localhost:8080/api/admin/posts/1 \
  -H "Authorization: Bearer token" \
  -H "Content-Type: application/json" \
  -d '{
    "excerpt": "",
    "cover_image": ""
  }'
```

## 前端使用示例

### React/Vue示例
```javascript
// 更新文章
const updatePost = async (postId, formData) => {
  const response = await fetch(`/api/admin/posts/${postId}`, {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      title: formData.title,
      content: formData.content,
      excerpt: formData.excerpt,
      cover_image: formData.coverImage,
      status: formData.status,
      // 支持多种字段名格式
      categories: formData.categoryIds,  // 或 category_ids, categoryIds
      tags: formData.tagIds              // 或 tag_ids, tagIds
    })
  });
  
  const data = await response.json();
  return data;
};
```

### 支持字符串数组的情况
```javascript
// 如果前端发送的是字符串数组，也能自动转换
const formData = {
  categories: ["1", "2", "3"],  // 字符串数组
  tags: ["1", "2"]              // 字符串数组
};

// 后端会自动转换为数字数组 [1, 2, 3] 和 [1, 2]
```

## 响应格式

### 成功响应
```json
{
  "post": {
    "id": 1,
    "title": "更新的标题",
    "slug": "geng-xin-de-biao-ti",
    "content": "更新的内容",
    "excerpt": "",
    "cover_image": "",
    "status": "published",
    "visibility": "public",
    "category_ids": [1, 2, 3],
    "tag_ids": [1, 2],
    ...
  },
  "categories": [
    {
      "id": 1,
      "name": "技术",
      "slug": "tech",
      ...
    },
    {
      "id": 2,
      "name": "编程",
      "slug": "programming",
      ...
    }
  ],
  "tags": [
    {
      "id": 1,
      "name": "Go",
      "slug": "go",
      ...
    },
    {
      "id": 2,
      "name": "Golang",
      "slug": "golang",
      ...
    }
  ]
}
```

### 错误响应

**文章不存在**：
```json
{
  "error": "文章不存在"
}
```

**参数格式错误**：
```json
{
  "error": "参数格式错误: ..."
}
```

**更新失败**：
```json
{
  "error": "更新失败: ..."
}
```

## 核心优化点

### 1. 灵活的字段名匹配
- 支持 `categories` / `category_ids` / `categoryIds`
- 支持 `tags` / `tag_ids` / `tagIds`
- 自动识别并处理

### 2. 智能类型转换
- 自动将 JSON 数字（float64）转换为 uint64
- 自动解析字符串数组
- 支持混合类型数组

### 3. 完善的空值处理
- 空字符串可以清空字段
- 空数组可以清空关联
- 不传字段保持原值

### 4. 更好的错误提示
- 明确的错误信息
- 区分不同类型的错误

## 兼容性说明

✅ **完全向后兼容**：
- 原有的参数格式仍然支持
- 新增的格式只是扩展，不影响旧接口

✅ **前端框架兼容**：
- React/Vue（驼峰命名）
- Angular（下划线命名）
- 原生 JavaScript（任意格式）

## 测试建议

1. **测试基本更新**：更新标题、内容等基本字段
2. **测试分类标签更新**：使用不同的字段名格式
3. **测试类型转换**：发送字符串数组、数字数组等
4. **测试清空功能**：发送空数组、空字符串
5. **测试部分更新**：只更新部分字段

## 总结

本次优化使文章更新接口：
- ✅ 更加灵活，支持多种参数格式
- ✅ 更加健壮，自动处理类型转换
- ✅ 更加易用，支持清空字段
- ✅ 更加友好，返回完整更新后的数据

无论前端使用什么格式发送数据，接口都能正确处理！🎉

