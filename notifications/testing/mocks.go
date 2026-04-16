package testing

import (
	"context"
	"sync"
	"time"

	"github.com/cjcox17/sudokit/notifications"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
	DeleteCalls         []DeleteCall
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

type DeleteCall struct {
	NotificationID primitive.ObjectID
	UserID         primitive.ObjectID
}

func NewMockNotificationsService() *MockNotificationsService {
	return &MockNotificationsService{
		CreateCalls:         make([]CreateCall, 0),
		CreateForUsersCalls: make([]CreateForUsersCall, 0),
		ListCalls:           make([]ListCall, 0),
		MarkAsReadCalls:     make([]MarkAsReadCall, 0),
		DeleteCalls:         make([]DeleteCall, 0),
	}
}

func (m *MockNotificationsService) Create(ctx context.Context, userID, orgID primitive.ObjectID, input notifications.CreateInput) (*notifications.Notification, error) {
	m.mu.Lock()
	m.CreateCalls = append(m.CreateCalls, CreateCall{
		UserID: userID,
		OrgID:  orgID,
		Input:  input,
	})
	m.mu.Unlock()

	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, userID, orgID, input)
	}

	return &notifications.Notification{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		Type:      input.Type,
		Category:  input.Category,
		Title:     input.Title,
		Message:   input.Message,
		CreatedAt: time.Now(),
	}, nil
}

func (m *MockNotificationsService) CreateForUsers(ctx context.Context, userIDs []primitive.ObjectID, orgID primitive.ObjectID, input notifications.CreateInput) error {
	m.mu.Lock()
	m.CreateForUsersCalls = append(m.CreateForUsersCalls, CreateForUsersCall{
		UserIDs: userIDs,
		OrgID:   orgID,
		Input:   input,
	})
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
	m.ListCalls = append(m.ListCalls, ListCall{
		UserID: userID,
		Opts:   opts,
	})
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
	m.MarkAsReadCalls = append(m.MarkAsReadCalls, MarkAsReadCall{
		NotificationID: notificationID,
		UserID:         userID,
	})
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
	m.mu.Lock()
	m.DeleteCalls = append(m.DeleteCalls, DeleteCall{
		NotificationID: notificationID,
		UserID:         userID,
	})
	m.mu.Unlock()

	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, notificationID, userID)
	}
	return nil
}

func (m *MockNotificationsService) EnsureIndexes(ctx context.Context) error {
	return nil
}

func (m *MockNotificationsService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CreateCalls = make([]CreateCall, 0)
	m.CreateForUsersCalls = make([]CreateForUsersCall, 0)
	m.ListCalls = make([]ListCall, 0)
	m.MarkAsReadCalls = make([]MarkAsReadCall, 0)
	m.DeleteCalls = make([]DeleteCall, 0)
}

func (m *MockNotificationsService) GetLastCreateCall() *CreateCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.CreateCalls) == 0 {
		return nil
	}
	return &m.CreateCalls[len(m.CreateCalls)-1]
}

type MockBroadcaster struct {
	mu sync.RWMutex

	BroadcastCalls []BroadcastCall
}

type BroadcastCall struct {
	UserID primitive.ObjectID
	Event  string
	Data   any
}

func NewMockBroadcaster() *MockBroadcaster {
	return &MockBroadcaster{
		BroadcastCalls: make([]BroadcastCall, 0),
	}
}

func (m *MockBroadcaster) Broadcast(userID primitive.ObjectID, event string, data any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BroadcastCalls = append(m.BroadcastCalls, BroadcastCall{
		UserID: userID,
		Event:  event,
		Data:   data,
	})
}

func (m *MockBroadcaster) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BroadcastCalls = make([]BroadcastCall, 0)
}

func (m *MockBroadcaster) GetCallsByEvent(event string) []BroadcastCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []BroadcastCall
	for _, call := range m.BroadcastCalls {
		if call.Event == event {
			result = append(result, call)
		}
	}
	return result
}

func AssertNotificationCreated(t interface{ Errorf(string, ...any) }, mock *MockNotificationsService, expectedUserID primitive.ObjectID) {
	found := false
	for _, call := range mock.CreateCalls {
		if call.UserID == expectedUserID {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected notification to be created for user '%s', but it was not", expectedUserID.Hex())
	}
}

func AssertNotificationMarkedAsRead(t interface{ Errorf(string, ...any) }, mock *MockNotificationsService, expectedNotificationID primitive.ObjectID) {
	found := false
	for _, call := range mock.MarkAsReadCalls {
		if call.NotificationID == expectedNotificationID {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected notification '%s' to be marked as read, but it was not", expectedNotificationID.Hex())
	}
}

func CreateTestNotification(userID primitive.ObjectID, notifType notifications.Type, category notifications.Category) *notifications.Notification {
	return &notifications.Notification{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		Type:      notifType,
		Category:  category,
		Title:     "Test Notification",
		Message:   "This is a test notification",
		IsRead:    false,
		CreatedAt: time.Now(),
	}
}
