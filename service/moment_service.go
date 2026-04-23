package service

import (
	"api/dao"
	"api/models"
	"time"
)

func CreateMoment(moment *models.Moment) error {
	if moment.Status == "published" && moment.PublishedAt == nil {
		now := time.Now()
		moment.PublishedAt = &now
	}
	return dao.CreateMoment(moment)
}

func GetMomentByID(id uint64) (*models.Moment, error) {
	return dao.GetMomentByID(id)
}

func UpdateMoment(moment *models.Moment) error {
	if moment.Status == "published" && moment.PublishedAt == nil {
		now := time.Now()
		moment.PublishedAt = &now
	}
	if moment.Status != "published" {
		moment.PublishedAt = nil
	}
	return dao.UpdateMoment(moment)
}

func DeleteMoment(id uint64) error {
	return dao.DeleteMoment(id)
}

func ListPublishedMoments(limit int) ([]models.Moment, error) {
	return dao.ListMoments("published", limit)
}

func ListAllMoments(limit int) ([]models.Moment, error) {
	return dao.ListMoments("", limit)
}
