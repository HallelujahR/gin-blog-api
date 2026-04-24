package routes

import (
	"api/internal/middleware"
	"api/internal/modules/drawguess"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterDrawGuessRoutes(r *gin.Engine) {
	group := r.Group("/api/tools/draw-guess")
	{
		group.POST("/rooms", middleware.RateLimitMiddleware(30, time.Minute), drawguess.CreateRoom)
		group.POST("/rooms/:roomId/join", middleware.RateLimitMiddleware(60, time.Minute), drawguess.JoinRoom)
		group.GET("/rooms/:roomId", middleware.RateLimitMiddleware(240, time.Minute), drawguess.GetRoom)
		group.GET("/rooms/:roomId/ws", middleware.RateLimitMiddleware(60, time.Minute), drawguess.SocketRoom)
		group.POST("/rooms/:roomId/start", middleware.RateLimitMiddleware(30, time.Minute), drawguess.StartGame)
		group.POST("/rooms/:roomId/guess", middleware.RateLimitMiddleware(240, time.Minute), drawguess.SubmitGuess)
		group.POST("/rooms/:roomId/strokes", middleware.RateLimitMiddleware(6000, time.Minute), drawguess.SubmitStroke)
		group.POST("/rooms/:roomId/clear", middleware.RateLimitMiddleware(120, time.Minute), drawguess.ClearCanvas)
		group.POST("/rooms/:roomId/leave", middleware.RateLimitMiddleware(120, time.Minute), drawguess.LeaveRoom)
	}
}
