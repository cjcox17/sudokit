package testing

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/cjcox17/sudokit/email"
	"github.com/cjcox17/sudokit/events"
	"github.com/cjcox17/sudokit/jobs"
	"github.com/cjcox17/sudokit/notifications"
	"github.com/cjcox17/sudokit/sms"
	"github.com/cjcox17/sudokit/websocket"

	"github.com/gofiber/contrib/socketio"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MockJobsService struct {
	mu sync.RWMutex

	EnqueueFunc func(ctx context.Context, jobType string, payload map[string]any, opts ...jobs.Option) (primitive.ObjectID, error)
	GetFunc     func(ctx context.Context, jobID primitive.ObjectID) (*jobs.Job, error)
	CancelFunc  func(ctx context.Context, jobID primitive.ObjectID) error
	StartFunc   func() error
	StopFunc    func(ctx context.Context) error

	EnqueueCalls []EnqueueCall
	GetCalls     []GetCall
	CancelCalls  []CancelCall
}

type EnqueueCall struct {
	JobType string
	Payload map[string]any
	Options []jobs.Option
}

type GetCall struct {
	JobID primitive.ObjectID
}

type CancelCall struct {
	JobID primitive.ObjectID
}

func NewMockJobsService() *MockJobsService {
	return &MockJobsService{
		EnqueueCalls: make([]EnqueueCall, 0),
		GetCalls:     make([]GetCall, 0),
		CancelCalls:  make([]CancelCall, 0),
	}
}

func (m *MockJobsService) Enqueue(ctx context.Context, jobType string, payload map[string]any, opts ...jobs.Option) (primitive.ObjectID, error) {
	m.mu.Lock()
	m.EnqueueCalls = append(m.EnqueueCalls, EnqueueCall{JobType: jobType, Payload: payload, Options: opts})
	m.mu.Unlock()

	if m.EnqueueFunc != nil {
		return m.EnqueueFunc(ctx, jobType, payload, opts...)
	}
	return primitive.NewObjectID(), nil
}

func (m *MockJobsService) Get(ctx context.Context, jobID primitive.ObjectID) (*jobs.Job, error) {
	m.mu.Lock()
	m.GetCalls = append(m.GetCalls, GetCall{JobID: jobID})
	m.mu.Unlock()

	if m.GetFunc != nil {
		return m.GetFunc(ctx, jobID)
	}
	return &jobs.Job{ID: jobID, Status: jobs.StatusPending}, nil
}

func (m *MockJobsService) GetByUserID(ctx context.Context, userID primitive.ObjectID, limit int) ([]*jobs.Job, error) {
	return []*jobs.Job{}, nil
}

func (m *MockJobsService) GetByOrganizationID(ctx context.Context, orgID primitive.ObjectID, limit, skip int) ([]*jobs.Job, int, error) {
	return []*jobs.Job{}, 0, nil
}

func (m *MockJobsService) Cancel(ctx context.Context, jobID primitive.ObjectID) error {
	m.mu.Lock()
	m.CancelCalls = append(m.CancelCalls, CancelCall{JobID: jobID})
	m.mu.Unlock()

	if m.CancelFunc != nil {
		return m.CancelFunc(ctx, jobID)
	}
	return nil
}

func (m *MockJobsService) RegisterHandler(jobType string, handler jobs.Handler) {}

func (m *MockJobsService) RegisterHandlerFunc(jobType string, handler jobs.HandlerFunc) {}

func (m *MockJobsService) UpdateProgress(ctx context.Context, jobID primitive.ObjectID, progress int, message string) error {
	return nil
}

func (m *MockJobsService) BroadcastUpdate(ctx context.Context, jobID primitive.ObjectID, progress int, message string) error {
	return nil
}

func (m *MockJobsService) KeepAlive(ctx context.Context, jobID primitive.ObjectID) error {
	return nil
}

func (m *MockJobsService) Start() error {
	if m.StartFunc != nil {
		return m.StartFunc()
	}
	return nil
}

func (m *MockJobsService) Stop(ctx context.Context) error {
	if m.StopFunc != nil {
		return m.StopFunc(ctx)
	}
	return nil
}

