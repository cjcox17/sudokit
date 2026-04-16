package testing

import (
	"testing"
	"time"

	"github.com/cjcox17/sudokit/websocket"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMockHub_Broadcast(t *testing.T) {
	mock := NewMockHub()

	mock.Broadcast("test-room", "test.event", map[string]any{"foo": "bar"})

	if len(mock.BroadcastCalls) != 1 {
		t.Errorf("expected 1 broadcast call, got %d", len(mock.BroadcastCalls))
	}

	call := mock.BroadcastCalls[0]
	if call.Room != "test-room" {
		t.Errorf("expected room 'test-room', got '%s'", call.Room)
	}

	if call.Event != "test.event" {
		t.Errorf("expected event 'test.event', got '%s'", call.Event)
	}
}

func TestMockHub_BroadcastToUser(t *testing.T) {
	mock := NewMockHub()
	userID := primitive.NewObjectID()

	mock.BroadcastToUser(userID, "notification_new", map[string]any{"title": "Test"})

	if len(mock.BroadcastToUserCalls) != 1 {
		t.Errorf("expected 1 broadcast to user call, got %d", len(mock.BroadcastToUserCalls))
	}

	call := mock.BroadcastToUserCalls[0]
	if call.UserID != userID {
		t.Error("expected user ID to match")
	}

	if call.Event != "notification_new" {
		t.Errorf("expected event 'notification_new', got '%s'", call.Event)
	}
}

func TestMockHub_BroadcastToOrg(t *testing.T) {
	mock := NewMockHub()
	orgID := primitive.NewObjectID()

	mock.BroadcastToOrg(orgID, "event", map[string]any{"data": "test"})

	if len(mock.BroadcastToOrgCalls) != 1 {
		t.Errorf("expected 1 broadcast to org call, got %d", len(mock.BroadcastToOrgCalls))
	}

	call := mock.BroadcastToOrgCalls[0]
	if call.OrgID != orgID {
		t.Error("expected org ID to match")
	}
}

func TestMockHub_JoinRoom(t *testing.T) {
	mock := NewMockHub()
	userID := primitive.NewObjectID()
	orgID := primitive.NewObjectID()
	client := CreateTestClient(userID, orgID)

	mock.RegisterClient(client)
	mock.JoinRoom(client.ID, "org:"+orgID.Hex())

	if len(mock.JoinRoomCalls) != 1 {
		t.Errorf("expected 1 join room call, got %d", len(mock.JoinRoomCalls))
	}

	call := mock.JoinRoomCalls[0]
	if call.ClientID != client.ID {
		t.Error("expected client ID to match")
	}

	if call.Room != "org:"+orgID.Hex() {
		t.Errorf("expected room 'org:%s', got '%s'", orgID.Hex(), call.Room)
	}
}

func TestMockHub_LeaveRoom(t *testing.T) {
	mock := NewMockHub()
	userID := primitive.NewObjectID()
	orgID := primitive.NewObjectID()
	client := CreateTestClient(userID, orgID)

	mock.RegisterClient(client)
	mock.JoinRoom(client.ID, "test-room")
	mock.LeaveRoom(client.ID, "test-room")

	if len(mock.LeaveRoomCalls) != 1 {
		t.Errorf("expected 1 leave room call, got %d", len(mock.LeaveRoomCalls))
	}
}

func TestMockHub_GetRoomClients(t *testing.T) {
	mock := NewMockHub()
	userID := primitive.NewObjectID()
	orgID := primitive.NewObjectID()
	client := CreateTestClient(userID, orgID)

	mock.RegisterClient(client)
	mock.JoinRoom(client.ID, "test-room")

	clients := mock.GetRoomClients("test-room")

	if len(clients) != 1 {
		t.Errorf("expected 1 client in room, got %d", len(clients))
	}
}

func TestMockHub_GetRoomInfo(t *testing.T) {
	mock := NewMockHub()
	userID := primitive.NewObjectID()
	orgID := primitive.NewObjectID()
	client := CreateTestClient(userID, orgID)

	mock.RegisterClient(client)
	mock.JoinRoom(client.ID, "test-room")

	info := mock.GetRoomInfo("test-room")

	if info == nil {
		t.Fatal("expected room info, got nil")
	}

	if info.ClientCount != 1 {
		t.Errorf("expected client count 1, got %d", info.ClientCount)
	}
}

func TestMockHub_RegisterClient(t *testing.T) {
	mock := NewMockHub()
	userID := primitive.NewObjectID()
	orgID := primitive.NewObjectID()
	client := CreateTestClient(userID, orgID)

	mock.RegisterClient(client)

	if len(mock.RegisterClientCalls) != 1 {
		t.Errorf("expected 1 register call, got %d", len(mock.RegisterClientCalls))
	}

	if mock.GetClient(client.ID) == nil {
		t.Error("expected client to be registered")
	}
}

func TestMockHub_UnregisterClient(t *testing.T) {
	mock := NewMockHub()
	userID := primitive.NewObjectID()
	orgID := primitive.NewObjectID()
	client := CreateTestClient(userID, orgID)

	mock.RegisterClient(client)
	mock.UnregisterClient(client.ID)

	if len(mock.UnregisterClientCalls) != 1 {
		t.Errorf("expected 1 unregister call, got %d", len(mock.UnregisterClientCalls))
	}

	if mock.GetClient(client.ID) != nil {
		t.Error("expected client to be unregistered")
	}
}

func TestMockHub_GetOnlineUsers(t *testing.T) {
	mock := NewMockHub()
	orgID := primitive.NewObjectID()
	userID1 := primitive.NewObjectID()
	userID2 := primitive.NewObjectID()

	client1 := CreateTestClient(userID1, orgID)
	client2 := CreateTestClient(userID2, orgID)

	mock.RegisterClient(client1)
	mock.RegisterClient(client2)

	users := mock.GetOnlineUsers(orgID)

	if len(users) != 2 {
		t.Errorf("expected 2 online users, got %d", len(users))
	}
}

func TestMockHub_GetTeamOnlineUsers(t *testing.T) {
	mock := NewMockHub()
	orgID := primitive.NewObjectID()
	teamID := primitive.NewObjectID()
	userID1 := primitive.NewObjectID()
	userID2 := primitive.NewObjectID()

	client1 := CreateTestClient(userID1, orgID)
	client1.TeamID = teamID

	client2 := CreateTestClient(userID2, orgID)

	mock.RegisterClient(client1)
	mock.RegisterClient(client2)

	users := mock.GetTeamOnlineUsers(orgID, teamID)

	if len(users) != 1 {
		t.Errorf("expected 1 team online user, got %d", len(users))
	}
}

func TestMockHub_Reset(t *testing.T) {
	mock := NewMockHub()
	userID := primitive.NewObjectID()
	orgID := primitive.NewObjectID()

	mock.Broadcast("room", "event", nil)
	mock.BroadcastToUser(userID, "event", nil)
	mock.BroadcastToOrg(orgID, "event", nil)
	mock.JoinRoom("client", "room")
	mock.LeaveRoom("client", "room")

	mock.Reset()

	if len(mock.BroadcastCalls) != 0 {
		t.Error("expected broadcast calls to be reset")
	}

	if len(mock.BroadcastToUserCalls) != 0 {
		t.Error("expected broadcast to user calls to be reset")
	}

	if len(mock.BroadcastToOrgCalls) != 0 {
		t.Error("expected broadcast to org calls to be reset")
	}
}

func TestMockHub_GetLastBroadcastToUser(t *testing.T) {
	mock := NewMockHub()
	userID1 := primitive.NewObjectID()
	userID2 := primitive.NewObjectID()

	mock.BroadcastToUser(userID1, "event1", nil)
	mock.BroadcastToUser(userID2, "event2", nil)

	call := mock.GetLastBroadcastToUser()

	if call == nil {
		t.Fatal("expected call, got nil")
	}

	if call.UserID != userID2 {
		t.Error("expected last user ID to match")
	}
}

func TestAssertBroadcastToUser(t *testing.T) {
	mock := NewMockHub()
	userID := primitive.NewObjectID()

	mock.BroadcastToUser(userID, "event", nil)

	AssertBroadcastToUser(t, mock, userID)
}

func TestAssertBroadcastToOrg(t *testing.T) {
	mock := NewMockHub()
	orgID := primitive.NewObjectID()

	mock.BroadcastToOrg(orgID, "event", nil)

	AssertBroadcastToOrg(t, mock, orgID)
}

func TestCreateTestClient(t *testing.T) {
	userID := primitive.NewObjectID()
	orgID := primitive.NewObjectID()

	client := CreateTestClient(userID, orgID)

	if client.UserID != userID {
		t.Error("expected user ID to match")
	}

	if client.OrganizationID != orgID {
		t.Error("expected org ID to match")
	}

	if client.PresenceState != websocket.PresenceOnline {
		t.Errorf("expected presence state '%s', got '%s'", websocket.PresenceOnline, client.PresenceState)
	}

	if client.Rooms == nil {
		t.Error("expected rooms map to be initialized")
	}
}

func TestCreateTestClientWithID(t *testing.T) {
	clientID := "custom-client-id"
	userID := primitive.NewObjectID()
	orgID := primitive.NewObjectID()

	client := CreateTestClientWithID(clientID, userID, orgID)

	if client.ID != clientID {
		t.Errorf("expected client ID '%s', got '%s'", clientID, client.ID)
	}
}

func TestPresenceConstants(t *testing.T) {
	tests := []struct {
		name     string
		state    string
		expected string
	}{
		{"online", websocket.PresenceOnline, "online"},
		{"away", websocket.PresenceAway, "away"},
		{"offline", websocket.PresenceOffline, "offline"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.state != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.state)
			}
		})
	}
}

