package stats

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	defaultCacheTTL = 24 * time.Hour // 统计结果缓存时间
)

// Handler 返回博客访问统计信息。
// 读取最近30天访问日志，当日志不存在时返回空统计结果。
// 结果会缓存一段时间，减少频繁的磁盘读取与解析。
func Handler(c *gin.Context) {
	ensureScheduler()

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if cached, ok := GetStatsCache(ctx); ok {
		c.JSON(http.StatusOK, cached)
		return
	}

	entries, err := LoadRecentEntries(ctx, 30)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := Aggregate(entries, 3)
	result.GeneratedAt = time.Now().UTC()

	if err := SetStatsCache(ctx, result, defaultCacheTTL); err != nil {
		// 缓存失败不影响主流程
	}

	c.JSON(http.StatusOK, result)
}