func (m *MockJobsService) IsStarted() bool { return true }

func (m *MockJobsService) RegisteredTypes() []string { return []string{} }

type MockEventsService struct {
	mu sync.RWMutex

	PublishFunc   func(ctx context.Context, eventType string, payload any, opts ...events.EventOption) error
	SubscribeFunc func(eventType string, handler events.Handler) error
	StartFunc     func() error
	StopFunc      func(ctx context.Context) error

	PublishCalls   []PublishCall
	SubscribeCalls []SubscribeCall
}

type PublishCall struct {
	EventType string
	Payload   any
	Options   []events.EventOption
}

type SubscribeCall struct {
	EventType string
}

func NewMockEventsService() *MockEventsService {
	return &MockEventsService{
		PublishCalls:   make([]PublishCall, 0),
		SubscribeCalls: make([]SubscribeCall, 0),
	}
}

func (m *MockEventsService) Publish(ctx context.Context, eventType string, payload any, opts ...events.EventOption) error {
	m.mu.Lock()
	m.PublishCalls = append(m.PublishCalls, PublishCall{EventType: eventType, Payload: payload, Options: opts})
	m.mu.Unlock()

	if m.PublishFunc != nil {
		return m.PublishFunc(ctx, eventType, payload, opts...)
	}
	return nil
}

func (m *MockEventsService) Subscribe(eventType string, handler events.Handler) error {
	m.mu.Lock()
	m.SubscribeCalls = append(m.SubscribeCalls, SubscribeCall{EventType: eventType})
	m.mu.Unlock()

	if m.SubscribeFunc != nil {
		return m.SubscribeFunc(eventType, handler)
	}
	return nil
}

func (m *MockEventsService) SubscribeSync(eventType string, handler events.Handler) error {
	return nil
}

func (m *MockEventsService) Start() error {
	if m.StartFunc != nil {
		return m.StartFunc()
	}
	return nil
}

func (m *MockEventsService) Stop(ctx context.Context) error {
	if m.StopFunc != nil {
		return m.StopFunc(ctx)
	}
	return nil
}

type MockNotificationsService struct {
	mu sync.RWMutex

	CreateFunc         func(ctx context.Context, userID, orgID primitive.ObjectID, input notifications.CreateInput) (*notifications.Notification, error)
	CreateForUsersFunc func(ctx context.Context, userIDs []primitive.ObjectID, orgID primitive.ObjectID, input notifications.CreateInput) error
	ListFunc           func(ctx context.Context, userID primitive.ObjectID, opts notifications.ListOptions) ([]*notifications.Notification, int, error)
	GetUnreadCountFunc func(ctx context.Context, userID primitive.ObjectID) (int, error)
	MarkAsReadFunc     func(ctx context.Context, notificationID, userID primitive.ObjectID) error
	MarkAllAsReadFunc  func(ctx context.Context, userID primitive.ObjectID) (int, error)
	DeleteFunc         func(ctx context.Context, notificationID, userID primitive.ObjectID) error

	CreateCalls         []CreateCall
	CreateForUsersCalls []CreateForUsersCall
	ListCalls           []ListCall
	MarkAsReadCalls     []MarkAsReadCall
}

type CreateCall struct {
	UserID primitive.ObjectID
	OrgID  primitive.ObjectID
	Input  notifications.CreateInput
}

type CreateForUsersCall struct {
	UserIDs []primitive.ObjectID
	OrgID   primitive.ObjectID
	Input   notifications.CreateInput
}

type ListCall struct {
	UserID primitive.ObjectID
	Opts   notifications.ListOptions
}

type MarkAsReadCall struct {
	NotificationID primitive.ObjectID
	UserID         primitive.ObjectID
}

func NewMockNotificationsService() *MockNotificationsService {
	return &MockNotificationsService{
		CreateCalls:         make([]CreateCall, 0),
		CreateForUsersCalls: make([]CreateForUsersCall, 0),
		ListCalls:           make([]ListCall, 0),
		MarkAsReadCalls:     make([]MarkAsReadCall, 0),
	}
}

