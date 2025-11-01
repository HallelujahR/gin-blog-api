package service

import (
	"api/dao"
	"api/models"
)

// 创建文章，并分配分类和标签
func CreatePost(post *models.Post, categoryIDs, tagIDs []uint64) error {
	err := dao.CreatePost(post)
	if err != nil {
		return err
	}
	for _, cid := range categoryIDs {
		dao.AddCategoryToPost(post.ID, cid)
	}
	for _, tid := range tagIDs {
		dao.AddTagToPost(post.ID, tid)
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
func ListPostsWithParams(page, pageSize int, q, sort, category, tag string) ([]models.Post, error) {
	return dao.ListPostsWithParams(page, pageSize, q, sort, category, tag)
}

// 删除
func DeletePost(id uint64) error {
	return dao.DeletePost(id)
}
