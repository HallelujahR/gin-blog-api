package drawguess

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var manager = NewManager()

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

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

func SocketRoom(c *gin.Context) {
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

	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	defer manager.Disconnect(roomID, playerID)

	_ = conn.SetReadDeadline(time.Now().Add(65 * time.Second))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(65 * time.Second))
	})

	if err := conn.WriteMessage(websocket.TextMessage, initial); err != nil {
		return
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, payload, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if err := handleSocketMessage(roomID, playerID, payload); err != nil {
				_ = writeSocketJSON(conn, eventEnvelope{
					Type: "error",
					Data: gin.H{"message": err.Error()},
				})
			}
		}
	}()

	pingTicker := time.NewTicker(20 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case <-done:
			return
		case payload, ok := <-ch:
			if !ok {
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				return
			}
		case <-pingTicker.C:
			if err := conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
				return
			}
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

type socketInboundEnvelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func handleSocketMessage(roomID, playerID string, payload []byte) error {
	var envelope socketInboundEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return err
	}

	switch envelope.Type {
	case "guess":
		var input GuessInput
		if err := json.Unmarshal(envelope.Data, &input); err != nil {
			return err
		}
		input.PlayerID = playerID
		_, err := manager.SubmitGuess(roomID, input)
		return err
	case "stroke":
		var input StrokeInput
		if err := json.Unmarshal(envelope.Data, &input); err != nil {
			return err
		}
		input.PlayerID = playerID
		return manager.AddStroke(roomID, input)
	case "start":
		var input StartGameInput
		if len(envelope.Data) > 0 {
			if err := json.Unmarshal(envelope.Data, &input); err != nil {
				return err
			}
		}
		input.PlayerID = playerID
		_, err := manager.StartGame(roomID, input)
		return err
	case "clear":
		input := ClearInput{PlayerID: playerID}
		if len(envelope.Data) > 0 {
			if err := json.Unmarshal(envelope.Data, &input); err != nil {
				return err
			}
			input.PlayerID = playerID
		}
		return manager.ClearCanvas(roomID, input)
	case "leave":
		manager.RemovePlayer(roomID, playerID, false)
		return nil
	default:
		return nil
	}
}

func writeSocketJSON(conn *websocket.Conn, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, data)
}
