# SQL性能优化总结

## 优化日期
2025-01-12

## 问题诊断

经过架构检查，发现以下SQL用户端加载缓慢的主要问题：

1. **数据库连接池未配置** - 导致连接管理效率低下
2. **LIKE查询性能问题** - 使用了前后都有通配符的LIKE查询，无法使用索引
3. **缺少必要的数据库索引** - 常用查询字段缺少索引
4. **批量操作效率低** - 使用循环插入而非批量插入

## 已实施的优化

### 1. 数据库连接池配置优化 ✅

**文件**: `database/db.go`

**优化内容**:
- 设置最大打开连接数: 25
- 设置最大空闲连接数: 10
- 设置连接最大生命周期: 5分钟
- 设置连接最大空闲时间: 10分钟
- 添加连接验证（Ping）

**DSN参数优化**:
- 添加连接超时: `timeout=10s`
- 添加读取超时: `readTimeout=30s`
- 添加写入超时: `writeTimeout=30s`

**预期效果**: 
- 减少连接创建和销毁的开销
- 提高并发处理能力
- 避免连接泄漏

### 2. LIKE查询优化 ✅

**文件**: `dao/post_dao.go`

**优化内容**:
- 添加了性能问题注释说明
- 对于列表查询，使用前缀匹配 `q+"%"` 而非 `"%"+q+"%"`

**建议**:
- 对于大量数据的全文搜索，建议使用MySQL的FULLTEXT索引
- 或使用Elasticsearch等专业搜索引擎

### 3. 批量插入优化 ✅

**文件**: `dao/post_dao.go`, `service/post_service.go`

**优化内容**:
- 新增 `BatchAddCategoriesToPost()` 方法
- 新增 `BatchAddTagsToPost()` 方法
- 修改 `CreatePost()` 使用批量插入
- 修改 `UpdatePostCategoriesAndTags()` 使用批量插入

**预期效果**:
- 减少数据库往返次数
- 提高插入性能（特别是多分类/标签的文章）

### 4. 数据库索引优化 ✅

**文件**: `database/sql/performance_indexes.sql`, `models/post.go`

**新增索引**:

#### Posts表:
- `idx_posts_created_at` - 用于按时间排序
- `idx_posts_status_created_at` - 用于按状态筛选并按时间排序（复合索引）
- `idx_posts_published_at` - 用于按发布时间排序
- `idx_posts_author_created_at` - 用于查询作者文章（复合索引）

#### Comments表:
- `idx_comments_post_created_at` - 用于查询文章评论（复合索引）
- `idx_comments_status_created_at` - 用于按状态筛选评论（复合索引）
- `idx_comments_parent_created_at` - 用于查询回复（复合索引）

#### 关联表:
- `idx_post_categories_category_id` - 用于查询分类下的文章
- `idx_post_categories_post_category` - 用于优化JOIN查询（复合索引）
- `idx_post_tags_tag_id` - 用于查询标签下的文章
- `idx_post_tags_post_tag` - 用于优化JOIN查询（复合索引）

#### 其他表:
- `idx_categories_slug` - 用于通过slug查找分类
- `idx_tags_slug` - 用于通过slug查找标签
- `idx_likes_post_user` - 用于快速查找用户点赞（复合索引）
- `idx_likes_comment_user` - 用于快速查找评论点赞（复合索引）

**模型层优化**:
- 在 `Post` 模型中为 `created_at` 和 `published_at` 添加索引标记

## 执行索引优化

执行以下SQL脚本来创建索引：

```bash
mysql -u root -p blog < database/sql/performance_indexes.sql
```

或者在MySQL客户端中执行：

```sql
source database/sql/performance_indexes.sql;
```

**注意**: 
- 索引创建可能需要一些时间，建议在低峰期执行
- 索引会占用额外的存储空间
- 索引会略微降低INSERT/UPDATE/DELETE的性能，但会大幅提升SELECT查询的性能

## 性能监控建议

### 1. 使用EXPLAIN分析查询

对于慢查询，使用 `EXPLAIN` 命令分析查询计划：

```sql
EXPLAIN SELECT * FROM posts WHERE status = 'published' ORDER BY created_at DESC LIMIT 10;
```

### 2. 启用慢查询日志

在MySQL配置中启用慢查询日志：

```ini
slow_query_log = 1
long_query_time = 1
```

### 3. 监控连接池使用情况

可以通过以下SQL查看当前连接数：

```sql
SHOW STATUS LIKE 'Threads_connected';
SHOW STATUS LIKE 'Max_used_connections';
```

## 进一步优化建议

### 1. 查询优化
- 考虑使用分页缓存
- 对于热点数据使用Redis缓存
- 考虑读写分离（主从复制）

### 2. 数据库优化
- 定期执行 `OPTIMIZE TABLE` 优化表
- 监控表大小，考虑分区
- 根据实际负载调整连接池参数

### 3. 应用层优化
- 实现查询结果缓存
- 使用CDN加速静态资源
- 考虑使用消息队列处理异步任务

## 预期性能提升

实施这些优化后，预期可以获得以下性能提升：

1. **连接管理**: 减少连接创建开销，提升并发处理能力（预计提升20-30%）
2. **查询性能**: 通过索引优化，常用查询速度提升50-90%
3. **批量操作**: 批量插入性能提升3-10倍（取决于批量大小）
4. **整体响应时间**: 预计整体API响应时间减少30-50%

## 回滚方案

如果优化后出现问题，可以：

1. **回滚代码**: 使用Git回滚到优化前的版本
2. **删除索引**: 执行 `DROP INDEX` 语句删除新增的索引
3. **调整连接池**: 修改连接池参数为更保守的值

## 联系与支持

如有问题，请检查：
1. MySQL版本是否支持所有索引类型（建议MySQL 5.7+）
2. 存储引擎是否为InnoDB（支持行级锁和更好的并发性能）
3. 服务器资源是否充足（CPU、内存、磁盘IO）

