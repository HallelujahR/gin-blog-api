package dao

import (
	"api/database"
	"api/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PostViewRank 用于返回文章热度排名。
type PostViewRank struct {
	PostID uint64
	Title  string
	Views  int64
}

// IncrementPostViewStat 将指定文章在某天的访问次数 +1。
func IncrementPostViewStat(postID uint64, viewTime time.Time) error {
	date := normalizeDate(viewTime)
	stat := models.PostViewStat{
		PostID: postID,
		Date:   date,
		Views:  1,
	}
	return database.GetDB().
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "post_id"}, {Name: "date"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"views": gorm.Expr("post_view_stats.views + VALUES(views)")}),
		}).
		Create(&stat).Error
}

// SumVisitsSince 统计最近days天的总访问次数。
func SumVisitsSince(days int) (int64, error) {
	if days <= 0 {
		days = 30
	}
	start := normalizeDate(time.Now().AddDate(0, 0, -(days - 1)))
	var total int64
	err := database.GetDB().
		Model(&models.PostViewStat{}).
		Where("date >= ?", start).
		Select("COALESCE(SUM(views), 0)").
		Scan(&total).Error
	return total, err
}

// TopPostsByViewsSince 获取最近days天内访问量最高的文章。
func TopPostsByViewsSince(days, limit int) ([]PostViewRank, error) {
	if limit <= 0 {
		limit = 3
	}
	if days <= 0 {
		days = 30
	}
	start := normalizeDate(time.Now().AddDate(0, 0, -(days - 1)))
	var ranks []PostViewRank
	err := database.GetDB().
		Table("post_view_stats pvs").
		Select("pvs.post_id as post_id, posts.title as title, COALESCE(SUM(pvs.views),0) as views").
		Joins("JOIN posts ON posts.id = pvs.post_id").
		Where("pvs.date >= ?", start).
		Group("pvs.post_id, posts.title").
		Order("views DESC").
		Limit(limit).
		Scan(&ranks).Error
	return ranks, err
}

// SumAllPostViews 统计文章表中的总阅读量（作为回退）。
func SumAllPostViews() (int64, error) {
	var total int64
	err := database.GetDB().
		Model(&models.Post{}).
		Select("COALESCE(SUM(view_count), 0)").
		Scan(&total).Error
	return total, err
}

// TopPostsByTotalViews 使用文章表中的view_count获取热度排行（作为回退）。
func TopPostsByTotalViews(limit int) ([]PostViewRank, error) {
	if limit <= 0 {
		limit = 3
	}
	var ranks []PostViewRank
	err := database.GetDB().
		Model(&models.Post{}).
		Select("id as post_id, title, view_count as views").
		Order("view_count DESC").
		Limit(limit).
		Scan(&ranks).Error
	return ranks, err
}

func normalizeDate(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
