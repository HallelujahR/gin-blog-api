-- 性能优化索引脚本
-- 这些索引用于优化常用查询的性能
-- 执行前请先备份数据库

-- ========== Posts表索引优化 ==========

-- 1. created_at索引：用于按时间排序（最常用的查询）
-- 如果已存在则忽略错误
CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at DESC);

-- 2. status + created_at复合索引：用于按状态筛选并按时间排序（列表查询常用）
CREATE INDEX IF NOT EXISTS idx_posts_status_created_at ON posts(status, created_at DESC);

-- 3. published_at索引：用于按发布时间排序
CREATE INDEX IF NOT EXISTS idx_posts_published_at ON posts(published_at DESC);

-- 4. author_id + created_at复合索引：用于查询某个作者的文章
CREATE INDEX IF NOT EXISTS idx_posts_author_created_at ON posts(author_id, created_at DESC);

-- 5. slug索引：虽然已有uniqueIndex，但确保存在（用于快速查找）
-- slug字段已有uniqueIndex，这里不需要额外索引

-- ========== Comments表索引优化 ==========

-- 6. post_id + created_at复合索引：用于查询某篇文章的评论并按时间排序
CREATE INDEX IF NOT EXISTS idx_comments_post_created_at ON comments(post_id, created_at DESC);

-- 7. status + created_at复合索引：用于按状态筛选评论
CREATE INDEX IF NOT EXISTS idx_comments_status_created_at ON comments(status, created_at DESC);

-- 8. parent_id + created_at复合索引：用于查询回复并按时间排序
CREATE INDEX IF NOT EXISTS idx_comments_parent_created_at ON comments(parent_id, created_at ASC);

-- ========== Post_Categories表索引优化 ==========

-- 9. category_id索引：用于查询某个分类下的所有文章
CREATE INDEX IF NOT EXISTS idx_post_categories_category_id ON post_categories(category_id);

-- 10. post_id + category_id复合索引：用于快速查找关联关系
-- post_id和category_id已有单独索引，但复合索引可以优化JOIN查询
CREATE INDEX IF NOT EXISTS idx_post_categories_post_category ON post_categories(post_id, category_id);

-- ========== Post_Tags表索引优化 ==========

-- 11. tag_id索引：用于查询某个标签下的所有文章
CREATE INDEX IF NOT EXISTS idx_post_tags_tag_id ON post_tags(tag_id);

-- 12. post_id + tag_id复合索引：用于快速查找关联关系
CREATE INDEX IF NOT EXISTS idx_post_tags_post_tag ON post_tags(post_id, tag_id);

-- ========== Categories表索引优化 ==========

-- 13. slug索引：用于通过slug快速查找分类
CREATE INDEX IF NOT EXISTS idx_categories_slug ON categories(slug);

-- 14. parent_id索引：用于查询子分类（已有，但确保存在）
-- parent_id字段已有index，这里不需要额外索引

-- ========== Tags表索引优化 ==========

-- 15. slug索引：用于通过slug快速查找标签
CREATE INDEX IF NOT EXISTS idx_tags_slug ON tags(slug);

-- ========== Likes表索引优化 ==========

-- 16. post_id + user_id复合索引：用于快速查找用户是否已点赞某篇文章
CREATE INDEX IF NOT EXISTS idx_likes_post_user ON likes(post_id, user_id);

-- 17. comment_id + user_id复合索引：用于快速查找用户是否已点赞某条评论
CREATE INDEX IF NOT EXISTS idx_likes_comment_user ON likes(comment_id, user_id);

-- ========== 全文索引（可选，用于全文搜索优化） ==========

-- 注意：全文索引需要MyISAM或InnoDB（MySQL 5.6+）存储引擎
-- 如果使用InnoDB，需要MySQL 5.6+版本

-- 18. Posts表的全文索引：用于优化标题和内容的全文搜索
-- ALTER TABLE posts ADD FULLTEXT INDEX ft_posts_title_content (title, content);

-- ========== 索引使用说明 ==========
-- 1. 索引会占用额外的存储空间
-- 2. 索引会略微降低INSERT/UPDATE/DELETE的性能
-- 3. 但会大幅提升SELECT查询的性能
-- 4. 建议在低峰期执行此脚本
-- 5. 执行后可以通过 EXPLAIN 命令查看查询计划，确认索引是否被使用

