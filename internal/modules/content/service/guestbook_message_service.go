package service

import (
	"api/internal/modules/content/dao"
	"api/internal/modules/content/models"
	"math"
)

func CreateGuestbookMessage(message *models.GuestbookMessage) error {
	return dao.CreateGuestbookMessage(message)
}

func GetGuestbookMessageByID(id uint64) (*models.GuestbookMessage, error) {
	return dao.GetGuestbookMessageByID(id)
}

func UpdateGuestbookMessage(message *models.GuestbookMessage) error {
	return dao.UpdateGuestbookMessage(message)
}

func DeleteGuestbookMessage(id uint64) error {
	return dao.DeleteGuestbookMessage(id)
}

type GuestbookMessageListResponse struct {
	Messages   []models.GuestbookMessage `json:"messages"`
	Total      int64                     `json:"total"`
	Page       int                       `json:"page"`
	PageSize   int                       `json:"page_size"`
	TotalPages int                       `json:"total_pages"`
}

func ListGuestbookMessagesWithPagination(page, pageSize int, status, q, sort string) (*GuestbookMessageListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	if sort == "" {
		sort = "DESC"
	}

	total, err := dao.CountGuestbookMessages(status, q)
	if err != nil {
		return nil, err
	}

	messages, err := dao.ListGuestbookMessagesWithParams(page, pageSize, status, q, sort)
	if err != nil {
		return nil, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(pageSize)))
	}

	return &GuestbookMessageListResponse{
		Messages:   messages,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