func TestEventConstants(t *testing.T) {
	tests := []struct {
		name     string
		event    string
		expected string
	}{
		{"subscribe", websocket.EventSubscribe, "subscribe"},
		{"unsubscribe", websocket.EventUnsubscribe, "unsubscribe"},
		{"ping", websocket.EventPing, "ping"},
		{"pong", websocket.EventPong, "pong"},
		{"user_joined", websocket.EventUserJoined, "user_joined"},
		{"user_left", websocket.EventUserLeft, "user_left"},
		{"user_online", websocket.EventUserOnline, "user_online"},
		{"user_offline", websocket.EventUserOffline, "user_offline"},
		{"notification", websocket.EventNotification, "notification_new"},
		{"unread_count", websocket.EventUnreadCount, "notification_unread_count"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.event != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.event)
			}
		})
	}
}

func TestMessageStruct(t *testing.T) {
	userID := primitive.NewObjectID()
	now := time.Now()

	msg := websocket.Message{
		Type:      "test.event",
		Room:      "test-room",
		Data:      map[string]any{"key": "value"},
		UserID:    userID,
		Timestamp: now,
	}

	if msg.Type != "test.event" {
		t.Errorf("expected type 'test.event', got '%s'", msg.Type)
	}

	if msg.Room != "test-room" {
		t.Errorf("expected room 'test-room', got '%s'", msg.Room)
	}

	if msg.UserID != userID {
		t.Error("expected user ID to match")
	}
}

