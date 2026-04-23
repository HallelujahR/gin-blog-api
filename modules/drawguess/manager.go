package drawguess

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	maxRooms              = 60
	maxPlayersPerRoom     = 5
	maxPlayerNameLen      = 16
	maxMessagesPerRoom    = 40
	maxCanvasActions      = 320
	maxStrokePoints       = 24
	roundDuration         = 80 * time.Second
	roundCooldown         = 3 * time.Second
	roomIdleTTL           = 20 * time.Minute
	disconnectGracePeriod = 20 * time.Second
	minStrokeInterval     = 35 * time.Millisecond
	eventBufferSize       = 48
	defaultBrushWidth     = 4
)

var wordLibrary = []string{
	"咖啡", "键盘", "猫咪", "火箭", "海浪", "钢琴", "雨伞", "蜗牛", "相机", "西瓜",
	"耳机", "书包", "月亮", "雪人", "牙刷", "香蕉", "奶茶", "灯塔", "自行车", "向日葵",
	"拖鞋", "星星", "饺子", "狐狸", "烤鱼", "风筝", "汉堡", "绿茶", "仙人掌", "羽毛球",
}

type Manager struct {
	mu    sync.Mutex
	rooms map[string]*Room
}

type Room struct {
	ID                string
	Status            string
	HostPlayerID      string
	Round             int
	CurrentDrawerID   string
	CurrentWord       string
	RoundEndsAt       time.Time
	RoundToken        int
	LastActiveAt      time.Time
	CreatedAt         time.Time
	Messages          []Message
	CanvasActions     []CanvasAction
	Players           map[string]*Player
	PlayerOrder       []string
	Subscribers       map[string]chan []byte
	LastSystemMessage string
}

type Player struct {
	ID             string
	Name           string
	Score          int
	IsConnected    bool
	LastSeenAt     time.Time
	LastStrokeAt   time.Time
	PendingRemoval int64
}

