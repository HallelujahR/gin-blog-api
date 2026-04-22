package dao

import (
	"api/database"
	"api/models"
	"strings"

	"gorm.io/gorm"
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
	postIDs, err := listOrderedPostIDs(page, pageSize, q, sort, category, tag, status)
	if err != nil {
		return nil, err
	}

	if len(postIDs) == 0 {
		return []models.PostWithRelations{}, nil
	}

	var posts []models.Post
	if err := database.GetDB().Where("id IN ?", postIDs).Find(&posts).Error; err != nil {
		return nil, err
	}

	postMap := make(map[uint64]models.Post, len(posts))
	for _, post := range posts {
		postMap[post.ID] = post
	}

	orderedPosts := make([]models.Post, 0, len(postIDs))
	for _, id := range postIDs {
		post, ok := postMap[id]
		if !ok {
			continue
		}
		orderedPosts = append(orderedPosts, post)
	}

	return hydratePostRelations(orderedPosts)
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

// 批量给文章添加分类（性能优化：使用批量插入）
func BatchAddCategoriesToPost(postID uint64, categoryIDs []uint64) error {
	if len(categoryIDs) == 0 {
		return nil
	}
	rels := make([]models.PostCategory, 0, len(categoryIDs))
	for _, cid := range categoryIDs {
		rels = append(rels, models.PostCategory{PostID: postID, CategoryID: cid})
	}
	return database.GetDB().Create(&rels).Error
}

// 给文章添加标签
func AddTagToPost(postID, tagID uint64) error {
	rel := models.PostTag{PostID: postID, TagID: tagID}
	return database.GetDB().Create(&rel).Error
}

// 批量给文章添加标签（性能优化：使用批量插入）
func BatchAddTagsToPost(postID uint64, tagIDs []uint64) error {
	if len(tagIDs) == 0 {
		return nil
	}
	rels := make([]models.PostTag, 0, len(tagIDs))
	for _, tid := range tagIDs {
		rels = append(rels, models.PostTag{PostID: postID, TagID: tid})
	}
	return database.GetDB().Create(&rels).Error
}

// 统计文章总数（用于分页）
func CountPosts(q, sort, category, tag, status string) (int64, error) {
	var count int64
	db := buildPostFilterQuery(q, sort, category, tag, status)
	err := db.Distinct("posts.id").Count(&count).Error
	return count, err
}

func listOrderedPostIDs(page int, pageSize int, q, sort, category, tag, status string) ([]uint64, error) {
	type postIDRow struct {
		ID uint64
	}

	var rows []postIDRow
	err := buildPostFilterQuery(q, sort, category, tag, status).
		Select("DISTINCT posts.id, posts.published_at, posts.created_at").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	ids := make([]uint64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids, nil
}

func buildPostFilterQuery(q, sort, category, tag, status string) *gorm.DB {
	db := database.GetDB().Model(&models.Post{})

	if status != "" {
		db = db.Where("posts.status = ?", status)
	}

	q = strings.TrimSpace(q)
	if q != "" {
		likePattern := "%" + q + "%"
		db = db.Where("posts.title LIKE ? OR posts.excerpt LIKE ?", likePattern, likePattern)
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

	orderDirection := sanitizeSortOrder(sort)
	return db.Order("posts.published_at IS NULL ASC").
		Order("posts.published_at " + orderDirection).
		Order("posts.created_at " + orderDirection)
}

func hydratePostRelations(posts []models.Post) ([]models.PostWithRelations, error) {
	if len(posts) == 0 {
		return []models.PostWithRelations{}, nil
	}

	postIDs := make([]uint64, 0, len(posts))
	for _, post := range posts {
		postIDs = append(postIDs, post.ID)
	}

	categoryMap, err := getPostCategoryRelations(postIDs)
	if err != nil {
		return nil, err
	}
	tagMap, err := getPostTagRelations(postIDs)
	if err != nil {
		return nil, err
	}

	result := make([]models.PostWithRelations, 0, len(posts))
	for _, post := range posts {
		item := models.PostWithRelations{Post: post}
		if rel := categoryMap[post.ID]; rel != nil {
			item.CategoryNames = rel.names
			item.CategoryIDs = rel.ids
		}
		if rel := tagMap[post.ID]; rel != nil {
			item.TagNames = rel.names
			item.TagIDs = rel.ids
		}
		result = append(result, item)
	}
	return result, nil
}

type relationValues struct {
	names []string
	ids   []uint64
}

func getPostCategoryRelations(postIDs []uint64) (map[uint64]*relationValues, error) {
	type row struct {
		PostID       uint64
		CategoryID   uint64
		CategoryName string
	}
	var rows []row
	err := database.GetDB().Table("post_categories").
		Select("post_categories.post_id, categories.id AS category_id, categories.name AS category_name").
		Joins("JOIN categories ON categories.id = post_categories.category_id").
		Where("post_categories.post_id IN ?", postIDs).
		Order("post_categories.post_id ASC, categories.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uint64]*relationValues, len(postIDs))
	for _, row := range rows {
		if _, ok := result[row.PostID]; !ok {
			result[row.PostID] = &relationValues{}
		}
		result[row.PostID].ids = append(result[row.PostID].ids, row.CategoryID)
		result[row.PostID].names = append(result[row.PostID].names, row.CategoryName)
	}
	return result, nil
}

func getPostTagRelations(postIDs []uint64) (map[uint64]*relationValues, error) {
	type row struct {
		PostID  uint64
		TagID   uint64
		TagName string
	}
	var rows []row
	err := database.GetDB().Table("post_tags").
		Select("post_tags.post_id, tags.id AS tag_id, tags.name AS tag_name").
		Joins("JOIN tags ON tags.id = post_tags.tag_id").
		Where("post_tags.post_id IN ?", postIDs).
		Order("post_tags.post_id ASC, tags.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uint64]*relationValues, len(postIDs))
	for _, row := range rows {
		if _, ok := result[row.PostID]; !ok {
			result[row.PostID] = &relationValues{}
		}
		result[row.PostID].ids = append(result[row.PostID].ids, row.TagID)
		result[row.PostID].names = append(result[row.PostID].names, row.TagName)
	}
	return result, nil
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
