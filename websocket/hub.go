package websocket

import (
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/gofiber/contrib/socketio"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type HubImpl struct {
	clients    map[string]*Client
	rooms      map[string]map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage
	mu         sync.RWMutex
	BuildHash  string
}

type BroadcastMessage struct {
	Room  string
	Event string
	Data  any
}

func NewHub() *HubImpl {
	return &HubImpl{
		clients:    make(map[string]*Client),
		rooms:      make(map[string]map[string]*Client),
		register:   make(chan *Client, 256),
		unregister: make(chan *Client, 256),
		broadcast:  make(chan *BroadcastMessage, 256),
	}
}

func (h *HubImpl) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case msg := <-h.broadcast:
			h.broadcastToRoom(msg.Room, msg.Event, msg.Data)
		}
	}
}

func (h *HubImpl) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client.ID] = client
	slog.Debug("Client registered", "client_id", client.ID, "user_id", client.UserID)
}

func (h *HubImpl) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client.ID]; !ok {
		return
	}

	for room := range client.Rooms {
		if clients, ok := h.rooms[room]; ok {
			delete(clients, client.ID)
			if len(clients) == 0 {
				delete(h.rooms, room)
			}
		}
	}

	delete(h.clients, client.ID)
	close(client.Send)
	slog.Debug("Client unregistered", "client_id", client.ID)
}

func (h *HubImpl) broadcastToRoom(room, event string, data any) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.rooms[room]
	if !ok {
		return
	}

	msg := Message{
		Type:      event,
		Room:      room,
		Data:      data,
		Timestamp: time.Now(),
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		slog.Error("Failed to marshal message", "error", err)
		return
	}

	for _, client := range clients {
		select {
		case client.Send <- payload:
		default:
			slog.Debug("Client send buffer full, skipping", "client_id", client.ID)
		}
	}
}

func (h *HubImpl) RegisterClient(client *Client) {
	h.register <- client
}

func (h *HubImpl) RegisterSocketIO(client *Client, kws *socketio.Websocket) {
	client.SocketIOConn = kws
	h.register <- client
	go h.socketIOWritePump(client)
}

func (h *HubImpl) socketIOWritePump(client *Client) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		if client.SocketIOConn != nil {
			client.SocketIOConn.Close()
		}
	}()

	for {
		select {
		case message, ok := <-client.Send:
			if !ok {
				return
			}

			if client.SocketIOConn != nil {
				client.SocketIOConn.Emit(message, socketio.TextMessage)
			}

		case <-ticker.C:
			if client.SocketIOConn != nil {
				pingData := []byte(`{"type":"ping"}`)
				client.SocketIOConn.Emit(pingData, socketio.TextMessage)
			}
		}
	}
}

func (h *HubImpl) UnregisterClient(clientID string) {
	h.mu.RLock()
	client, ok := h.clients[clientID]
	h.mu.RUnlock()

	if ok {
		h.unregister <- client
	}
}

func (h *HubImpl) Unregister(client *Client) {
	h.unregister <- client
}

func (h *HubImpl) Broadcast(room string, event string, data any) {
	h.broadcast <- &BroadcastMessage{
		Room:  room,
		Event: event,
		Data:  data,
	}
}

func (h *HubImpl) BroadcastToUser(userID primitive.ObjectID, event string, data any) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	msg := Message{
		Type:      event,
		Data:      data,
		UserID:    userID,
		Timestamp: time.Now(),
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		slog.Error("Failed to marshal message", "error", err)
		return
	}

	sent := false
	for _, client := range h.clients {
		if client.UserID == userID {
			select {
			case client.Send <- payload:
				sent = true
				slog.Debug("Sent message to user", "user_id", userID.Hex(), "event", event)
			default:
				slog.Warn("Client send buffer full, skipping", "user_id", userID.Hex(), "event", event)
			}
		}
	}

	if !sent {
		slog.Warn("No connected client found for user", "user_id", userID.Hex(), "event", event)
	}
}