type Message struct {
	ID        string    `json:"id"`
	PlayerID  string    `json:"player_id,omitempty"`
	Player    string    `json:"player_name,omitempty"`
	Content   string    `json:"content"`
	Kind      string    `json:"kind"`
	CreatedAt time.Time `json:"created_at"`
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type CanvasAction struct {
	ID        string    `json:"id"`
	PlayerID  string    `json:"player_id"`
	Player    string    `json:"player_name"`
	Kind      string    `json:"kind"`
	Color     string    `json:"color,omitempty"`
	Width     int       `json:"width,omitempty"`
	Tool      string    `json:"tool,omitempty"`
	Points    []Point   `json:"points,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type PlayerView struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Score       int    `json:"score"`
	IsHost      bool   `json:"is_host"`
	IsDrawer    bool   `json:"is_drawer"`
	IsConnected bool   `json:"is_connected"`
}

type RoomView struct {
	ID                string         `json:"id"`
	Status            string         `json:"status"`
	HostPlayerID      string         `json:"host_player_id"`
	CurrentDrawerID   string         `json:"current_drawer_id"`
	CurrentDrawerName string         `json:"current_drawer_name"`
	Round             int            `json:"round"`
	WordDisplay       string         `json:"word_display"`
	RemainingSeconds  int            `json:"remaining_seconds"`
	MaxPlayers        int            `json:"max_players"`
	Players           []PlayerView   `json:"players"`
	Messages          []Message      `json:"messages"`
	CanvasActions     []CanvasAction `json:"canvas_actions"`
	LastSystemMessage string         `json:"last_system_message"`
}

type RoomJoinResult struct {
	RoomID   string   `json:"room_id"`
	PlayerID string   `json:"player_id"`
	Room     RoomView `json:"room"`
}

type CreateRoomInput struct {
	PlayerName string `json:"player_name"`
}

type JoinRoomInput struct {
	PlayerName string `json:"player_name"`
}

type GuessInput struct {
	PlayerID string `json:"player_id"`
	Content  string `json:"content"`
}

type StartGameInput struct {
	PlayerID string `json:"player_id"`
}

type LeaveRoomInput struct {
	PlayerID string `json:"player_id"`
}

type StrokeInput struct {
	PlayerID string  `json:"player_id"`
	Color    string  `json:"color"`
	Width    int     `json:"width"`
	Tool     string  `json:"tool"`
	Points   []Point `json:"points"`
}

type ClearInput struct {
	PlayerID string `json:"player_id"`
}

func NewManager() *Manager {
	m := &Manager{
		rooms: make(map[string]*Room),
	}
	go m.cleanupLoop()
	return m
}

func (m *Manager) CreateRoom(playerName string) (RoomJoinResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.rooms) >= maxRooms {
		return RoomJoinResult{}, fmt.Errorf("当前房间数量已达上限，请稍后再试")
	}

	name, err := normalizePlayerName(playerName)
	if err != nil {
		return RoomJoinResult{}, err
	}

	roomID := m.generateRoomIDLocked()
	playerID := newID("player")
	now := time.Now()

	room := &Room{
		ID:            roomID,
		Status:        "waiting",
		HostPlayerID:  playerID,
		LastActiveAt:  now,
		CreatedAt:     now,
		Players:       make(map[string]*Player),
		PlayerOrder:   []string{playerID},
		Subscribers:   make(map[string]chan []byte),
		Messages:      []Message{},
		CanvasActions: []CanvasAction{},
	}
	room.Players[playerID] = &Player{
		ID:          playerID,
		Name:        name,
		IsConnected: true,
		LastSeenAt:  now,
	}
	room.LastSystemMessage = fmt.Sprintf("%s 创建了房间", name)
	room.Messages = append(room.Messages, Message{
		ID:        newID("msg"),
		PlayerID:  playerID,
		Player:    name,
		Content:   room.LastSystemMessage,
		Kind:      "system",
		CreatedAt: now,
	})

	m.rooms[roomID] = room
	return RoomJoinResult{RoomID: roomID, PlayerID: playerID, Room: room.snapshotFor(playerID)}, nil
}

func (m *Manager) JoinRoom(roomID, playerName string) (RoomJoinResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, ok := m.rooms[strings.ToUpper(strings.TrimSpace(roomID))]
	if !ok {
		return RoomJoinResult{}, fmt.Errorf("房间不存在")
	}
	if len(room.Players) >= maxPlayersPerRoom {
		return RoomJoinResult{}, fmt.Errorf("房间已满，最多支持 %d 人", maxPlayersPerRoom)
	}

	name, err := normalizePlayerName(playerName)
	if err != nil {
		return RoomJoinResult{}, err
	}
	for _, p := range room.Players {
		if strings.EqualFold(p.Name, name) {
			return RoomJoinResult{}, fmt.Errorf("房间内已存在同名玩家")
		}
	}

	now := time.Now()
	playerID := newID("player")
	room.Players[playerID] = &Player{
		ID:          playerID,
		Name:        name,
		IsConnected: true,
		LastSeenAt:  now,
	}
	room.PlayerOrder = append(room.PlayerOrder, playerID)
	room.LastActiveAt = now
	room.LastSystemMessage = fmt.Sprintf("%s 加入了房间", name)
	room.appendMessageLocked(Message{
		ID:        newID("msg"),
		PlayerID:  playerID,
		Player:    name,
		Content:   room.LastSystemMessage,
		Kind:      "system",
		CreatedAt: now,
	})
	snapshot := room.snapshotFor(playerID)
	dispatches := m.collectSnapshotsLocked(room)
	go sendDispatches(dispatches)
	return RoomJoinResult{RoomID: room.ID, PlayerID: playerID, Room: snapshot}, nil
}

func (m *Manager) GetRoom(roomID, playerID string) (RoomView, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, player, err := m.roomAndPlayerLocked(roomID, playerID)
	if err != nil {
		return RoomView{}, err
	}
	player.LastSeenAt = time.Now()
	return room.snapshotFor(player.ID), nil
}

func (m *Manager) Subscribe(roomID, playerID string) (<-chan []byte, []byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, player, err := m.roomAndPlayerLocked(roomID, playerID)
	if err != nil {
		return nil, nil, err
	}

	if old, ok := room.Subscribers[playerID]; ok {
		close(old)
	}
	ch := make(chan []byte, eventBufferSize)
	room.Subscribers[playerID] = ch
	player.IsConnected = true
	player.LastSeenAt = time.Now()

	initial, _ := json.Marshal(eventEnvelope{
		Type: "snapshot",
		Data: room.snapshotFor(playerID),
	})
	return ch, initial, nil
}

func (m *Manager) Disconnect(roomID, playerID string) {
	m.RemovePlayer(roomID, playerID, true)
}

func (m *Manager) RemovePlayer(roomID, playerID string, silent bool) {
	m.mu.Lock()
	room, ok := m.rooms[strings.ToUpper(strings.TrimSpace(roomID))]
	if !ok {
		m.mu.Unlock()
		return
	}
	player, ok := room.Players[playerID]
	if !ok {
		m.mu.Unlock()
		return
	}
	if ch, ok := room.Subscribers[playerID]; ok {
		delete(room.Subscribers, playerID)
		close(ch)
	}
	delete(room.Players, playerID)
	room.PlayerOrder = filterPlayerOrder(room.PlayerOrder, playerID)
	room.LastActiveAt = time.Now()

	if room.HostPlayerID == playerID && len(room.PlayerOrder) > 0 {
		room.HostPlayerID = room.PlayerOrder[0]
	}

	if !silent {
		room.LastSystemMessage = fmt.Sprintf("%s 离开了房间", player.Name)
		room.appendMessageLocked(Message{
			ID:        newID("msg"),
			PlayerID:  player.ID,
			Player:    player.Name,
			Content:   room.LastSystemMessage,
			Kind:      "system",
			CreatedAt: room.LastActiveAt,
		})
	}

	if len(room.PlayerOrder) == 0 {
		delete(m.rooms, room.ID)
		m.mu.Unlock()
		return
	}

	if room.CurrentDrawerID == playerID && room.Status == "playing" {
		m.advanceRoundLocked(room, "当前画手离开了房间，重新抽取下一轮")
	}
	dispatches := m.collectSnapshotsLocked(room)
	m.mu.Unlock()
	sendDispatches(dispatches)
}

func (m *Manager) StartGame(roomID string, input StartGameInput) (RoomView, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, player, err := m.roomAndPlayerLocked(roomID, input.PlayerID)
	if err != nil {
		return RoomView{}, err
	}
	if room.HostPlayerID != player.ID {
		return RoomView{}, fmt.Errorf("只有房主可以开始游戏")
	}
	if len(room.PlayerOrder) < 2 {
		return RoomView{}, fmt.Errorf("至少需要 2 名玩家才能开始")
	}
	if room.Status == "playing" {
		return RoomView{}, fmt.Errorf("游戏已经开始")
	}

	m.startRoundLocked(room, 0)
	view := room.snapshotFor(player.ID)
	dispatches := m.collectSnapshotsLocked(room)
	go sendDispatches(dispatches)
	return view, nil
}

func (m *Manager) SubmitGuess(roomID string, input GuessInput) (RoomView, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, player, err := m.roomAndPlayerLocked(roomID, input.PlayerID)
	if err != nil {
		return RoomView{}, err
	}
	if room.Status != "playing" {
		return RoomView{}, fmt.Errorf("当前不在游戏中")
	}
	if room.CurrentDrawerID == player.ID {
		return RoomView{}, fmt.Errorf("画手不能猜词")
	}

	content := strings.TrimSpace(input.Content)
	if content == "" {
		return RoomView{}, fmt.Errorf("请输入猜测内容")
	}
	if len([]rune(content)) > 24 {
		return RoomView{}, fmt.Errorf("猜测内容过长")
	}

	now := time.Now()
	room.LastActiveAt = now
	isCorrect := strings.EqualFold(strings.TrimSpace(content), strings.TrimSpace(room.CurrentWord))
	kind := "guess"
	if isCorrect {
		kind = "correct"
		remaining := room.remainingSeconds()
		guessScore := 60 + remaining*2
		drawerScore := 35 + remaining
		player.Score += guessScore
		if drawer, ok := room.Players[room.CurrentDrawerID]; ok {
			drawer.Score += drawerScore
		}
		room.LastSystemMessage = fmt.Sprintf("%s 猜中了答案「%s」", player.Name, room.CurrentWord)
		content = room.LastSystemMessage
	}

	room.appendMessageLocked(Message{
		ID:        newID("msg"),
		PlayerID:  player.ID,
		Player:    player.Name,
		Content:   content,
		Kind:      kind,
		CreatedAt: now,
	})

	view := room.snapshotFor(player.ID)
	if isCorrect {
		m.advanceRoundLocked(room, room.LastSystemMessage)
	}
	dispatches := m.collectSnapshotsLocked(room)
	go sendDispatches(dispatches)
	return view, nil
}

func (m *Manager) AddStroke(roomID string, input StrokeInput) error {
	m.mu.Lock()
	room, player, err := m.roomAndPlayerLocked(roomID, input.PlayerID)
	if err != nil {
		m.mu.Unlock()
		return err
	}
	if room.Status != "playing" {
		m.mu.Unlock()
		return fmt.Errorf("当前不在绘画状态")
	}
	if room.CurrentDrawerID != player.ID {
		m.mu.Unlock()
		return fmt.Errorf("只有当前画手可以绘画")
	}
	if time.Since(player.LastStrokeAt) < minStrokeInterval {
		m.mu.Unlock()
		return nil
	}
	if len(input.Points) < 2 || len(input.Points) > maxStrokePoints {
		m.mu.Unlock()
		return fmt.Errorf("单次绘制点数不合法")
	}

	action := CanvasAction{
		ID:        newID("canvas"),
		PlayerID:  player.ID,
		Player:    player.Name,
		Kind:      "stroke",
		Color:     normalizeColor(input.Color),
		Width:     normalizeBrushWidth(input.Width),
		Tool:      normalizeTool(input.Tool),
		Points:    input.Points,
		CreatedAt: time.Now(),
	}
	player.LastStrokeAt = action.CreatedAt
	room.LastActiveAt = action.CreatedAt
	room.appendCanvasActionLocked(action)
	dispatches := m.collectCanvasDispatchesLocked(room, action)
	m.mu.Unlock()
	sendDispatches(dispatches)
	return nil
}

func (m *Manager) ClearCanvas(roomID string, input ClearInput) error {
	m.mu.Lock()
	room, player, err := m.roomAndPlayerLocked(roomID, input.PlayerID)
	if err != nil {
		m.mu.Unlock()
		return err
	}
	if room.CurrentDrawerID != player.ID {
		m.mu.Unlock()
		return fmt.Errorf("只有当前画手可以清空画布")
	}
	action := CanvasAction{
		ID:        newID("canvas"),
		PlayerID:  player.ID,
		Player:    player.Name,
		Kind:      "clear",
		CreatedAt: time.Now(),
	}
	room.CanvasActions = append(room.CanvasActions, action)
	if len(room.CanvasActions) > maxCanvasActions {
		room.CanvasActions = room.CanvasActions[len(room.CanvasActions)-maxCanvasActions:]
	}
	room.LastActiveAt = action.CreatedAt
	dispatches := m.collectCanvasDispatchesLocked(room, action)
	m.mu.Unlock()
	sendDispatches(dispatches)
	return nil
}

func (m *Manager) cleanupLoop() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		var channels []chan []byte
		m.mu.Lock()
		for id, room := range m.rooms {
			if len(room.PlayerOrder) == 0 || now.Sub(room.LastActiveAt) > roomIdleTTL {
				for _, ch := range room.Subscribers {
					channels = append(channels, ch)
				}
				delete(m.rooms, id)
			}
		}
		m.mu.Unlock()
		for _, ch := range channels {
			close(ch)
		}
	}
}

func (m *Manager) startRoundLocked(room *Room, roundOffset int) {
	if len(room.PlayerOrder) < 2 {
		room.Status = "waiting"
		room.CurrentDrawerID = ""
		room.CurrentWord = ""
		room.RoundEndsAt = time.Time{}
		return
	}
	nextIndex := 0
	if room.CurrentDrawerID != "" {
		for i, id := range room.PlayerOrder {
			if id == room.CurrentDrawerID {
				nextIndex = (i + 1 + roundOffset) % len(room.PlayerOrder)
				break
			}
		}
	}
	room.Round++
	room.Status = "playing"
	room.CurrentDrawerID = room.PlayerOrder[nextIndex]
	room.CurrentWord = pickWord()
	room.RoundEndsAt = time.Now().Add(roundDuration)
	room.RoundToken++
	room.CanvasActions = nil
	room.LastActiveAt = time.Now()
	drawerName := room.currentDrawerName()
	room.LastSystemMessage = fmt.Sprintf("第 %d 轮开始，轮到 %s 作画", room.Round, drawerName)
	room.appendMessageLocked(Message{
		ID:        newID("msg"),
		Content:   room.LastSystemMessage,
		Kind:      "system",
		CreatedAt: room.LastActiveAt,
	})
	token := room.RoundToken
	roomID := room.ID
	go m.waitRoundTimeout(roomID, token)
}

func (m *Manager) advanceRoundLocked(room *Room, reason string) {
	room.Status = "cooldown"
	room.LastSystemMessage = reason
	room.RoundToken++
	token := room.RoundToken
	roomID := room.ID
	go func() {
		time.Sleep(roundCooldown)
		m.mu.Lock()
		defer m.mu.Unlock()
		r, ok := m.rooms[roomID]
		if !ok || r.RoundToken != token || len(r.PlayerOrder) < 2 {
			if ok && len(r.PlayerOrder) < 2 {
				r.Status = "waiting"
				r.CurrentDrawerID = ""
				r.CurrentWord = ""
				r.RoundEndsAt = time.Time{}
			}
			return
		}
		m.startRoundLocked(r, 0)
		dispatches := m.collectSnapshotsLocked(r)
		go sendDispatches(dispatches)
	}()
}

func (m *Manager) waitRoundTimeout(roomID string, token int) {
	<-time.After(roundDuration)
	m.mu.Lock()
	room, ok := m.rooms[roomID]
	if !ok || room.RoundToken != token || room.Status != "playing" {
		m.mu.Unlock()
		return
	}
	room.appendMessageLocked(Message{
		ID:        newID("msg"),
		Content:   fmt.Sprintf("时间到，本轮答案是「%s」", room.CurrentWord),
		Kind:      "system",
		CreatedAt: time.Now(),
	})
	room.LastSystemMessage = fmt.Sprintf("时间到，本轮答案是「%s」", room.CurrentWord)
	m.advanceRoundLocked(room, room.LastSystemMessage)
	dispatches := m.collectSnapshotsLocked(room)
	m.mu.Unlock()
	sendDispatches(dispatches)
}

func (m *Manager) roomAndPlayerLocked(roomID, playerID string) (*Room, *Player, error) {
	room, ok := m.rooms[strings.ToUpper(strings.TrimSpace(roomID))]
	if !ok {
		return nil, nil, fmt.Errorf("房间不存在")
	}
	player, ok := room.Players[strings.TrimSpace(playerID)]
	if !ok {
		return nil, nil, fmt.Errorf("玩家不存在，请重新加入房间")
	}
	return room, player, nil
}

func (m *Manager) collectSnapshotsLocked(room *Room) []dispatch {
	dispatches := make([]dispatch, 0, len(room.Subscribers))
	for playerID, ch := range room.Subscribers {
		payload, _ := json.Marshal(eventEnvelope{
			Type: "snapshot",
			Data: room.snapshotFor(playerID),
		})
		dispatches = append(dispatches, dispatch{ch: ch, payload: payload})
	}
	return dispatches
}

func (m *Manager) collectCanvasDispatchesLocked(room *Room, action CanvasAction) []dispatch {
	dispatches := make([]dispatch, 0, len(room.Subscribers))
	for _, ch := range room.Subscribers {
		payload, _ := json.Marshal(eventEnvelope{
			Type: "canvas",
			Data: action,
		})
		dispatches = append(dispatches, dispatch{ch: ch, payload: payload})
	}
	return dispatches
}

func (m *Manager) generateRoomIDLocked() string {
	for {
		code := randomCode(6)
		if _, exists := m.rooms[code]; !exists {
			return code
		}
	}
}

func (r *Room) snapshotFor(playerID string) RoomView {
	players := make([]PlayerView, 0, len(r.PlayerOrder))
	for _, id := range r.PlayerOrder {
		p, ok := r.Players[id]
		if !ok {
			continue
		}
		players = append(players, PlayerView{
			ID:          p.ID,
			Name:        p.Name,
			Score:       p.Score,
			IsHost:      id == r.HostPlayerID,
			IsDrawer:    id == r.CurrentDrawerID,
			IsConnected: p.IsConnected,
		})
	}
	sort.SliceStable(players, func(i, j int) bool {
		if players[i].Score == players[j].Score {
			return players[i].Name < players[j].Name
		}
		return players[i].Score > players[j].Score
	})

	wordDisplay := maskWord(r.CurrentWord)
	if playerID == r.CurrentDrawerID && r.CurrentWord != "" {
		wordDisplay = r.CurrentWord
	}

	actions := make([]CanvasAction, len(r.CanvasActions))
	copy(actions, r.CanvasActions)
	messages := make([]Message, len(r.Messages))
	copy(messages, r.Messages)

	return RoomView{
		ID:                r.ID,
		Status:            r.Status,
		HostPlayerID:      r.HostPlayerID,
		CurrentDrawerID:   r.CurrentDrawerID,
		CurrentDrawerName: r.currentDrawerName(),
		Round:             r.Round,
		WordDisplay:       wordDisplay,
		RemainingSeconds:  r.remainingSeconds(),
		MaxPlayers:        maxPlayersPerRoom,
		Players:           players,
		Messages:          messages,
		CanvasActions:     actions,
		LastSystemMessage: r.LastSystemMessage,
	}
}

func (r *Room) currentDrawerName() string {
	if p, ok := r.Players[r.CurrentDrawerID]; ok {
		return p.Name
	}
	return ""
}

func (r *Room) remainingSeconds() int {
	if r.RoundEndsAt.IsZero() {
		return 0
	}
	seconds := int(time.Until(r.RoundEndsAt).Seconds())
	if seconds < 0 {
		return 0
	}
	return seconds
}

func (r *Room) appendMessageLocked(msg Message) {
	r.Messages = append(r.Messages, msg)
	if len(r.Messages) > maxMessagesPerRoom {
		r.Messages = r.Messages[len(r.Messages)-maxMessagesPerRoom:]
	}
}

func (r *Room) appendCanvasActionLocked(action CanvasAction) {
	r.CanvasActions = append(r.CanvasActions, action)
	if len(r.CanvasActions) > maxCanvasActions {
		r.CanvasActions = r.CanvasActions[len(r.CanvasActions)-maxCanvasActions:]
	}
}

type dispatch struct {
	ch      chan []byte
	payload []byte
}

type eventEnvelope struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func sendDispatches(dispatches []dispatch) {
	for _, item := range dispatches {
		select {
		case item.ch <- item.payload:
		default:
		}
	}
}

func normalizePlayerName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("请先填写玩家名称")
	}
	if len([]rune(name)) > maxPlayerNameLen {
		return "", fmt.Errorf("玩家名称不能超过 %d 个字符", maxPlayerNameLen)
	}
	return name, nil
}

func normalizeColor(color string) string {
	color = strings.TrimSpace(color)
	if color == "" {
		return "#2f3e46"
	}
	if len(color) > 16 {
		return "#2f3e46"
	}
	return color
}

func normalizeTool(tool string) string {
	switch strings.ToLower(strings.TrimSpace(tool)) {
	case "eraser":
		return "eraser"
	default:
		return "pen"
	}
}

func normalizeBrushWidth(width int) int {
	if width <= 0 {
		return defaultBrushWidth
	}
	if width > 24 {
		return 24
	}
	return width
}

func filterPlayerOrder(order []string, playerID string) []string {
	out := make([]string, 0, len(order))
	for _, id := range order {
		if id != playerID {
			out = append(out, id)
		}
	}
	return out
}

func newID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

func randomCode(length int) string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	var b strings.Builder
	for i := 0; i < length; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		b.WriteByte(alphabet[n.Int64()])
	}
	return b.String()
}

func pickWord() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(wordLibrary))))
	return wordLibrary[n.Int64()]
}

func maskWord(word string) string {
	if word == "" {
		return "等待开始"
	}
	rs := []rune(word)
	return strings.Repeat("●", len(rs)) + fmt.Sprintf("（%d 字）", len(rs))
}
