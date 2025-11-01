package dao

import (
	"api/database"
	"api/models"
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
func ListPostsWithParams(page int, pageSize int, q, sort, category, tag string) ([]models.Post, error) {
	var posts []models.Post
	db := database.GetDB().Model(&models.Post{}).Where("status = 'published'")
	if q != "" {
		db = db.Where("title LIKE ? OR content LIKE ?", q+"%", q+"%")
	}
	if sort != "" {
		db = db.Order("created_at " + sort)
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
	err := db.Limit(pageSize).Offset((page - 1) * pageSize).Find(&posts).Error

	return posts, err
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
