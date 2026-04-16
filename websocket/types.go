package websocket

import (
	"sync"
	"time"

	"github.com/gofiber/contrib/socketio"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Client struct {
	ID             string              `json:"id"`
	UserID         primitive.ObjectID  `json:"user_id"`
	OrganizationID primitive.ObjectID  `json:"organization_id"`
	TeamID         primitive.ObjectID  `json:"team_id,omitempty"`
	TeamName       string              `json:"team_name,omitempty"`
	UserName       string              `json:"user_name"`
	UserEmail      string              `json:"user_email"`
	SocketIOConn   *socketio.Websocket `json:"-"`
	Rooms          map[string]bool     `json:"rooms"`
	Send           chan []byte         `json:"-"`
	Hub            *HubImpl            `json:"-"`
	mu             sync.RWMutex        `json:"-"`

	PresenceState string     `json:"presence_state"`
	LastActivity  time.Time  `json:"last_activity"`
	OnlineSince   time.Time  `json:"online_since"`
	AwaySince     *time.Time `json:"away_since,omitempty"`
	IdleThreshold int        `json:"idle_threshold"`
}

type Room struct {
	Name        string    `json:"name"`
	ClientCount int       `json:"client_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type Message struct {
	Type      string             `json:"type"`
	Room      string             `json:"room,omitempty"`
	Data      any                `json:"data"`
	UserID    primitive.ObjectID `json:"user_id,omitempty"`
	Timestamp time.Time          `json:"timestamp"`
}

type PresenceUpdate struct {
	UserID       primitive.ObjectID `json:"user_id"`
	UserName     string             `json:"user_name"`
	State        string             `json:"state"`
	LastActivity time.Time          `json:"last_activity"`
	OnlineSince  time.Time          `json:"online_since"`
	AwaySince    *time.Time         `json:"away_since,omitempty"`
	TeamID       primitive.ObjectID `json:"team_id,omitempty"`
	TeamName     string             `json:"team_name,omitempty"`
}

type RoomInfo struct {
	Name        string   `json:"name"`
	ClientCount int      `json:"client_count"`
	UserIDs     []string `json:"user_ids,omitempty"`
}

const (
	PresenceOnline  = "online"
	PresenceAway    = "away"
	PresenceOffline = "offline"
)

const (
	EventSubscribe   = "subscribe"
	EventUnsubscribe = "unsubscribe"
	EventPing        = "ping"
	EventPong        = "pong"
	EventActivity    = "activity"
	EventAway        = "away"
	EventBack        = "back"

	EventUserJoined  = "user_joined"
	EventUserLeft    = "user_left"
	EventUserOnline  = "user_online"
	EventUserOffline = "user_offline"
	EventUserAway    = "user_away"
	EventUserBack    = "user_back"

	EventPresenceSnapshot = "presence_snapshot"
	EventVersionCheck     = "version_check"
	EventNotification     = "notification_new"
	EventUnreadCount      = "notification_unread_count"
)

type Hub interface {
	Broadcast(room string, event string, data any)
	BroadcastToUser(userID primitive.ObjectID, event string, data any)
	BroadcastToOrg(orgID primitive.ObjectID, event string, data any)
	BroadcastToAll(event string, data any)

	SendToUser(userID primitive.ObjectID, data map[string]any)
	SendToOrganization(orgID primitive.ObjectID, data map[string]any)

	JoinRoom(clientID string, room string)
	LeaveRoom(clientID string, room string)
	GetRoomClients(room string) []string
	GetRoomInfo(room string) *RoomInfo

	GetClient(clientID string) *Client
	GetClientByUUID(uuid string) *Client
	GetClientsByUser(userID primitive.ObjectID) []*Client
	GetClientsByOrg(orgID primitive.ObjectID) []*Client

	RegisterClient(client *Client)
	RegisterSocketIO(client *Client, kws *socketio.Websocket)
	UnregisterClient(clientID string)
	Unregister(client *Client)

	GetOnlineUsers(orgID primitive.ObjectID) []*PresenceUpdate
	GetOnlineUsersForOrg(orgID string) []string
	GetOnlineUsersForTeam(orgID, teamID string) []string
	GetTeamOnlineUsers(orgID, teamID primitive.ObjectID) []*PresenceUpdate

	UserLogin(orgID, userID, userName, teamID, teamName string)
	UserLogout(orgID, userID, teamID, userName string)
	UserWentAway(orgID, userID, teamID, userName string, onlineDurationSecs int64, idleThreshold int)
	UserCameBack(orgID, userID, teamID, userName string, awayDurationSecs int64)

	SetBuildHash(hash string)
	GetBuildHash() string
}

type UpgradeFunc func(c *fiber.Ctx) (*Client, error)
