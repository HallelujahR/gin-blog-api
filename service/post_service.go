package service

import (
	"api/dao"
	"api/database"
	"api/models"
	"fmt"
	"gorm.io/gorm"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 创建文章，并分配分类和标签
func CreatePost(post *models.Post, categoryIDs, tagIDs []uint64) error {
	err := dao.CreatePost(post)
	if err != nil {
		return err
	}
	// 性能优化：使用批量插入而不是循环插入
	if len(categoryIDs) > 0 {
		if err := dao.BatchAddCategoriesToPost(post.ID, categoryIDs); err != nil {
			return err
		}
	}
	if len(tagIDs) > 0 {
		if err := dao.BatchAddTagsToPost(post.ID, tagIDs); err != nil {
			return err
		}
	}
	return nil
}

// 查单篇文章（不带预加载）
func GetPostByID(id uint64) (*models.Post, error) {
	return dao.GetPostByID(id)
}

// 更新文章
func UpdatePost(post *models.Post) error {
	return dao.UpdatePost(post)
}

// 查询文章列表
func ListPosts() ([]models.Post, error) {
	return dao.ListPosts()
}

// 查询文章列表带参数
func ListPostsWithParams(page, pageSize int, q, sort, category, tag string, status string) ([]models.PostWithRelations, error) {
	posts, err := dao.ListPostsWithParams(page, pageSize, q, sort, category, tag, status)
	if err != nil {
		return nil, err
	}

	// 转换分类和标签数据
	for i := range posts {
		fmt.Println(posts[i].CategoryNamesStr)
		if posts[i].CategoryNamesStr != "" {
			posts[i].CategoryNames = strings.Split(posts[i].CategoryNamesStr, ",")
			posts[i].CategoryIDs = stringToUint64Slice(posts[i].CategoryIDsStr)
		}
		if posts[i].TagNamesStr != "" {
			posts[i].TagNames = strings.Split(posts[i].TagNamesStr, ",")
			posts[i].TagIDs = stringToUint64Slice(posts[i].TagIDsStr)
		}
	}

	return posts, nil
}

// 辅助函数：将逗号分隔的字符串转换为uint64切片
func stringToUint64Slice(s string) []uint64 {
	if s == "" {
		return nil
	}
	strSlice := strings.Split(s, ",")
	var ids []uint64
	for _, str := range strSlice {
		id, err := strconv.ParseUint(str, 10, 64)
		if err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

// 删除
func DeletePost(id uint64) error {
	// 先删除关联的分类和标签
	dao.DeletePostCategories(id)
	dao.DeletePostTags(id)
	return dao.DeletePost(id)
}

// 生成slug（如果未提供）
func GenerateSlug(title string) string {
	// 转小写
	slug := strings.ToLower(title)
	// 替换中文空格和英文空格为短横线
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "　", "-")
	// 移除特殊字符，只保留字母、数字、短横线
	reg := regexp.MustCompile(`[^a-z0-9-]+`)
	slug = reg.ReplaceAllString(slug, "-")
	// 移除连续的短横线
	reg = regexp.MustCompile(`-+`)
	slug = reg.ReplaceAllString(slug, "-")
	// 移除首尾短横线
	slug = strings.Trim(slug, "-")
	// 如果为空，使用时间戳
	if slug == "" {
		slug = fmt.Sprintf("post-%d", time.Now().Unix())
	}
	return slug
}

// 确保slug唯一（如果已存在则添加数字后缀）
func EnsureUniqueSlug(slug string, excludeID uint64) string {
	if !dao.PostSlugExists(slug, excludeID) {
		return slug
	}

	// 如果已存在，添加数字后缀
	for i := 1; i < 1000; i++ {
		newSlug := fmt.Sprintf("%s-%d", slug, i)
		if !dao.PostSlugExists(newSlug, excludeID) {
			return newSlug
		}
	}

	// 如果1000次都失败，使用时间戳
	return fmt.Sprintf("%s-%d", slug, time.Now().Unix())
}

// 更新文章的分类和标签
func UpdatePostCategoriesAndTags(postID uint64, categoryIDs, tagIDs []uint64) error {
	// 删除旧的关联
	dao.DeletePostCategories(postID)
	dao.DeletePostTags(postID)

	// 性能优化：使用批量插入而不是循环插入
	if len(categoryIDs) > 0 {
		if err := dao.BatchAddCategoriesToPost(postID, categoryIDs); err != nil {
			return err
		}
	}
	if len(tagIDs) > 0 {
		if err := dao.BatchAddTagsToPost(postID, tagIDs); err != nil {
			return err
		}
	}
	return nil
}

// 获取文章详情（包含分类和标签ID）
func GetPostWithRelations(id uint64) (*models.Post, []uint64, []uint64, error) {
	post, err := dao.GetPostByID(id)
	if err != nil {
		return nil, nil, nil, err
	}

	// 获取关联的分类和标签ID
	categoryIDs, _ := dao.GetPostCategoryIDs(id)
	tagIDs, _ := dao.GetPostTagIDs(id)

	return post, categoryIDs, tagIDs, nil
}

// 获取文章详情（包含完整的分类和标签信息）
func GetPostWithFullRelations(id uint64) (*models.Post, []models.Category, []models.Tag, error) {
	post, err := dao.GetPostByID(id)
	if err != nil {
		return nil, nil, nil, err
	}

	// 获取关联的分类和标签完整信息
	categories, err := dao.GetPostCategories(id)
	if err != nil {
		// 如果查询失败，返回空数组而不是nil
		categories = []models.Category{}
	}

	tags, err := dao.GetPostTags(id)
	if err != nil {
		// 如果查询失败，返回空数组而不是nil
		tags = []models.Tag{}
	}

	// 确保返回的不是nil
	if categories == nil {
		categories = []models.Category{}
	}
	if tags == nil {
		tags = []models.Tag{}
	}

	return post, categories, tags, nil
}

// RecordView 记录浏览流量（每次访问详情都增加一次浏览数）
func RecordView(postID uint64) {
	// 直接更新文章浏览数，不记录IP
	db := database.GetDB()
	db.Model(&models.Post{}).Where("id = ?", postID).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))
	// 记录每日访问统计（忽略错误，避免影响主流程）
	_ = dao.IncrementPostViewStat(postID, time.Now())
}