func (h *HubImpl) BroadcastToOrg(orgID primitive.ObjectID, event string, data any) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	msg := Message{
		Type:      event,
		Data:      data,
		Timestamp: time.Now(),
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		slog.Error("Failed to marshal message", "error", err)
		return
	}

	for _, client := range h.clients {
		if client.OrganizationID == orgID {
			select {
			case client.Send <- payload:
			default:
				slog.Debug("Failed to send message to org", "org_id", orgID.Hex())
			}
		}
	}
}

func (h *HubImpl) BroadcastToAll(event string, data any) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	msg := Message{
		Type:      event,
		Data:      data,
		Timestamp: time.Now(),
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		slog.Error("Failed to marshal message", "error", err)
		return
	}

	for _, client := range h.clients {
		select {
		case client.Send <- payload:
		default:
			slog.Debug("Failed to send message to client", "client_id", client.ID)
		}
	}
}

func (h *HubImpl) SendToUser(userID primitive.ObjectID, data map[string]any) {
	msg := Message{
		Type:      "message",
		Data:      data,
		Timestamp: time.Now(),
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		slog.Error("Failed to marshal user message", "error", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		if client.UserID == userID {
			select {
			case client.Send <- payload:
			default:
				slog.Debug("Failed to send message to user", "user_id", userID.Hex())
			}
		}
	}
}

func (h *HubImpl) SendToOrganization(orgID primitive.ObjectID, data map[string]any) {
	msg := Message{
		Type:      "message",
		Data:      data,
		Timestamp: time.Now(),
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		slog.Error("Failed to marshal organization message", "error", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		if client.OrganizationID == orgID {
			select {
			case client.Send <- payload:
			default:
				slog.Debug("Failed to send message to org", "org_id", orgID.Hex())
			}
		}
	}
}

func (h *HubImpl) JoinRoom(clientID string, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client, ok := h.clients[clientID]
	if !ok {
		return
	}

	if _, ok := h.rooms[room]; !ok {
		h.rooms[room] = make(map[string]*Client)
	}

	h.rooms[room][clientID] = client
	client.Rooms[room] = true

	slog.Debug("Client joined room", "client_id", clientID, "room", room)
}

func (h *HubImpl) LeaveRoom(clientID string, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client, ok := h.clients[clientID]
	if !ok {
		return
	}

	if clients, ok := h.rooms[room]; ok {
		delete(clients, clientID)
		if len(clients) == 0 {
			delete(h.rooms, room)
		}
	}

	delete(client.Rooms, room)

	slog.Debug("Client left room", "client_id", clientID, "room", room)
}

func (h *HubImpl) GetRoomClients(room string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.rooms[room]
	if !ok {
		return []string{}
	}

	ids := make([]string, 0, len(clients))
	for id := range clients {
		ids = append(ids, id)
	}
	return ids
}

func (h *HubImpl) GetRoomInfo(room string) *RoomInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.rooms[room]
	if !ok {
		return nil
	}

	userIDs := make([]string, 0, len(clients))
	for id, client := range clients {
		userIDs = append(userIDs, client.UserID.Hex())
		_ = id
	}

	return &RoomInfo{
		Name:        room,
		ClientCount: len(clients),
		UserIDs:     userIDs,
	}
}

func (h *HubImpl) GetClient(clientID string) *Client {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.clients[clientID]
}

func (h *HubImpl) GetClientByUUID(uuid string) *Client {
	return h.GetClient(uuid)
}

func (h *HubImpl) GetClientsByUser(userID primitive.ObjectID) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var clients []*Client
	for _, client := range h.clients {
		if client.UserID == userID {
			clients = append(clients, client)
		}
	}
	return clients
}

func (h *HubImpl) GetClientsByOrg(orgID primitive.ObjectID) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var clients []*Client
	for _, client := range h.clients {
		if client.OrganizationID == orgID {
			clients = append(clients, client)
		}
	}
	return clients
}

