package testing

import (
	"context"
	"testing"
	"time"

	"github.com/cjcox17/sudokit/email"
	"github.com/cjcox17/sudokit/jobs"
	"github.com/cjcox17/sudokit/kernel"
	"github.com/cjcox17/sudokit/notifications"
	"github.com/cjcox17/sudokit/sms"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestNewTestKernel(t *testing.T) {
	tk := NewTestKernel(t)
	if tk == nil {
		t.Fatal("expected test kernel, got nil")
	}

	if tk.Kernel == nil {
		t.Error("expected kernel to be initialized")
	}

	if tk.MockJobs == nil {
		t.Error("expected mock jobs service to be initialized")
	}

	if tk.MockEvents == nil {
		t.Error("expected mock events service to be initialized")
	}

	if tk.MockNotify == nil {
		t.Error("expected mock notifications service to be initialized")
	}

	if tk.MockWebSocket == nil {
		t.Error("expected mock websocket hub to be initialized")
	}
}

func TestTestKernel_Context(t *testing.T) {
	tk := NewTestKernel(t)
	ctx := tk.Context()

	if ctx == nil {
		t.Fatal("expected context, got nil")
	}

	k := kernel.FromContext(ctx)
	if k == nil {
		t.Error("expected kernel from context, got nil")
	}
}

func TestJobsFromContext(t *testing.T) {
	tk := NewTestKernel(t)
	ctx := tk.Context()

	jobs := kernel.Jobs(ctx)
	if jobs == nil {
		t.Error("expected jobs service from context, got nil")
	}
}

func TestEventsFromContext(t *testing.T) {
	tk := NewTestKernel(t)
	ctx := tk.Context()

	events := kernel.Events(ctx)
	if events == nil {
		t.Error("expected events service from context, got nil")
	}
}

func TestNotifyFromContext(t *testing.T) {
	tk := NewTestKernel(t)
	ctx := tk.Context()

	notify := kernel.Notify(ctx)
	if notify == nil {
		t.Error("expected notifications service from context, got nil")
	}
}

func TestBroadcastFromContext(t *testing.T) {
	tk := NewTestKernel(t)
	ctx := tk.Context()

	hub := kernel.Broadcast(ctx)
	if hub == nil {
		t.Error("expected websocket hub from context, got nil")
	}
}

func TestMockJobs_Enqueue(t *testing.T) {
	tk := NewTestKernel(t)
	ctx := tk.Context()

	jobID, err := kernel.Jobs(ctx).Enqueue(ctx, "test_job", map[string]any{"foo": "bar"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if jobID == primitive.NilObjectID {
		t.Error("expected job ID, got nil")
	}

	if len(tk.MockJobs.EnqueueCalls) != 1 {
		t.Errorf("expected 1 enqueue call, got %d", len(tk.MockJobs.EnqueueCalls))
	}

	call := tk.MockJobs.EnqueueCalls[0]
	if call.JobType != "test_job" {
		t.Errorf("expected job type 'test_job', got '%s'", call.JobType)
	}
}

func TestMockEvents_Publish(t *testing.T) {
	tk := NewTestKernel(t)
	ctx := tk.Context()

	err := kernel.Events(ctx).Publish(ctx, "test.event", map[string]any{"foo": "bar"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(tk.MockEvents.PublishCalls) != 1 {
		t.Errorf("expected 1 publish call, got %d", len(tk.MockEvents.PublishCalls))
	}

	call := tk.MockEvents.PublishCalls[0]
	if call.EventType != "test.event" {
		t.Errorf("expected event type 'test.event', got '%s'", call.EventType)
	}
}

func TestMockNotify_Create(t *testing.T) {
	tk := NewTestKernel(t)
	ctx := tk.Context()

	userID := primitive.NewObjectID()
	orgID := primitive.NewObjectID()

	_, err := kernel.Notify(ctx).Create(ctx, userID, orgID, notifications.CreateInput{
		Type:     notifications.TypeInfo,
		Category: notifications.CategorySystem,
		Title:    "Test",
		Message:  "Test message",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(tk.MockNotify.CreateCalls) != 1 {
		t.Errorf("expected 1 create call, got %d", len(tk.MockNotify.CreateCalls))
	}

	call := tk.MockNotify.CreateCalls[0]
	if call.UserID != userID {
		t.Errorf("expected user ID '%s', got '%s'", userID.Hex(), call.UserID.Hex())
	}
}

func TestMockWebSocket_Broadcast(t *testing.T) {
	tk := NewTestKernel(t)
	ctx := tk.Context()

	userID := primitive.NewObjectID()
	kernel.Broadcast(ctx).BroadcastToUser(userID, "test.event", map[string]any{"foo": "bar"})

	if len(tk.MockWebSocket.BroadcastToUserCalls) != 1 {
		t.Errorf("expected 1 broadcast call, got %d", len(tk.MockWebSocket.BroadcastToUserCalls))
	}

	call := tk.MockWebSocket.BroadcastToUserCalls[0]
	if call.UserID != userID {
		t.Errorf("expected user ID '%s', got '%s'", userID.Hex(), call.UserID.Hex())
	}
}

func TestTestKernel_Reset(t *testing.T) {
	tk := NewTestKernel(t)
	ctx := tk.Context()

	userID := primitive.NewObjectID()
	orgID := primitive.NewObjectID()

	kernel.Jobs(ctx).Enqueue(ctx, "test", map[string]any{})
	kernel.Events(ctx).Publish(ctx, "test", nil)
	kernel.Notify(ctx).Create(ctx, userID, orgID, notifications.CreateInput{
		Title:   "Test",
		Message: "Test",
	})
	kernel.Broadcast(ctx).BroadcastToUser(userID, "event", nil)

	tk.Reset()

	if len(tk.MockJobs.EnqueueCalls) != 0 {
		t.Error("expected jobs calls to be reset")
	}

	if len(tk.MockEvents.PublishCalls) != 0 {
		t.Error("expected events calls to be reset")
	}

	if len(tk.MockNotify.CreateCalls) != 0 {
		t.Error("expected notify calls to be reset")
	}

	if len(tk.MockWebSocket.BroadcastToUserCalls) != 0 {
		t.Error("expected websocket calls to be reset")
	}
}

func TestAssertJobEnqueued(t *testing.T) {
	tk := NewTestKernel(t)
	ctx := tk.Context()

	kernel.Jobs(ctx).Enqueue(ctx, "expected_job", map[string]any{})

	AssertJobEnqueued(t, tk, "expected_job")
}

func TestAssertEventPublished(t *testing.T) {
	tk := NewTestKernel(t)
	ctx := tk.Context()

	kernel.Events(ctx).Publish(ctx, "expected.event", nil)

	AssertEventPublished(t, tk, "expected.event")
}

func TestAssertNotificationCreated(t *testing.T) {
	tk := NewTestKernel(t)
	ctx := tk.Context()

	userID := primitive.NewObjectID()
	orgID := primitive.NewObjectID()
	kernel.Notify(ctx).Create(ctx, userID, orgID, notifications.CreateInput{
		Title:   "Test",
		Message: "Test",
	})

	AssertNotificationCreated(t, tk, userID)
}

func TestAssertBroadcastToUser(t *testing.T) {
	tk := NewTestKernel(t)
	ctx := tk.Context()

	userID := primitive.NewObjectID()
	kernel.Broadcast(ctx).BroadcastToUser(userID, "event", nil)

	AssertBroadcastToUser(t, tk, userID)
}

func TestNewWithServices(t *testing.T) {
	mockJobs := NewMockJobsService()
	mockEvents := NewMockEventsService()

	services := &kernel.Services{
		Jobs:   mockJobs,
		Events: mockEvents,
	}

	k := kernel.NewWithServices(services)

	if k == nil {
		t.Fatal("expected kernel, got nil")
	}

	if k.Jobs() == nil {
		t.Error("expected Jobs() to return non-nil")
	}

	if k.Events() == nil {
		t.Error("expected Events() to return non-nil")
	}
}

func TestKernel_ServiceGetters(t *testing.T) {
	tk := NewTestKernel(t)

	if tk.Kernel.Jobs() == nil {
		t.Error("expected Jobs() to return non-nil")
	}

	if tk.Kernel.Events() == nil {
		t.Error("expected Events() to return non-nil")
	}

	if tk.Kernel.Notifications() == nil {
		t.Error("expected Notifications() to return non-nil")
	}

	if tk.Kernel.WebSocket() == nil {
		t.Error("expected WebSocket() to return non-nil")
	}

	if tk.Kernel.Storage() == nil {
		t.Error("expected Storage() to return non-nil")
	}

	if tk.Kernel.Email() == nil {
		t.Error("expected Email() to return non-nil")
	}

	if tk.Kernel.SMS() == nil {
		t.Error("expected SMS() to return non-nil")
	}

	if tk.Kernel.Cache() == nil {
		t.Error("expected Cache() to return non-nil")
	}
}

func TestMockJobs_CustomBehavior(t *testing.T) {
	tk := NewTestKernel(t)

	customID := primitive.NewObjectID()
	tk.MockJobs.EnqueueFunc = func(ctx context.Context, jobType string, payload map[string]any, opts ...jobs.Option) (primitive.ObjectID, error) {
		return customID, nil
	}

	ctx := tk.Context()
	jobID, _ := kernel.Jobs(ctx).Enqueue(ctx, "test", map[string]any{})

	if jobID != customID {
		t.Errorf("expected '%s', got '%s'", customID.Hex(), jobID.Hex())
	}
}

func TestMockJobs_ErrorBehavior(t *testing.T) {
	tk := NewTestKernel(t)

	expectedErr := context.DeadlineExceeded
	tk.MockJobs.EnqueueFunc = func(ctx context.Context, jobType string, payload map[string]any, opts ...jobs.Option) (primitive.ObjectID, error) {
		return primitive.NilObjectID, expectedErr
	}

	ctx := tk.Context()
	_, err := kernel.Jobs(ctx).Enqueue(ctx, "test", map[string]any{})

	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestJobOptions(t *testing.T) {
	opts := jobs.Options{
		Priority:    10,
		MaxAttempts: 5,
	}

	if opts.Priority != 10 {
		t.Errorf("expected priority 10, got %d", opts.Priority)
	}

	if opts.MaxAttempts != 5 {
		t.Errorf("expected max attempts 5, got %d", opts.MaxAttempts)
	}
}

func TestNotificationTypes(t *testing.T) {
	tests := []struct {
		name     string
		nt       notifications.Type
		expected string
	}{
		{"info", notifications.TypeInfo, "info"},
		{"warning", notifications.TypeWarning, "warning"},
		{"success", notifications.TypeSuccess, "success"},
		{"error", notifications.TypeError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.nt) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.nt)
			}
		})
	}
}

func TestNotificationCategories(t *testing.T) {
	tests := []struct {
		name     string
		nc       notifications.Category
		expected string
	}{
		{"system", notifications.CategorySystem, "system"},
		{"task", notifications.CategoryTask, "task"},
		{"mention", notifications.CategoryMention, "mention"},
		{"deadline", notifications.CategoryDeadline, "deadline"},
		{"esign", notifications.CategoryESign, "esign"},
		{"import", notifications.CategoryImport, "import"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.nc) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.nc)
			}
		})
	}
}

