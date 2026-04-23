package drawguess

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var manager = NewManager()

func CreateRoom(c *gin.Context) {
	var input CreateRoomInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	result, err := manager.CreateRoom(input.PlayerName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func JoinRoom(c *gin.Context) {
	var input JoinRoomInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	result, err := manager.JoinRoom(c.Param("roomId"), input.PlayerName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func GetRoom(c *gin.Context) {
	view, err := manager.GetRoom(c.Param("roomId"), c.Query("player_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"room": view})
}

func StreamRoom(c *gin.Context) {
	roomID := c.Param("roomId")
	playerID := c.Query("player_id")
	if strings.TrimSpace(playerID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 player_id"})
		return
	}

	ch, initial, err := manager.Subscribe(roomID, playerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "stream not supported"})
		return
	}

	writeSSE(c, initial)
	flusher.Flush()

	pingTicker := time.NewTicker(15 * time.Second)
	defer pingTicker.Stop()
	defer manager.Disconnect(roomID, playerID)

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case payload, ok := <-ch:
			if !ok {
				return
			}
			writeSSE(c, payload)
			flusher.Flush()
		case <-pingTicker.C:
			_, _ = c.Writer.Write([]byte(": ping\n\n"))
			flusher.Flush()
		}
	}
}

func StartGame(c *gin.Context) {
	var input StartGameInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	view, err := manager.StartGame(c.Param("roomId"), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"room": view})
}

func SubmitGuess(c *gin.Context) {
	var input GuessInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	view, err := manager.SubmitGuess(c.Param("roomId"), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"room": view})
}

func SubmitStroke(c *gin.Context) {
	var input StrokeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	if err := manager.AddStroke(c.Param("roomId"), input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ClearCanvas(c *gin.Context) {
	var input ClearInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	if err := manager.ClearCanvas(c.Param("roomId"), input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func LeaveRoom(c *gin.Context) {
	var input LeaveRoomInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	manager.RemovePlayer(c.Param("roomId"), input.PlayerID, false)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func writeSSE(c *gin.Context, payload []byte) {
	_, _ = c.Writer.Write([]byte("data: "))
	_, _ = c.Writer.Write(payload)
	_, _ = c.Writer.Write([]byte("\n\n"))
}