func (h *HubImpl) IsUserInOrg(userID, orgID primitive.ObjectID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		if client.UserID == userID && client.OrganizationID == orgID {
			return true
		}
	}
	return false
}

func (h *HubImpl) GetOnlineUsers(orgID primitive.ObjectID) []*PresenceUpdate {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var users []*PresenceUpdate
	seen := make(map[primitive.ObjectID]bool)

	for _, client := range h.clients {
		if client.OrganizationID == orgID && !seen[client.UserID] {
			seen[client.UserID] = true
			users = append(users, &PresenceUpdate{
				UserID:       client.UserID,
				UserName:     client.UserName,
				State:        client.PresenceState,
				LastActivity: client.LastActivity,
				OnlineSince:  client.OnlineSince,
				AwaySince:    client.AwaySince,
				TeamID:       client.TeamID,
				TeamName:     client.TeamName,
			})
		}
	}
	return users
}

func (h *HubImpl) GetOnlineUsersForOrg(orgID string) []string {
	orgRoom := "org:" + orgID + ":users"
	clients := h.GetRoomClients(orgRoom)

	h.mu.RLock()
	defer h.mu.RUnlock()

	seen := make(map[string]bool)
	users := []string{}

	for _, clientID := range clients {
		client := h.clients[clientID]
		if client != nil {
			userID := client.UserID.Hex()
			if !seen[userID] {
				seen[userID] = true
				users = append(users, userID)
			}
		}
	}

	return users
}

func (h *HubImpl) GetOnlineUsersForTeam(orgID, teamID string) []string {
	teamRoom := "org:" + orgID + ":team:" + teamID + ":users"
	clients := h.GetRoomClients(teamRoom)

	h.mu.RLock()
	defer h.mu.RUnlock()

	seen := make(map[string]bool)
	users := []string{}

	for _, clientID := range clients {
		client := h.clients[clientID]
		if client != nil {
			userID := client.UserID.Hex()
			if !seen[userID] {
				seen[userID] = true
				users = append(users, userID)
			}
		}
	}

	return users
}

func (h *HubImpl) GetTeamOnlineUsers(orgID, teamID primitive.ObjectID) []*PresenceUpdate {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var users []*PresenceUpdate
	seen := make(map[primitive.ObjectID]bool)

	for _, client := range h.clients {
		if client.OrganizationID == orgID &&
			client.TeamID == teamID &&
			!seen[client.UserID] {
			seen[client.UserID] = true
			users = append(users, &PresenceUpdate{
				UserID:       client.UserID,
				UserName:     client.UserName,
				State:        client.PresenceState,
				LastActivity: client.LastActivity,
				OnlineSince:  client.OnlineSince,
				AwaySince:    client.AwaySince,
				TeamID:       client.TeamID,
				TeamName:     client.TeamName,
			})
		}
	}
	return users
}

func (h *HubImpl) UserLogin(orgID, userID, userName, teamID, teamName string) {
	slog.Debug("User logged in", "org_id", orgID, "user_id", userID, "name", userName)

	orgRoom := "org:" + orgID + ":users"
	h.BroadcastToRoom(orgRoom, EventUserOnline, map[string]any{
		"user_id":   userID,
		"name":      userName,
		"team_id":   teamID,
		"team_name": teamName,
	})

	if teamID != "" {
		teamRoom := "org:" + orgID + ":team:" + teamID + ":users"
		h.BroadcastToRoom(teamRoom, EventUserOnline, map[string]any{
			"user_id":   userID,
			"name":      userName,
			"team_id":   teamID,
			"team_name": teamName,
		})
	}
}

func (h *HubImpl) UserLogout(orgID, userID, teamID, userName string) {
	slog.Debug("User logged out", "org_id", orgID, "user_id", userID)

	orgRoom := "org:" + orgID + ":users"
	h.BroadcastToRoom(orgRoom, EventUserOffline, map[string]any{
		"user_id": userID,
		"name":    userName,
	})

	if teamID != "" {
		teamRoom := "org:" + orgID + ":team:" + teamID + ":users"
		h.BroadcastToRoom(teamRoom, EventUserOffline, map[string]any{
			"user_id": userID,
			"name":    userName,
			"team_id": teamID,
		})
	}
}

