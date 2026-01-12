package stats

import (
	"api/dao"
	"fmt"
)

const (
	defaultLookbackDays = 30
	defaultTopPosts     = 3
)

// BuildVisitSummary 汇总数据库中的访问量与文章热度信息。
// 如果days <= 0，则统计所有历史数据。
func BuildVisitSummary(days, topN int) (VisitSummary, error) {
	if topN <= 0 {
		topN = defaultTopPosts
	}

	total, err := dao.SumVisitsSince(days)
	if err != nil {
		return VisitSummary{}, err
	}
	ranks, err := dao.TopPostsByViewsSince(days, topN)
	if err != nil {
		return VisitSummary{}, err
	}

	// 若查询结果无数据，回退到文章表的累计数据
	if total == 0 {
		if total, err = dao.SumAllPostViews(); err != nil {
			return VisitSummary{}, err
		}
	}
	if len(ranks) == 0 {
		if ranks, err = dao.TopPostsByTotalViews(topN); err != nil {
			return VisitSummary{}, err
		}
	}

	topPosts := make([]TopPost, 0, len(ranks))
	for _, rank := range ranks {
		topPosts = append(topPosts, TopPost{
			PostID: rank.PostID,
			Title:  rank.Title,
			Path:   fmt.Sprintf("/posts/%d", rank.PostID),
			Count:  int(rank.Views),
		})
	}

	return VisitSummary{
		TotalVisits: int(total),
		TopPosts:    topPosts,
	}, nil
}
