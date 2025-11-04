# 热点数据接口优化说明

## 优化内容

根据前端页面的热点数据需求，优化了后端热点数据接口，确保只返回前10条热点数据。

## 前端数据结构需求

根据 `Sidebar.vue` 的代码分析，前端期望的数据结构：

### 1. 返回格式
```json
{
  "list": [
    {
      "id": 1,
      "data_type": "trending_posts",
      "data_key": "...",
      "data_value": "[{\"id\":1,\"title\":\"文章标题\"}, ...]",
      "score": 100.5,
      "period": "all_time",
      "calculated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### 2. 前端处理逻辑
- 过滤 `data_type === 'trending_posts'` 的数据
- 解析 `data_value`（如果是字符串就JSON.parse）
- 如果是数组，展开并提取文章信息（id, title）
- 只取前6条显示

## 后端优化

### 1. DAO层 (`dao/hot_data_dao.go`)

**主要改动**：
- 添加 `limit` 参数，默认返回10条
- 按 `score` 降序排序
- 支持可选的 `dataType` 和 `period` 过滤

```go
func ListHotData(dataType, period string, limit int) ([]models.HotData, error) {
    var hd []models.HotData
    db := database.GetDB()
    
    // 如果指定了dataType，添加过滤条件
    if dataType != "" {
        db = db.Where("data_type = ?", dataType)
    }
    
    // 如果指定了period，添加过滤条件
    if period != "" {
        db = db.Where("period = ?", period)
    }
    
    // 设置默认limit为10
    if limit <= 0 {
        limit = 10
    }
    
    // 按score降序排序，限制返回数量
    err := db.Order("score DESC").Limit(limit).Find(&hd).Error
    return hd, err
}
```

### 2. Service层 (`service/hot_data_service.go`)

**主要改动**：
- 添加 `limit` 参数传递

```go
func ListHotData(dataType, period string, limit int) ([]models.HotData, error) {
    return dao.ListHotData(dataType, period, limit)
}
```

### 3. Controller层 (`controllers/hot_data_controller.go`)

**主要改动**：
- 默认返回前10条热点数据
- 支持可选的 `limit` 查询参数（最多20条）
- 保持返回格式 `{ list: [...] }` 符合前端期望

```go
func ListHotData(c *gin.Context) {
    // 获取查询参数
    dataType := c.Query("data_type")  // 可选：trending_posts, popular_tags, active_users
    period := c.Query("period")        // 可选：daily, weekly, monthly, all_time
    
    // 默认返回前10条热点数据
    limit := 10
    
    // 如果前端指定了limit参数，使用指定值（但最多不超过20条）
    if limitStr := c.Query("limit"); limitStr != "" {
        if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
            if parsedLimit > 20 {
                limit = 20 // 最多返回20条
            } else {
                limit = parsedLimit
            }
        }
    }
    
    list, err := service.ListHotData(dataType, period, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
        return
    }
    
    // 返回格式：{ list: [...] }，符合前端期望
    c.JSON(http.StatusOK, gin.H{"list": list})
}
```

## API使用示例

### 1. 获取所有类型的热点数据（默认前10条）

```bash
GET /api/hotdata
```

**响应**：
```json
{
  "list": [
    {
      "id": 1,
      "data_type": "trending_posts",
      "data_key": "posts",
      "data_value": "[{\"id\":1,\"title\":\"文章1\"},{\"id\":2,\"title\":\"文章2\"}]",
      "score": 100.5,
      "period": "all_time",
      "calculated_at": "2024-01-01T00:00:00Z"
    },
    ...
  ]
}
```

### 2. 获取指定类型的热点数据

```bash
GET /api/hotdata?data_type=trending_posts
```

### 3. 获取指定周期和类型的热点数据

```bash
GET /api/hotdata?data_type=trending_posts&period=weekly
```

### 4. 自定义返回数量（最多20条）

```bash
GET /api/hotdata?limit=15
```

## 数据格式说明

### data_value 格式

`data_value` 字段存储JSON格式的数据，对于 `trending_posts` 类型，应该是一个文章数组：

```json
[
  {
    "id": 1,
    "title": "文章标题",
    "slug": "article-slug",
    "cover_image": "http://localhost:8080/uploads/images/xxx.jpg",
    ...
  },
  {
    "id": 2,
    "title": "另一篇文章",
    ...
  }
]
```

### 创建热点数据示例

```bash
POST /api/hotdata
Content-Type: application/json

{
  "data_type": "trending_posts",
  "data_key": "posts",
  "data_value": "[{\"id\":1,\"title\":\"文章1\"},{\"id\":2,\"title\":\"文章2\"}]",
  "score": 100.5,
  "period": "all_time"
}
```

## 优化总结

✅ **默认返回10条**：按热度分数降序排序，只返回前10条  
✅ **支持过滤**：可按 `data_type` 和 `period` 过滤  
✅ **支持自定义数量**：可通过 `limit` 参数调整（最多20条）  
✅ **保持兼容**：返回格式 `{ list: [...] }` 符合前端期望  
✅ **性能优化**：使用数据库 `LIMIT` 限制，减少数据传输  

现在热点数据接口已经优化完成，只返回前10条热点数据，满足前端需求！🎉

