package testing

import (
	"context"
	"testing"

	"github.com/cjcox17/sudokit/kernel"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TestKernel struct {
	Kernel        *kernel.Kernel
	MockJobs      *MockJobsService
	MockEvents    *MockEventsService
	MockNotify    *MockNotificationsService
	MockWebSocket *MockWebSocketHub
	MockStorage   *MockStorageService
	MockEmail     *MockEmailService
	MockSMS       *MockSMSService
	MockCache     *MockCacheService
}

func NewTestKernel(t *testing.T) *TestKernel {
	mockJobs := NewMockJobsService()
	mockEvents := NewMockEventsService()
	mockNotify := NewMockNotificationsService()
	mockWS := NewMockWebSocketHub()
	mockStorage := NewMockStorageService()
	mockEmail := NewMockEmailService()
	mockSMS := NewMockSMSService()
	mockCache := NewMockCacheService()

	services := &kernel.Services{
		Jobs:          mockJobs,
		Events:        mockEvents,
		Notifications: mockNotify,
		WebSocket:     mockWS,
		Storage:       mockStorage,
		Email:         mockEmail,
		SMS:           mockSMS,
		Cache:         mockCache,
	}

	tk := &TestKernel{
		MockJobs:      mockJobs,
		MockEvents:    mockEvents,
		MockNotify:    mockNotify,
		MockWebSocket: mockWS,
		MockStorage:   mockStorage,
		MockEmail:     mockEmail,
		MockSMS:       mockSMS,
		MockCache:     mockCache,
	}

	tk.Kernel = kernel.NewWithServices(services)

	return tk
}

func (tk *TestKernel) Context() context.Context {
	return kernel.WithServices(context.Background(), tk.Kernel)
}

func (tk *TestKernel) ContextWithValue(key, value any) context.Context {
	ctx := tk.Context()
	return context.WithValue(ctx, key, value)
}

func (tk *TestKernel) Reset() {
	tk.MockJobs.EnqueueCalls = nil
	tk.MockJobs.GetCalls = nil
	tk.MockJobs.CancelCalls = nil

	tk.MockEvents.PublishCalls = nil
	tk.MockEvents.SubscribeCalls = nil

	tk.MockNotify.CreateCalls = nil
	tk.MockNotify.CreateForUsersCalls = nil
	tk.MockNotify.ListCalls = nil
	tk.MockNotify.MarkAsReadCalls = nil

	tk.MockWebSocket.BroadcastCalls = nil
	tk.MockWebSocket.BroadcastToUserCalls = nil
	tk.MockWebSocket.BroadcastToOrgCalls = nil
}

func AssertJobEnqueued(t *testing.T, tk *TestKernel, expectedJobType string) {
	t.Helper()

	found := false
	for _, call := range tk.MockJobs.EnqueueCalls {
		if call.JobType == expectedJobType {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected job of type '%s' to be enqueued, but it was not", expectedJobType)
	}
}

func AssertEventPublished(t *testing.T, tk *TestKernel, expectedEventType string) {
	t.Helper()

	found := false
	for _, call := range tk.MockEvents.PublishCalls {
		if call.EventType == expectedEventType {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected event of type '%s' to be published, but it was not", expectedEventType)
	}
}

func AssertNotificationCreated(t *testing.T, tk *TestKernel, expectedUserID primitive.ObjectID) {
	t.Helper()

	found := false
	for _, call := range tk.MockNotify.CreateCalls {
		if call.UserID == expectedUserID {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected notification to be created for user '%s', but it was not", expectedUserID.Hex())
	}
}

func AssertBroadcastToUser(t *testing.T, tk *TestKernel, expectedUserID primitive.ObjectID) {
	t.Helper()

	found := false
	for _, call := range tk.MockWebSocket.BroadcastToUserCalls {
		if call.UserID == expectedUserID {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected broadcast to user '%s', but it was not sent", expectedUserID.Hex())
	}
}

func AssertBroadcastToOrg(t *testing.T, tk *TestKernel, expectedOrgID primitive.ObjectID) {
	t.Helper()

	found := false
	for _, call := range tk.MockWebSocket.BroadcastToOrgCalls {
		if call.OrgID == expectedOrgID {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected broadcast to org '%s', but it was not sent", expectedOrgID.Hex())
	}
}