func (h *HubImpl) UserWentAway(orgID, userID, teamID, userName string, onlineDurationSecs int64, idleThreshold int) {
	slog.Debug("User went away", "org_id", orgID, "user_id", userID, "online_duration", onlineDurationSecs)

	orgRoom := "org:" + orgID + ":users"
	h.BroadcastToRoom(orgRoom, EventUserAway, map[string]any{
		"user_id":              userID,
		"name":                 userName,
		"online_duration_secs": onlineDurationSecs,
		"idle_threshold_secs":  idleThreshold,
	})

	if teamID != "" {
		teamRoom := "org:" + orgID + ":team:" + teamID + ":users"
		h.BroadcastToRoom(teamRoom, EventUserAway, map[string]any{
			"user_id":              userID,
			"name":                 userName,
			"team_id":              teamID,
			"online_duration_secs": onlineDurationSecs,
			"idle_threshold_secs":  idleThreshold,
		})
	}
}

func (h *HubImpl) UserCameBack(orgID, userID, teamID, userName string, awayDurationSecs int64) {
	slog.Debug("User came back", "org_id", orgID, "user_id", userID, "away_duration", awayDurationSecs)

	orgRoom := "org:" + orgID + ":users"
	h.BroadcastToRoom(orgRoom, EventUserBack, map[string]any{
		"user_id":            userID,
		"name":               userName,
		"away_duration_secs": awayDurationSecs,
	})

	if teamID != "" {
		teamRoom := "org:" + orgID + ":team:" + teamID + ":users"
		h.BroadcastToRoom(teamRoom, EventUserBack, map[string]any{
			"user_id":            userID,
			"name":               userName,
			"team_id":            teamID,
			"away_duration_secs": awayDurationSecs,
		})
	}
}

func (h *HubImpl) BroadcastToRoom(room string, event string, data any) {
	msg := Message{
		Type: event,
		Room: room,
		Data: data,
	}
	jsonData, err := json.Marshal(msg)
	if err != nil {
		slog.Error("Error marshaling broadcast message", "error", err)
		return
	}

	slog.Debug("BroadcastToRoom", "room", room, "event", event, "clients_in_room", len(h.rooms[room]))

	h.broadcast <- &BroadcastMessage{
		Room:  room,
		Event: event,
		Data:  data,
	}

	h.mu.RLock()
	roomClients, exists := h.rooms[room]
	if !exists {
		h.mu.RUnlock()
		return
	}

	var clientsToRemove []string
	for _, client := range roomClients {
		select {
		case client.Send <- jsonData:
			slog.Debug("Message sent to client", "client_id", client.ID, "user_id", client.UserID.Hex())
		default:
			clientsToRemove = append(clientsToRemove, client.ID)
		}
	}
	h.mu.RUnlock()

	if len(clientsToRemove) > 0 {
		h.mu.Lock()
		for _, clientID := range clientsToRemove {
			if client, ok := h.clients[clientID]; ok {
				close(client.Send)
				delete(h.clients, clientID)
			}
		}
		h.mu.Unlock()
	}
}

func (h *HubImpl) SetBuildHash(hash string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.BuildHash = hash
}

func (h *HubImpl) GetBuildHash() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.BuildHash
}

func (h *HubImpl) GetStats() map[string]any {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return map[string]any{
		"clients":    len(h.clients),
		"rooms":      len(h.rooms),
		"build_hash": h.BuildHash,
	}
}

var (
	globalHub *HubImpl
	hubOnce   sync.Once
)

func GetHub() *HubImpl {
	hubOnce.Do(func() {
		globalHub = NewHub()
	})
	return globalHub
}

func InitHub() *HubImpl {
	return GetHub()
}
