package utils

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// GetClientIP 获取客户端真实IP
func GetClientIP(c *gin.Context) string {
	// 优先从X-Forwarded-For获取（经过代理时）
	ip := c.GetHeader("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For可能包含多个IP，取第一个
		ips := strings.Split(ip, ",")
		if len(ips) > 0 {
			ip = strings.TrimSpace(ips[0])
		}
	}
	// 其次从X-Real-IP获取
	if ip == "" {
		ip = c.GetHeader("X-Real-IP")
	}
	// 最后从RemoteAddr获取
	if ip == "" {
		ip = strings.Split(c.ClientIP(), ":")[0]
	}
	return ip
}