func TestJobStatuses(t *testing.T) {
	tests := []struct {
		name     string
		status   jobs.Status
		expected string
	}{
		{"pending", jobs.StatusPending, "pending"},
		{"running", jobs.StatusRunning, "running"},
		{"completed", jobs.StatusCompleted, "completed"},
		{"failed", jobs.StatusFailed, "failed"},
		{"cancelled", jobs.StatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.status)
			}
		})
	}
}

func TestMockStorage(t *testing.T) {
	mock := NewMockStorageService()
	ctx := context.Background()

	err := mock.Upload(ctx, "test-key", nil, "text/plain")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = mock.Download(ctx, "test-key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	url, err := mock.PresignedURL(ctx, "test-key", time.Hour)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if url == "" {
		t.Error("expected URL, got empty string")
	}

	exists, err := mock.Exists(ctx, "test-key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected exists to be true")
	}

	err = mock.Delete(ctx, "test-key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMockEmail(t *testing.T) {
	mock := NewMockEmailService()
	ctx := context.Background()

	_, err := mock.Send(ctx, &email.SendInput{
		To:       []string{"test@example.com"},
		Subject:  "Test Subject",
		HTMLBody: "Test Body",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = mock.SendTemplate(ctx, "welcome", []string{"test@example.com"}, map[string]any{"name": "Test"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMockSMS(t *testing.T) {
	mock := NewMockSMSService()
	ctx := context.Background()

	_, err := mock.Send(ctx, &sms.SendInput{
		To:      "+1234567890",
		Message: "Test message",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMockCache(t *testing.T) {
	mock := NewMockCacheService()
	ctx := context.Background()

	err := mock.Set(ctx, "test-key", []byte("test-value"), time.Hour)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	data, err := mock.Get(ctx, "test-key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if data == nil {
		t.Error("expected data, got nil")
	}

	exists, err := mock.Exists(ctx, "test-key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected exists to be true")
	}

	err = mock.Delete(ctx, "test-key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	data, err = mock.Remember(ctx, "new-key", time.Hour, func() ([]byte, error) {
		return []byte("computed-value"), nil
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(data) != "computed-value" {
		t.Errorf("expected 'computed-value', got '%s'", string(data))
	}
}
