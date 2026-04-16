package testing

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/cjcox17/sudokit/events"
	"github.com/cjcox17/sudokit/events/payloads"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
	m.PublishCalls = append(m.PublishCalls, PublishCall{
		EventType: eventType,
		Payload:   payload,
		Options:   opts,
	})
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
	return m.Subscribe(eventType, handler)
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

func (m *MockEventsService) GetLastPublishPayload() (eventType string, payload map[string]any) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.PublishCalls) == 0 {
		return "", nil
	}

	call := m.PublishCalls[len(m.PublishCalls)-1]
	eventType = call.EventType

	if p, ok := call.Payload.(map[string]any); ok {
		payload = p
	} else {
		data, _ := json.Marshal(call.Payload)
		json.Unmarshal(data, &payload)
	}

	return eventType, payload
}

func (m *MockEventsService) GetPublishCallsByType(eventType string) []PublishCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []PublishCall
	for _, call := range m.PublishCalls {
		if call.EventType == eventType {
			result = append(result, call)
		}
	}
	return result
}

func (m *MockEventsService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PublishCalls = make([]PublishCall, 0)
	m.SubscribeCalls = make([]SubscribeCall, 0)
}

type MockBroadcaster struct {
	mu sync.RWMutex

	BroadcastToUserCalls []BroadcastToUserCall
	BroadcastToOrgCalls  []BroadcastToOrgCall
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

func NewMockBroadcaster() *MockBroadcaster {
	return &MockBroadcaster{
		BroadcastToUserCalls: make([]BroadcastToUserCall, 0),
		BroadcastToOrgCalls:  make([]BroadcastToOrgCall, 0),
	}
}

func (m *MockBroadcaster) BroadcastToUser(userID primitive.ObjectID, event string, data any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BroadcastToUserCalls = append(m.BroadcastToUserCalls, BroadcastToUserCall{
		UserID: userID,
		Event:  event,
		Data:   data,
	})
}

func (m *MockBroadcaster) BroadcastToOrg(orgID primitive.ObjectID, event string, data any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BroadcastToOrgCalls = append(m.BroadcastToOrgCalls, BroadcastToOrgCall{
		OrgID: orgID,
		Event: event,
		Data:  data,
	})
}

func (m *MockBroadcaster) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BroadcastToUserCalls = make([]BroadcastToUserCall, 0)
	m.BroadcastToOrgCalls = make([]BroadcastToOrgCall, 0)
}

func AssertEventPublished(t interface{ Errorf(string, ...any) }, mock *MockEventsService, expectedEventType string) {
	found := false
	for _, call := range mock.PublishCalls {
		if call.EventType == expectedEventType {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected event of type '%s' to be published, but it was not", expectedEventType)
	}
}

func AssertCaseFileCreated(t interface{ Errorf(string, ...any) }, mock *MockEventsService) {
	AssertEventPublished(t, mock, events.EventCaseFileCreated)
}

func AssertImportCompleted(t interface{ Errorf(string, ...any) }, mock *MockEventsService) {
	AssertEventPublished(t, mock, events.EventImportCompleted)
}

func GetPayload[T any](mock *MockEventsService) (T, bool) {
	var result T
	mock.mu.RLock()
	defer mock.mu.RUnlock()

	if len(mock.PublishCalls) == 0 {
		return result, false
	}

	call := mock.PublishCalls[len(mock.PublishCalls)-1]

	switch p := call.Payload.(type) {
	case T:
		return p, true
	case map[string]any:
		data, _ := json.Marshal(p)
		if json.Unmarshal(data, &result) == nil {
			return result, true
		}
	case payloads.CaseFileCreatedPayload:
		if typed, ok := any(p).(T); ok {
			return typed, true
		}
		data, _ := json.Marshal(p)
		if json.Unmarshal(data, &result) == nil {
			return result, true
		}
	}

	data, _ := json.Marshal(call.Payload)
	json.Unmarshal(data, &result)
	return result, true
}