func (m *MockNotificationsService) EnsureIndexes(ctx context.Context) error {
	return nil
}

func (m *MockNotificationsService) Create(ctx context.Context, userID, orgID primitive.ObjectID, input notifications.CreateInput) (*notifications.Notification, error) {
	m.mu.Lock()
	m.CreateCalls = append(m.CreateCalls, CreateCall{UserID: userID, OrgID: orgID, Input: input})
	m.mu.Unlock()

	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, userID, orgID, input)
	}
	return &notifications.Notification{ID: primitive.NewObjectID()}, nil
}

func (m *MockNotificationsService) CreateForUsers(ctx context.Context, userIDs []primitive.ObjectID, orgID primitive.ObjectID, input notifications.CreateInput) error {
	m.mu.Lock()
	m.CreateForUsersCalls = append(m.CreateForUsersCalls, CreateForUsersCall{UserIDs: userIDs, OrgID: orgID, Input: input})
	m.mu.Unlock()

	if m.CreateForUsersFunc != nil {
		return m.CreateForUsersFunc(ctx, userIDs, orgID, input)
	}
	return nil
}

func (m *MockNotificationsService) CreateForOrg(ctx context.Context, orgID primitive.ObjectID, input notifications.CreateInput) error {
	return nil
}

func (m *MockNotificationsService) List(ctx context.Context, userID primitive.ObjectID, opts notifications.ListOptions) ([]*notifications.Notification, int, error) {
	m.mu.Lock()
	m.ListCalls = append(m.ListCalls, ListCall{UserID: userID, Opts: opts})
	m.mu.Unlock()

	if m.ListFunc != nil {
		return m.ListFunc(ctx, userID, opts)
	}
	return []*notifications.Notification{}, 0, nil
}

func (m *MockNotificationsService) GetUnreadCount(ctx context.Context, userID primitive.ObjectID) (int, error) {
	if m.GetUnreadCountFunc != nil {
		return m.GetUnreadCountFunc(ctx, userID)
	}
	return 0, nil
}

func (m *MockNotificationsService) MarkAsRead(ctx context.Context, notificationID, userID primitive.ObjectID) error {
	m.mu.Lock()
	m.MarkAsReadCalls = append(m.MarkAsReadCalls, MarkAsReadCall{NotificationID: notificationID, UserID: userID})
	m.mu.Unlock()

	if m.MarkAsReadFunc != nil {
		return m.MarkAsReadFunc(ctx, notificationID, userID)
	}
	return nil
}

func (m *MockNotificationsService) MarkAllAsRead(ctx context.Context, userID primitive.ObjectID) (int, error) {
	if m.MarkAllAsReadFunc != nil {
		return m.MarkAllAsReadFunc(ctx, userID)
	}
	return 0, nil
}

func (m *MockNotificationsService) Delete(ctx context.Context, notificationID, userID primitive.ObjectID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, notificationID, userID)
	}
	return nil
}