func TestPresenceUpdateStruct(t *testing.T) {
	userID := primitive.NewObjectID()
	teamID := primitive.NewObjectID()
	now := time.Now()

	update := websocket.PresenceUpdate{
		UserID:       userID,
		UserName:     "Test User",
		State:        websocket.PresenceOnline,
		LastActivity: now,
		OnlineSince:  now,
		TeamID:       teamID,
		TeamName:     "Test Team",
	}

	if update.UserID != userID {
		t.Error("expected user ID to match")
	}

	if update.UserName != "Test User" {
		t.Errorf("expected user name 'Test User', got '%s'", update.UserName)
	}

	if update.State != websocket.PresenceOnline {
		t.Errorf("expected state '%s', got '%s'", websocket.PresenceOnline, update.State)
	}

	if update.TeamID != teamID {
		t.Error("expected team ID to match")
	}
}

func TestRoomInfoStruct(t *testing.T) {
	info := websocket.RoomInfo{
		Name:        "test-room",
		ClientCount: 5,
		UserIDs:     []string{"user1", "user2"},
	}

	if info.Name != "test-room" {
		t.Errorf("expected name 'test-room', got '%s'", info.Name)
	}

	if info.ClientCount != 5 {
		t.Errorf("expected client count 5, got %d", info.ClientCount)
	}

	if len(info.UserIDs) != 2 {
		t.Errorf("expected 2 user IDs, got %d", len(info.UserIDs))
	}
}

func TestClientStruct(t *testing.T) {
	userID := primitive.NewObjectID()
	orgID := primitive.NewObjectID()
	now := time.Now()

	client := &websocket.Client{
		ID:             "test-client",
		UserID:         userID,
		OrganizationID: orgID,
		UserName:       "Test User",
		UserEmail:      "test@example.com",
		Rooms:          map[string]bool{"room1": true, "room2": true},
		PresenceState:  websocket.PresenceOnline,
		LastActivity:   now,
		OnlineSince:    now,
		IdleThreshold:  90,
	}

	if client.ID != "test-client" {
		t.Errorf("expected ID 'test-client', got '%s'", client.ID)
	}

	if client.UserName != "Test User" {
		t.Errorf("expected user name 'Test User', got '%s'", client.UserName)
	}

	if len(client.Rooms) != 2 {
		t.Errorf("expected 2 rooms, got %d", len(client.Rooms))
	}

	if client.IdleThreshold != 90 {
		t.Errorf("expected idle threshold 90, got %d", client.IdleThreshold)
	}
}

func TestMockHub_BroadcastToAll(t *testing.T) {
	mock := NewMockHub()

	mock.BroadcastToAll("global.event", map[string]any{"data": "test"})

	if len(mock.BroadcastCalls) != 1 {
		t.Errorf("expected 1 broadcast call, got %d", len(mock.BroadcastCalls))
	}
}

func TestMockHub_GetClientsByUser(t *testing.T) {
	mock := NewMockHub()
	userID := primitive.NewObjectID()
	orgID := primitive.NewObjectID()

	client1 := CreateTestClientWithID("client1", userID, orgID)
	client2 := CreateTestClientWithID("client2", userID, orgID)

	mock.RegisterClient(client1)
	mock.RegisterClient(client2)

	clients := mock.GetClientsByUser(userID)

	if len(clients) != 2 {
		t.Errorf("expected 2 clients for user, got %d", len(clients))
	}
}

func TestMockHub_GetClientsByOrg(t *testing.T) {
	mock := NewMockHub()
	orgID := primitive.NewObjectID()
	userID1 := primitive.NewObjectID()
	userID2 := primitive.NewObjectID()

	client1 := CreateTestClient(userID1, orgID)
	client2 := CreateTestClient(userID2, orgID)

	mock.RegisterClient(client1)
	mock.RegisterClient(client2)

	clients := mock.GetClientsByOrg(orgID)

	if len(clients) != 2 {
		t.Errorf("expected 2 clients for org, got %d", len(clients))
	}
}
