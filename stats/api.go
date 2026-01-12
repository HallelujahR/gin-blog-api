package stats

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	defaultCacheTTL = time.Hour // 统计结果缓存时间
)

// Handler 返回博客访问统计信息。
// 优先命中 Redis 缓存；未命中时实时汇总并写回缓存。
func Handler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if cached, ok := GetStatsCache(ctx); ok {
		c.JSON(http.StatusOK, cached)
		return
	}

	// 使用0表示查询所有历史数据
	snapshot, err := LoadTrafficSnapshot(0)
	if err != nil {
		fmt.Printf("[stats] load snapshot error: %v\n", err)
		snapshot = SnapshotData{}
	}

	// 使用0表示查询所有历史数据
	summary, err := BuildVisitSummary(0, defaultTopPosts)
	if err != nil {
		fmt.Printf("[stats] build summary error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := Aggregate(snapshot, summary)
	result.GeneratedAt = time.Now().UTC()

	if err := SetStatsCache(ctx, result, defaultCacheTTL); err != nil {
		fmt.Printf("[stats] set cache error: %v\n", err)
	}

	c.JSON(http.StatusOK, result)
}