type MockWebSocketHub struct {
	mu sync.RWMutex

	BroadcastCalls       []BroadcastCall
	BroadcastToUserCalls []BroadcastToUserCall
	BroadcastToOrgCalls  []BroadcastToOrgCall
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

func NewMockWebSocketHub() *MockWebSocketHub {
	return &MockWebSocketHub{
		BroadcastCalls:       make([]BroadcastCall, 0),
		BroadcastToUserCalls: make([]BroadcastToUserCall, 0),
		BroadcastToOrgCalls:  make([]BroadcastToOrgCall, 0),
	}
}

func (m *MockWebSocketHub) Broadcast(room string, event string, data any) {
	m.mu.Lock()
	m.BroadcastCalls = append(m.BroadcastCalls, BroadcastCall{Room: room, Event: event, Data: data})
	m.mu.Unlock()
}

func (m *MockWebSocketHub) BroadcastToUser(userID primitive.ObjectID, event string, data any) {
	m.mu.Lock()
	m.BroadcastToUserCalls = append(m.BroadcastToUserCalls, BroadcastToUserCall{UserID: userID, Event: event, Data: data})
	m.mu.Unlock()
}

func (m *MockWebSocketHub) BroadcastToOrg(orgID primitive.ObjectID, event string, data any) {
	m.mu.Lock()
	m.BroadcastToOrgCalls = append(m.BroadcastToOrgCalls, BroadcastToOrgCall{OrgID: orgID, Event: event, Data: data})
	m.mu.Unlock()
}

func (m *MockWebSocketHub) BroadcastToAll(event string, data any) {}

func (m *MockWebSocketHub) SendToUser(userID primitive.ObjectID, data map[string]any) {}

func (m *MockWebSocketHub) SendToOrganization(orgID primitive.ObjectID, data map[string]any) {}

func (m *MockWebSocketHub) JoinRoom(clientID string, room string) {}

func (m *MockWebSocketHub) LeaveRoom(clientID string, room string) {}

func (m *MockWebSocketHub) GetRoomClients(room string) []string { return []string{} }

func (m *MockWebSocketHub) GetRoomInfo(room string) *websocket.RoomInfo { return nil }

func (m *MockWebSocketHub) GetClient(clientID string) *websocket.Client { return nil }

func (m *MockWebSocketHub) GetClientByUUID(uuid string) *websocket.Client { return nil }

func (m *MockWebSocketHub) GetClientsByUser(userID primitive.ObjectID) []*websocket.Client {
	return nil
}

func (m *MockWebSocketHub) GetClientsByOrg(orgID primitive.ObjectID) []*websocket.Client { return nil }

func (m *MockWebSocketHub) RegisterClient(client *websocket.Client) {}

func (m *MockWebSocketHub) RegisterSocketIO(client *websocket.Client, kws *socketio.Websocket) {}

func (m *MockWebSocketHub) UnregisterClient(clientID string) {}

func (m *MockWebSocketHub) Unregister(client *websocket.Client) {}

func (m *MockWebSocketHub) GetOnlineUsers(orgID primitive.ObjectID) []*websocket.PresenceUpdate {
	return nil
}

func (m *MockWebSocketHub) GetOnlineUsersForOrg(orgID string) []string { return []string{} }

func (m *MockWebSocketHub) GetOnlineUsersForTeam(orgID, teamID string) []string { return []string{} }

func (m *MockWebSocketHub) GetTeamOnlineUsers(orgID, teamID primitive.ObjectID) []*websocket.PresenceUpdate {
	return nil
}

func (m *MockWebSocketHub) UserLogin(orgID, userID, userName, teamID, teamName string) {}

func (m *MockWebSocketHub) UserLogout(orgID, userID, teamID, userName string) {}

func (m *MockWebSocketHub) UserWentAway(orgID, userID, teamID, userName string, onlineDurationSecs int64, idleThreshold int) {
}

func (m *MockWebSocketHub) UserCameBack(orgID, userID, teamID, userName string, awayDurationSecs int64) {
}

func (m *MockWebSocketHub) BroadcastToRoom(room string, event string, data any) {}

func (m *MockWebSocketHub) SetBuildHash(hash string) {}

func (m *MockWebSocketHub) GetBuildHash() string { return "" }

type MockStorageService struct {
	mu sync.RWMutex

	UploadCalls   []MockStorageUploadCall
	DownloadCalls []MockStorageDownloadCall
	DeleteCalls   []MockStorageDeleteCall
}

type MockStorageUploadCall struct {
	Key         string
	ContentType string
}

type MockStorageDownloadCall struct {
	Key string
}

type MockStorageDeleteCall struct {
	Key string
}

func NewMockStorageService() *MockStorageService {
	return &MockStorageService{
		UploadCalls:   make([]MockStorageUploadCall, 0),
		DownloadCalls: make([]MockStorageDownloadCall, 0),
		DeleteCalls:   make([]MockStorageDeleteCall, 0),
	}
}

func (m *MockStorageService) Upload(ctx context.Context, key string, reader io.Reader, contentType string) error {
	m.mu.Lock()
	m.UploadCalls = append(m.UploadCalls, MockStorageUploadCall{Key: key, ContentType: contentType})
	m.mu.Unlock()
	return nil
}

func (m *MockStorageService) UploadBytes(ctx context.Context, key string, data []byte, contentType string) error {
	return nil
}

func (m *MockStorageService) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	m.mu.Lock()
	m.DownloadCalls = append(m.DownloadCalls, MockStorageDownloadCall{Key: key})
	m.mu.Unlock()
	return nil, nil
}

