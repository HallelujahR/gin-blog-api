package dao

import (
	"api/database"
	"api/models"
	"strings"
)

// 新增文章
func CreatePost(post *models.Post) error {
	return database.GetDB().Create(post).Error
}

// 通过ID查找
func GetPostByID(id uint64) (*models.Post, error) {
	var post models.Post
	err := database.GetDB().First(&post, id).Error
	return &post, err
}

// 文章列表
func ListPosts() ([]models.Post, error) {
	var posts []models.Post
	err := database.GetDB().Find(&posts).Error
	return posts, err
}

// 文章列表带参数

func ListPostsWithParams(page int, pageSize int, q, sort, category, tag, status string) ([]models.PostWithRelations, error) {

	var posts []models.PostWithRelations
	db := database.GetDB().Model(&models.Post{}).
		Select(`posts.*, 
			GROUP_CONCAT(DISTINCT categories.name) as category_names_sql,
			GROUP_CONCAT(DISTINCT categories.id) as category_ids_sql,
			GROUP_CONCAT(DISTINCT tags.name) as tag_names_sql,
			GROUP_CONCAT(DISTINCT tags.id) as tag_ids_sql`).
		Joins("LEFT JOIN post_categories ON post_categories.post_id = posts.id").
		Joins("LEFT JOIN categories ON categories.id = post_categories.category_id").
		Joins("LEFT JOIN post_tags ON post_tags.post_id = posts.id").
		Joins("LEFT JOIN tags ON tags.id = post_tags.tag_id").
		Group("posts.id")

	if status != "" {
		db = db.Where("posts.status = ?", status)
	}

	if q != "" {
		db = db.Where("posts.title LIKE ? OR posts.content LIKE ?", q+"%", q+"%")
	}
	orderDirection := sanitizeSortOrder(sort)
	db = db.Order("posts.created_at " + orderDirection)
	if category != "" {
		db = db.Where("categories.slug = ?", category)
	}
	if tag != "" {
		db = db.Where("tags.slug = ?", tag)
	}
	err := db.Limit(pageSize).Offset((page - 1) * pageSize).Find(&posts).Error

	return posts, err
}

func sanitizeSortOrder(input string) string {
	dir := strings.ToUpper(strings.TrimSpace(input))
	if dir != "ASC" && dir != "DESC" {
		return "DESC"
	}
	return dir
}

// 更新
func UpdatePost(post *models.Post) error {
	return database.GetDB().Save(post).Error
}

// 删除
func DeletePost(id uint64) error {
	return database.GetDB().Delete(&models.Post{}, id).Error
}

// 给文章添加分类
func AddCategoryToPost(postID, categoryID uint64) error {
	rel := models.PostCategory{PostID: postID, CategoryID: categoryID}
	return database.GetDB().Create(&rel).Error
}

// 给文章添加标签
func AddTagToPost(postID, tagID uint64) error {
	rel := models.PostTag{PostID: postID, TagID: tagID}
	return database.GetDB().Create(&rel).Error
}

// 统计文章总数（用于分页）
func CountPosts(q, sort, category, tag, status string) (int64, error) {
	var count int64
	db := database.GetDB().Model(&models.Post{})

	if status != "" {
		db = db.Where("status = ?", status)
	}

	if q != "" {
		db = db.Where("title LIKE ? OR content LIKE ?", "%"+q+"%", "%"+q+"%")
	}

	if category != "" {
		db = db.Joins("JOIN post_categories ON post_categories.post_id = posts.id").
			Joins("JOIN categories ON categories.id = post_categories.category_id").
			Where("categories.slug = ?", category)
	}

	if tag != "" {
		db = db.Joins("JOIN post_tags ON post_tags.post_id = posts.id").
			Joins("JOIN tags ON tags.id = post_tags.tag_id").
			Where("tags.slug = ?", tag)
	}

	err := db.Count(&count).Error
	return count, err
}

// 删除文章的所有分类关联
func DeletePostCategories(postID uint64) error {
	return database.GetDB().Where("post_id = ?", postID).Delete(&models.PostCategory{}).Error
}

// 删除文章的所有标签关联
func DeletePostTags(postID uint64) error {
	return database.GetDB().Where("post_id = ?", postID).Delete(&models.PostTag{}).Error
}

// 检查slug是否已存在
func PostSlugExists(slug string, excludeID uint64) bool {
	var count int64
	db := database.GetDB().Model(&models.Post{}).Where("slug = ?", slug)
	if excludeID > 0 {
		db = db.Where("id != ?", excludeID)
	}
	db.Count(&count)
	return count > 0
}

// 通过slug查找文章
func GetPostBySlug(slug string) (*models.Post, error) {
	var post models.Post
	err := database.GetDB().Where("slug = ?", slug).First(&post).Error
	return &post, err
}

// 获取文章关联的分类ID列表
func GetPostCategoryIDs(postID uint64) ([]uint64, error) {
	var categoryIDs []uint64
	err := database.GetDB().Model(&models.PostCategory{}).
		Where("post_id = ?", postID).
		Pluck("category_id", &categoryIDs).Error
	return categoryIDs, err
}

// 获取文章关联的标签ID列表
func GetPostTagIDs(postID uint64) ([]uint64, error) {
	var tagIDs []uint64
	err := database.GetDB().Model(&models.PostTag{}).
		Where("post_id = ?", postID).
		Pluck("tag_id", &tagIDs).Error
	return tagIDs, err
}

// 获取文章关联的分类详细信息列表
func GetPostCategories(postID uint64) ([]models.Category, error) {
	var categories []models.Category
	err := database.GetDB().
		Model(&models.Category{}).
		Joins("INNER JOIN post_categories ON post_categories.category_id = categories.id").
		Where("post_categories.post_id = ?", postID).
		Find(&categories).Error

	// GORM的Find不会返回nil，但为了安全起见确保不是nil
	if categories == nil {
		categories = []models.Category{}
	}
	return categories, err
}

// 获取文章关联的标签详细信息列表
func GetPostTags(postID uint64) ([]models.Tag, error) {
	var tags []models.Tag
	err := database.GetDB().
		Model(&models.Tag{}).
		Joins("INNER JOIN post_tags ON post_tags.tag_id = tags.id").
		Where("post_tags.post_id = ?", postID).
		Find(&tags).Error

	// GORM的Find不会返回nil，但为了安全起见确保不是nil
	if tags == nil {
		tags = []models.Tag{}
	}
	return tags, err
}

// 批量获取分类信息（根据ID列表）
func GetCategoriesByIDs(ids []uint64) ([]models.Category, error) {
	if len(ids) == 0 {
		return []models.Category{}, nil
	}
	var categories []models.Category
	err := database.GetDB().Where("id IN ?", ids).Find(&categories).Error
	return categories, err
}

// 批量获取标签信息（根据ID列表）
func GetTagsByIDs(ids []uint64) ([]models.Tag, error) {
	if len(ids) == 0 {
		return []models.Tag{}, nil
	}
	var tags []models.Tag
	err := database.GetDB().Where("id IN ?", ids).Find(&tags).Error
	return tags, err
}
