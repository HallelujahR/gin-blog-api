package service

import (
	"api/dao"
	"api/models"
	"math"
)

func CreateComment(c *models.Comment) error {
	return dao.CreateComment(c)
}

func GetCommentByID(id uint64) (*models.Comment, error) {
	return dao.GetCommentByID(id)
}

func ListCommentsByPost(postID uint64) ([]models.Comment, error) {
	return dao.ListCommentsByPost(postID)
}

func UpdateComment(c *models.Comment) error {
	return dao.UpdateComment(c)
}

func DeleteComment(id uint64) error {
	return dao.DeleteComment(id)
}

// 分页响应结构
type CommentListResponse struct {
	Comments   []models.Comment `json:"comments"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

// 评论列表带分页和筛选
func ListCommentsWithPagination(page, pageSize int, postID uint64, status, q, sort string) (*CommentListResponse, error) {
	// 参数验证和默认值
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100 // 限制最大每页数量
	}
	if sort == "" {
		sort = "DESC"
	}
	
	// 获取总数
	total, err := dao.CountComments(postID, status, q)
	if err != nil {
		return nil, err
	}
	
	// 获取评论列表
	comments, err := dao.ListCommentsWithParams(page, pageSize, postID, status, q, sort)
	if err != nil {
		return nil, err
	}
	
	// 计算总页数
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	
	return &CommentListResponse{
		Comments:   comments,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// 批量删除评论
func DeleteComments(ids []uint64) error {
	return dao.DeleteComments(ids)
}

// 批量更新评论状态
func UpdateCommentsStatus(ids []uint64, status string) error {
	return dao.UpdateCommentsStatus(ids, status)
}

// 获取评论详情（包含回复）
func GetCommentWithReplies(id uint64) (*models.Comment, []models.Comment, error) {
	comment, err := dao.GetCommentByID(id)
	if err != nil {
		return nil, nil, err
	}
	
	replies, _ := dao.GetCommentReplies(id)
	return comment, replies, nil
}