func (m *MockStorageService) DownloadBytes(ctx context.Context, key string) ([]byte, error) {
	return []byte("mock-content"), nil
}

func (m *MockStorageService) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	m.DeleteCalls = append(m.DeleteCalls, MockStorageDeleteCall{Key: key})
	m.mu.Unlock()
	return nil
}

func (m *MockStorageService) DeleteMultiple(ctx context.Context, keys []string) error {
	return nil
}

func (m *MockStorageService) Exists(ctx context.Context, key string) (bool, error) {
	return true, nil
}

func (m *MockStorageService) PresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return "https://mock-presigned-url.com/" + key, nil
}

func (m *MockStorageService) PresignedUploadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return "https://mock-presigned-upload-url.com/" + key, nil
}

func (m *MockStorageService) Copy(ctx context.Context, srcKey, dstKey string) error {
	return nil
}

func (m *MockStorageService) GetMetadata(ctx context.Context, key string) (map[string]string, error) {
	return map[string]string{}, nil
}

func (m *MockStorageService) List(ctx context.Context, prefix string) ([]string, error) {
	return []string{}, nil
}

type MockEmailService struct {
	mu sync.RWMutex

	SendCalls         []MockEmailSendCall
	SendTemplateCalls []MockEmailSendTemplateCall
}

type MockEmailSendCall struct {
	Input *email.SendInput
}

type MockEmailSendTemplateCall struct {
	TemplateID string
	To         []string
	Data       map[string]any
}

func NewMockEmailService() *MockEmailService {
	return &MockEmailService{
		SendCalls:         make([]MockEmailSendCall, 0),
		SendTemplateCalls: make([]MockEmailSendTemplateCall, 0),
	}
}

func (m *MockEmailService) Send(ctx context.Context, input *email.SendInput) (*email.SendOutput, error) {
	m.mu.Lock()
	m.SendCalls = append(m.SendCalls, MockEmailSendCall{Input: input})
	m.mu.Unlock()
	return &email.SendOutput{MessageID: "mock-message-id"}, nil
}

func (m *MockEmailService) SendTemplate(ctx context.Context, templateID string, to []string, data map[string]any) (*email.SendOutput, error) {
	m.mu.Lock()
	m.SendTemplateCalls = append(m.SendTemplateCalls, MockEmailSendTemplateCall{TemplateID: templateID, To: to, Data: data})
	m.mu.Unlock()
	return &email.SendOutput{MessageID: "mock-message-id"}, nil
}

func (m *MockEmailService) Render(ctx context.Context, input *email.RenderInput) (*email.RenderOutput, error) {
	return &email.RenderOutput{}, nil
}

func (m *MockEmailService) GetTemplate(ctx context.Context, templateID string) (*email.Template, error) {
	return &email.Template{ID: templateID}, nil
}

func (m *MockEmailService) HealthCheck(ctx context.Context) error {
	return nil
}

type MockSMSService struct {
	mu sync.RWMutex

	SendCalls         []MockSMSSendCall
	SendTemplateCalls []MockSMSSendTemplateCall
}

type MockSMSSendCall struct {
	Input *sms.SendInput
}

type MockSMSSendTemplateCall struct {
	TemplateID string
	To         string
	Data       map[string]any
}

func NewMockSMSService() *MockSMSService {
	return &MockSMSService{
		SendCalls:         make([]MockSMSSendCall, 0),
		SendTemplateCalls: make([]MockSMSSendTemplateCall, 0),
	}
}

func (m *MockSMSService) Send(ctx context.Context, input *sms.SendInput) (*sms.SendOutput, error) {
	m.mu.Lock()
	m.SendCalls = append(m.SendCalls, MockSMSSendCall{Input: input})
	m.mu.Unlock()
	return &sms.SendOutput{MessageID: "mock-sms-id"}, nil
}

