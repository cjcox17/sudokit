package testing

import (
	"sync"
	"time"

	"github.com/cjcox17/sudokit/websocket"

	"github.com/gofiber/contrib/socketio"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MockHub struct {
	mu sync.RWMutex

	BroadcastCalls        []BroadcastCall
	BroadcastToUserCalls  []BroadcastToUserCall
	BroadcastToOrgCalls   []BroadcastToOrgCall
	JoinRoomCalls         []JoinRoomCall
	LeaveRoomCalls        []LeaveRoomCall
	RegisterClientCalls   []*websocket.Client
	UnregisterClientCalls []string

	clients map[string]*websocket.Client
	rooms   map[string]map[string]*websocket.Client
}

type BroadcastCall struct {
	Room  string
	Event string
	Data  any
}

type BroadcastToUserCall struct {
	UserID primitive.ObjectID
	Event  string
	Data   any
}

type BroadcastToOrgCall struct {
	OrgID primitive.ObjectID
	Event string
	Data  any
}

type JoinRoomCall struct {
	ClientID string
	Room     string
}

type LeaveRoomCall struct {
	ClientID string
	Room     string
}

func NewMockHub() *MockHub {
	return &MockHub{
		BroadcastCalls:        make([]BroadcastCall, 0),
		BroadcastToUserCalls:  make([]BroadcastToUserCall, 0),
		BroadcastToOrgCalls:   make([]BroadcastToOrgCall, 0),
		JoinRoomCalls:         make([]JoinRoomCall, 0),
		LeaveRoomCalls:        make([]LeaveRoomCall, 0),
		RegisterClientCalls:   make([]*websocket.Client, 0),
		UnregisterClientCalls: make([]string, 0),
		clients:               make(map[string]*websocket.Client),
		rooms:                 make(map[string]map[string]*websocket.Client),
	}
}

func (m *MockHub) Broadcast(room string, event string, data any) {
	m.mu.Lock()
	m.BroadcastCalls = append(m.BroadcastCalls, BroadcastCall{
		Room:  room,
		Event: event,
		Data:  data,
	})
	m.mu.Unlock()
}

func (m *MockHub) BroadcastToUser(userID primitive.ObjectID, event string, data any) {
	m.mu.Lock()
	m.BroadcastToUserCalls = append(m.BroadcastToUserCalls, BroadcastToUserCall{
		UserID: userID,
		Event:  event,
		Data:   data,
	})
	m.mu.Unlock()
}

func (m *MockHub) BroadcastToOrg(orgID primitive.ObjectID, event string, data any) {
	m.mu.Lock()
	m.BroadcastToOrgCalls = append(m.BroadcastToOrgCalls, BroadcastToOrgCall{
		OrgID: orgID,
		Event: event,
		Data:  data,
	})
	m.mu.Unlock()
}

func (m *MockHub) BroadcastToAll(event string, data any) {
	m.Broadcast("", event, data)
}

func (m *MockHub) SendToUser(userID primitive.ObjectID, data map[string]any) {
	m.BroadcastToUser(userID, "message", data)
}

func (m *MockHub) SendToOrganization(orgID primitive.ObjectID, data map[string]any) {
	m.BroadcastToOrg(orgID, "message", data)
}

func (m *MockHub) JoinRoom(clientID string, room string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.JoinRoomCalls = append(m.JoinRoomCalls, JoinRoomCall{
		ClientID: clientID,
		Room:     room,
	})

	if m.rooms[room] == nil {
		m.rooms[room] = make(map[string]*websocket.Client)
	}

	if client, ok := m.clients[clientID]; ok {
		m.rooms[room][clientID] = client
		client.Rooms[room] = true
	}
}

func (m *MockHub) LeaveRoom(clientID string, room string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.LeaveRoomCalls = append(m.LeaveRoomCalls, LeaveRoomCall{
		ClientID: clientID,
		Room:     room,
	})

	if clients, ok := m.rooms[room]; ok {
		delete(clients, clientID)
	}

	if client, ok := m.clients[clientID]; ok {
		delete(client.Rooms, room)
	}
}

func (m *MockHub) GetRoomClients(room string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clients, ok := m.rooms[room]
	if !ok {
		return []string{}
	}

	ids := make([]string, 0, len(clients))
	for id := range clients {
		ids = append(ids, id)
	}
	return ids
}

func (m *MockHub) GetRoomInfo(room string) *websocket.RoomInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clients, ok := m.rooms[room]
	if !ok {
		return nil
	}

	return &websocket.RoomInfo{
		Name:        room,
		ClientCount: len(clients),
	}
}

func (m *MockHub) GetClient(clientID string) *websocket.Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.clients[clientID]
}

func (m *MockHub) GetClientByUUID(uuid string) *websocket.Client {
	return m.GetClient(uuid)
}

func (m *MockHub) GetClientsByUser(userID primitive.ObjectID) []*websocket.Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var clients []*websocket.Client
	for _, client := range m.clients {
		if client.UserID == userID {
			clients = append(clients, client)
		}
	}
	return clients
}

func (m *MockHub) GetClientsByOrg(orgID primitive.ObjectID) []*websocket.Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var clients []*websocket.Client
	for _, client := range m.clients {
		if client.OrganizationID == orgID {
			clients = append(clients, client)
		}
	}
	return clients
}

func (m *MockHub) RegisterClient(client *websocket.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.RegisterClientCalls = append(m.RegisterClientCalls, client)
	m.clients[client.ID] = client
}

func (m *MockHub) RegisterSocketIO(client *websocket.Client, kws *socketio.Websocket) {
	m.RegisterClient(client)
}

func (m *MockHub) UnregisterClient(clientID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.UnregisterClientCalls = append(m.UnregisterClientCalls, clientID)
	delete(m.clients, clientID)
}

