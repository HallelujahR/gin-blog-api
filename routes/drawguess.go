package routes

import (
	"api/middleware"
	"api/modules/drawguess"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterDrawGuessRoutes(r *gin.Engine) {
	group := r.Group("/api/tools/draw-guess")
	group.Use(middleware.RateLimitMiddleware(240, time.Minute))
	{
		group.POST("/rooms", drawguess.CreateRoom)
		group.POST("/rooms/:roomId/join", drawguess.JoinRoom)
		group.GET("/rooms/:roomId", drawguess.GetRoom)
		group.GET("/rooms/:roomId/stream", drawguess.StreamRoom)
		group.POST("/rooms/:roomId/start", drawguess.StartGame)
		group.POST("/rooms/:roomId/guess", drawguess.SubmitGuess)
		group.POST("/rooms/:roomId/strokes", drawguess.SubmitStroke)
		group.POST("/rooms/:roomId/clear", drawguess.ClearCanvas)
		group.POST("/rooms/:roomId/leave", drawguess.LeaveRoom)
	}
}