// 分页响应结构
type PostListResponse struct {
	Posts      []models.PostWithRelations `json:"posts"`
	Total      int64                      `json:"total"`
	Page       int                        `json:"page"`
	PageSize   int                        `json:"page_size"`
	TotalPages int                        `json:"total_pages"`
}

// 查询文章列表带分页
func ListPostsWithPagination(page, pageSize int, q, sort, category, tag, status string) (*PostListResponse, error) {
	// 参数验证和默认值
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100 // 限制最大每页数量
	}

	// 获取总数
	total, err := dao.CountPosts(q, sort, category, tag, status)
	if err != nil {
		return nil, err
	}

	// 获取文章列表
	posts, err := dao.ListPostsWithParams(page, pageSize, q, sort, category, tag, status)
	if err != nil {
		return nil, err
	}

	// 转换分类和标签数据
	for i := range posts {
		if posts[i].CategoryNamesStr != "" {
			posts[i].CategoryNames = strings.Split(posts[i].CategoryNamesStr, ",")
			posts[i].CategoryIDs = stringToUint64Slice(posts[i].CategoryIDsStr)
		}
		if posts[i].TagNamesStr != "" {
			posts[i].TagNames = strings.Split(posts[i].TagNamesStr, ",")
			posts[i].TagIDs = stringToUint64Slice(posts[i].TagIDsStr)
		}
	}

	// 计算总页数
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &PostListResponse{
		Posts:      posts,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