func (m *MockHub) Unregister(client *websocket.Client) {
	m.UnregisterClient(client.ID)
}

func (m *MockHub) GetOnlineUsers(orgID primitive.ObjectID) []*websocket.PresenceUpdate {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var users []*websocket.PresenceUpdate
	seen := make(map[primitive.ObjectID]bool)

	for _, client := range m.clients {
		if client.OrganizationID == orgID && !seen[client.UserID] {
			seen[client.UserID] = true
			users = append(users, &websocket.PresenceUpdate{
				UserID:       client.UserID,
				UserName:     client.UserName,
				State:        client.PresenceState,
				LastActivity: client.LastActivity,
				OnlineSince:  client.OnlineSince,
			})
		}
	}
	return users
}

func (m *MockHub) GetOnlineUsersForOrg(orgID string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	seen := make(map[string]bool)
	var users []string

	for _, client := range m.clients {
		if client.OrganizationID.Hex() == orgID && !seen[client.UserID.Hex()] {
			seen[client.UserID.Hex()] = true
			users = append(users, client.UserID.Hex())
		}
	}
	return users
}

func (m *MockHub) GetOnlineUsersForTeam(orgID, teamID string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	seen := make(map[string]bool)
	var users []string

	for _, client := range m.clients {
		if client.OrganizationID.Hex() == orgID &&
			client.TeamID.Hex() == teamID &&
			!seen[client.UserID.Hex()] {
			seen[client.UserID.Hex()] = true
			users = append(users, client.UserID.Hex())
		}
	}
	return users
}

func (m *MockHub) GetTeamOnlineUsers(orgID, teamID primitive.ObjectID) []*websocket.PresenceUpdate {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var users []*websocket.PresenceUpdate
	seen := make(map[primitive.ObjectID]bool)

	for _, client := range m.clients {
		if client.OrganizationID == orgID &&
			client.TeamID == teamID &&
			!seen[client.UserID] {
			seen[client.UserID] = true
			users = append(users, &websocket.PresenceUpdate{
				UserID:       client.UserID,
				UserName:     client.UserName,
				State:        client.PresenceState,
				LastActivity: client.LastActivity,
				OnlineSince:  client.OnlineSince,
			})
		}
	}
	return users
}

func (m *MockHub) UserLogin(orgID, userID, userName, teamID, teamName string) {}

func (m *MockHub) UserLogout(orgID, userID, teamID, userName string) {}

func (m *MockHub) UserWentAway(orgID, userID, teamID, userName string, onlineDurationSecs int64, idleThreshold int) {
}

func (m *MockHub) UserCameBack(orgID, userID, teamID, userName string, awayDurationSecs int64) {}

func (m *MockHub) BroadcastToRoom(room string, event string, data any) {
	m.Broadcast(room, event, data)
}

func (m *MockHub) SetBuildHash(hash string) {}

func (m *MockHub) GetBuildHash() string {
	return "test-build-hash"
}

func (m *MockHub) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BroadcastCalls = make([]BroadcastCall, 0)
	m.BroadcastToUserCalls = make([]BroadcastToUserCall, 0)
	m.BroadcastToOrgCalls = make([]BroadcastToOrgCall, 0)
	m.JoinRoomCalls = make([]JoinRoomCall, 0)
	m.LeaveRoomCalls = make([]LeaveRoomCall, 0)
	m.RegisterClientCalls = make([]*websocket.Client, 0)
	m.UnregisterClientCalls = make([]string, 0)
}

func (m *MockHub) GetLastBroadcastToUser() *BroadcastToUserCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.BroadcastToUserCalls) == 0 {
		return nil
	}
	return &m.BroadcastToUserCalls[len(m.BroadcastToUserCalls)-1]
}

func (m *MockHub) GetLastBroadcastToOrg() *BroadcastToOrgCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.BroadcastToOrgCalls) == 0 {
		return nil
	}
	return &m.BroadcastToOrgCalls[len(m.BroadcastToOrgCalls)-1]
}

func AssertBroadcastToUser(t interface{ Errorf(string, ...any) }, mock *MockHub, expectedUserID primitive.ObjectID) {
	found := false
	for _, call := range mock.BroadcastToUserCalls {
		if call.UserID == expectedUserID {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected broadcast to user '%s', but it was not sent", expectedUserID.Hex())
	}
}

func AssertBroadcastToOrg(t interface{ Errorf(string, ...any) }, mock *MockHub, expectedOrgID primitive.ObjectID) {
	found := false
	for _, call := range mock.BroadcastToOrgCalls {
		if call.OrgID == expectedOrgID {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected broadcast to org '%s', but it was not sent", expectedOrgID.Hex())
	}
}

func CreateTestClient(userID primitive.ObjectID, orgID primitive.ObjectID) *websocket.Client {
	return &websocket.Client{
		ID:             "test-client-" + userID.Hex(),
		UserID:         userID,
		OrganizationID: orgID,
		UserName:       "Test User",
		Rooms:          make(map[string]bool),
		PresenceState:  websocket.PresenceOnline,
		LastActivity:   time.Now(),
		OnlineSince:    time.Now(),
		Send:           make(chan []byte, 256),
	}
}

func CreateTestClientWithID(clientID string, userID primitive.ObjectID, orgID primitive.ObjectID) *websocket.Client {
	return &websocket.Client{
		ID:             clientID,
		UserID:         userID,
		OrganizationID: orgID,
		UserName:       "Test User",
		Rooms:          make(map[string]bool),
		PresenceState:  websocket.PresenceOnline,
		LastActivity:   time.Now(),
		OnlineSince:    time.Now(),
		Send:           make(chan []byte, 256),
	}
}