func (m *MockSMSService) SendTemplate(ctx context.Context, templateID string, to string, data map[string]any) (*sms.SendOutput, error) {
	m.mu.Lock()
	m.SendTemplateCalls = append(m.SendTemplateCalls, MockSMSSendTemplateCall{TemplateID: templateID, To: to, Data: data})
	m.mu.Unlock()
	return &sms.SendOutput{MessageID: "mock-sms-id"}, nil
}

func (m *MockSMSService) GetTemplate(ctx context.Context, templateID string) (*sms.Template, error) {
	return &sms.Template{ID: templateID}, nil
}

func (m *MockSMSService) HealthCheck(ctx context.Context) error {
	return nil
}

type MockCacheService struct {
	mu sync.RWMutex

	GetCalls    []MockCacheGetCall
	SetCalls    []MockCacheSetCall
	DeleteCalls []MockCacheDeleteCall
}

type MockCacheGetCall struct {
	Key string
}

type MockCacheSetCall struct {
	Key   string
	Value []byte
	TTL   time.Duration
}

type MockCacheDeleteCall struct {
	Key string
}

func NewMockCacheService() *MockCacheService {
	return &MockCacheService{
		GetCalls:    make([]MockCacheGetCall, 0),
		SetCalls:    make([]MockCacheSetCall, 0),
		DeleteCalls: make([]MockCacheDeleteCall, 0),
	}
}

func (m *MockCacheService) Get(ctx context.Context, key string) ([]byte, error) {
	m.mu.Lock()
	m.GetCalls = append(m.GetCalls, MockCacheGetCall{Key: key})
	m.mu.Unlock()
	return []byte("mock-cache-value"), nil
}

func (m *MockCacheService) GetString(ctx context.Context, key string) (string, error) {
	return "mock-cache-value", nil
}

func (m *MockCacheService) GetInt(ctx context.Context, key string) (int64, error) {
	return 0, nil
}

func (m *MockCacheService) GetJSON(ctx context.Context, key string, v interface{}) error {
	return nil
}

func (m *MockCacheService) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	m.mu.Lock()
	m.SetCalls = append(m.SetCalls, MockCacheSetCall{Key: key, Value: value, TTL: ttl})
	m.mu.Unlock()
	return nil
}

func (m *MockCacheService) SetString(ctx context.Context, key string, value string, ttl time.Duration) error {
	return m.Set(ctx, key, []byte(value), ttl)
}

func (m *MockCacheService) SetInt(ctx context.Context, key string, value int64, ttl time.Duration) error {
	return nil
}

func (m *MockCacheService) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return nil
}

func (m *MockCacheService) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	m.DeleteCalls = append(m.DeleteCalls, MockCacheDeleteCall{Key: key})
	m.mu.Unlock()
	return nil
}

func (m *MockCacheService) DeleteByPattern(ctx context.Context, pattern string) error {
	return nil
}

func (m *MockCacheService) Exists(ctx context.Context, key string) (bool, error) {
	return true, nil
}

func (m *MockCacheService) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return nil
}

func (m *MockCacheService) TTL(ctx context.Context, key string) (time.Duration, error) {
	return 0, nil
}

func (m *MockCacheService) Incr(ctx context.Context, key string) (int64, error) {
	return 1, nil
}

func (m *MockCacheService) Decr(ctx context.Context, key string) (int64, error) {
	return 0, nil
}

func (m *MockCacheService) Remember(ctx context.Context, key string, ttl time.Duration, fn func() ([]byte, error)) ([]byte, error) {
	return fn()
}

func (m *MockCacheService) RememberString(ctx context.Context, key string, ttl time.Duration, fn func() (string, error)) (string, error) {
	return fn()
}

func (m *MockCacheService) RememberJSON(ctx context.Context, key string, ttl time.Duration, fn func() (interface{}, error), v interface{}) error {
	return nil
}

func (m *MockCacheService) HealthCheck(ctx context.Context) error {
	return nil
}
