-- 移除likes表的外键约束
-- 因为使用虚拟user_id（基于IP生成），不需要关联真实用户表

-- 查看当前外键约束名称
-- SELECT CONSTRAINT_NAME FROM information_schema.TABLE_CONSTRAINTS 
-- WHERE TABLE_SCHEMA = 'blog' AND TABLE_NAME = 'likes' AND CONSTRAINT_TYPE = 'FOREIGN KEY';

-- 移除外键约束（根据实际约束名称调整）
ALTER TABLE `likes` DROP FOREIGN KEY `likes_ibfk_1`;

-- 如果需要移除其他外键约束，继续执行：
-- ALTER TABLE `likes` DROP FOREIGN KEY `likes_ibfk_2`;
-- ALTER TABLE `likes` DROP FOREIGN KEY `likes_ibfk_3`;


