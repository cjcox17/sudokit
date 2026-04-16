package testing

import (
	"context"
	"testing"
	"time"

	"github.com/cjcox17/sudokit/jobs"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const testJobTypeCSVImport = "csv_import"

func TestMockJobsService_Enqueue(t *testing.T) {
	mock := NewMockJobsService()
	ctx := context.Background()

	jobID, err := mock.Enqueue(ctx, testJobTypeCSVImport, map[string]any{
		"file_key": "test.csv",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if jobID == primitive.NilObjectID {
		t.Error("expected job ID, got nil")
	}

	if len(mock.EnqueueCalls) != 1 {
		t.Errorf("expected 1 enqueue call, got %d", len(mock.EnqueueCalls))
	}

	call := mock.EnqueueCalls[0]
	if call.JobType != testJobTypeCSVImport {
		t.Errorf("expected job type '%s', got '%s'", testJobTypeCSVImport, call.JobType)
	}
}

func TestMockJobsService_EnqueueWithOptions(t *testing.T) {
	mock := NewMockJobsService()
	ctx := context.Background()
	userID := primitive.NewObjectID()
	orgID := primitive.NewObjectID()

	jobID, err := mock.Enqueue(ctx, testJobTypeCSVImport, map[string]any{},
		jobs.WithPriority(10),
		jobs.WithMaxAttempts(5),
		jobs.WithUser(userID),
		jobs.WithOrganization(orgID),
	)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if jobID == primitive.NilObjectID {
		t.Error("expected job ID, got nil")
	}

	if len(mock.EnqueueCalls) != 1 {
		t.Errorf("expected 1 enqueue call, got %d", len(mock.EnqueueCalls))
	}
}

func TestMockJobsService_Get(t *testing.T) {
	mock := NewMockJobsService()
	ctx := context.Background()
	jobID := primitive.NewObjectID()

	job, err := mock.Get(ctx, jobID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if job == nil {
		t.Error("expected job, got nil")
	}

	if job.ID != jobID {
		t.Error("expected job ID to match")
	}
}

func TestMockJobsService_GetByUserID(t *testing.T) {
	mock := NewMockJobsService()
	ctx := context.Background()
	userID := primitive.NewObjectID()

	_, err := mock.GetByUserID(ctx, userID, 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(mock.GetByUserIDCalls) != 1 {
		t.Errorf("expected 1 GetByUserID call, got %d", len(mock.GetByUserIDCalls))
	}

	call := mock.GetByUserIDCalls[0]
	if call.UserID != userID {
		t.Error("expected user ID to match")
	}

	if call.Limit != 10 {
		t.Errorf("expected limit 10, got %d", call.Limit)
	}
}

func TestMockJobsService_Cancel(t *testing.T) {
	mock := NewMockJobsService()
	ctx := context.Background()
	jobID := primitive.NewObjectID()

	err := mock.Cancel(ctx, jobID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(mock.CancelCalls) != 1 {
		t.Errorf("expected 1 cancel call, got %d", len(mock.CancelCalls))
	}

	call := mock.CancelCalls[0]
	if call.JobID != jobID {
		t.Error("expected job ID to match")
	}
}

func TestMockJobsService_RegisterHandler(t *testing.T) {
	mock := NewMockJobsService()

	handler := jobs.HandlerFunc(func(ctx context.Context, job *jobs.Job) error {
		return nil
	})

	mock.RegisterHandler(testJobTypeCSVImport, handler)

	if len(mock.RegisteredHandlers) != 1 {
		t.Errorf("expected 1 registered handler, got %d", len(mock.RegisteredHandlers))
	}

	if _, ok := mock.RegisteredHandlers[testJobTypeCSVImport]; !ok {
		t.Error("expected handler to be registered")
	}
}

func TestMockJobsService_UpdateProgress(t *testing.T) {
	mock := NewMockJobsService()
	ctx := context.Background()
	jobID := primitive.NewObjectID()

	err := mock.UpdateProgress(ctx, jobID, 50, "Processing...")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(mock.UpdateProgressCalls) != 1 {
		t.Errorf("expected 1 update progress call, got %d", len(mock.UpdateProgressCalls))
	}

	call := mock.UpdateProgressCalls[0]
	if call.JobID != jobID {
		t.Error("expected job ID to match")
	}

	if call.Progress != 50 {
		t.Errorf("expected progress 50, got %d", call.Progress)
	}

	if call.Message != "Processing..." {
		t.Errorf("expected message 'Processing...', got '%s'", call.Message)
	}
}

func TestMockJobsService_Reset(t *testing.T) {
	mock := NewMockJobsService()
	ctx := context.Background()

	mock.Enqueue(ctx, "test", nil)
	mock.Get(ctx, primitive.NewObjectID())
	mock.Cancel(ctx, primitive.NewObjectID())

	mock.Reset()

	if len(mock.EnqueueCalls) != 0 {
		t.Error("expected enqueue calls to be reset")
	}

	if len(mock.GetCalls) != 0 {
		t.Error("expected get calls to be reset")
	}

	if len(mock.CancelCalls) != 0 {
		t.Error("expected cancel calls to be reset")
	}
}

func TestMockJobsService_GetLastEnqueuedJob(t *testing.T) {
	mock := NewMockJobsService()
	ctx := context.Background()

	mock.Enqueue(ctx, "job1", map[string]any{"key": "value1"})
	mock.Enqueue(ctx, "job2", map[string]any{"key": "value2"})

	jobType, payload := mock.GetLastEnqueuedJob()

	if jobType != "job2" {
		t.Errorf("expected job type 'job2', got '%s'", jobType)
	}

	if payload["key"] != "value2" {
		t.Errorf("expected payload key 'value2', got '%v'", payload["key"])
	}
}

func TestMockJobsService_GetEnqueueCallsByType(t *testing.T) {
	mock := NewMockJobsService()
	ctx := context.Background()

	mock.Enqueue(ctx, "job1", nil)
	mock.Enqueue(ctx, "job2", nil)
	mock.Enqueue(ctx, "job1", nil)

	calls := mock.GetEnqueueCallsByType("job1")

	if len(calls) != 2 {
		t.Errorf("expected 2 calls for job1, got %d", len(calls))
	}
}

func TestAssertJobEnqueued(t *testing.T) {
	mock := NewMockJobsService()
	ctx := context.Background()

	mock.Enqueue(ctx, testJobTypeCSVImport, nil)

	AssertJobEnqueued(t, mock, testJobTypeCSVImport)
}

func TestAssertJobCancelled(t *testing.T) {
	mock := NewMockJobsService()
	ctx := context.Background()
	jobID := primitive.NewObjectID()

	mock.Cancel(ctx, jobID)

	AssertJobCancelled(t, mock, jobID)
}

func TestCreateTestJob(t *testing.T) {
	job := CreateTestJob(testJobTypeCSVImport, jobs.StatusPending)

	if job.JobType != testJobTypeCSVImport {
		t.Errorf("expected job type '%s', got '%s'", testJobTypeCSVImport, job.JobType)
	}

	if job.Status != jobs.StatusPending {
		t.Errorf("expected status '%s', got '%s'", jobs.StatusPending, job.Status)
	}

	if job.ID == primitive.NilObjectID {
		t.Error("expected job ID to be set")
	}
}

func TestCreateTestJobWithID(t *testing.T) {
	jobID := primitive.NewObjectID()
	job := CreateTestJobWithID(jobID, testJobTypeCSVImport, jobs.StatusRunning)

	if job.ID != jobID {
		t.Error("expected job ID to match")
	}

	if job.Status != jobs.StatusRunning {
		t.Errorf("expected status '%s', got '%s'", jobs.StatusRunning, job.Status)
	}
}

func TestJobStatus_IsValid(t *testing.T) {
	tests := []struct {
		status   jobs.Status
		expected bool
	}{
		{jobs.StatusPending, true},
		{jobs.StatusRunning, true},
		{jobs.StatusCompleted, true},
		{jobs.StatusFailed, true},
		{jobs.StatusCancelled, true},
		{jobs.Status("invalid"), false},
		{jobs.Status(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if tt.status.IsValid() != tt.expected {
				t.Errorf("expected IsValid() = %v, got %v", tt.expected, tt.status.IsValid())
			}
		})
	}
}

func TestJobOptions(t *testing.T) {
	opts := &jobs.Options{}

	jobs.WithPriority(10)(opts)
	jobs.WithMaxAttempts(5)(opts)
	jobs.WithDelay(time.Minute)(opts)
	jobs.WithUser(primitive.NewObjectID())(opts)
	jobs.WithOrganization(primitive.NewObjectID())(opts)

	if opts.Priority != 10 {
		t.Errorf("expected priority 10, got %d", opts.Priority)
	}

	if opts.MaxAttempts != 5 {
		t.Errorf("expected max attempts 5, got %d", opts.MaxAttempts)
	}

	if opts.Delay != time.Minute {
		t.Errorf("expected delay 1 minute, got %v", opts.Delay)
	}

	if opts.UserID == primitive.NilObjectID {
		t.Error("expected user ID to be set")
	}

	if opts.OrgID == primitive.NilObjectID {
		t.Error("expected org ID to be set")
	}
}

func TestHandlerFunc(t *testing.T) {
	called := false
	handler := jobs.HandlerFunc(func(ctx context.Context, job *jobs.Job) error {
		called = true
		return nil
	})

	job := CreateTestJob("test", jobs.StatusPending)
	err := handler.Handle(context.Background(), job)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !called {
		t.Error("expected handler to be called")
	}
}

func TestMockJobsService_CustomBehavior(t *testing.T) {
	mock := NewMockJobsService()

	expectedID := primitive.NewObjectID()
	mock.EnqueueFunc = func(ctx context.Context, jobType string, payload map[string]any, opts ...jobs.Option) (primitive.ObjectID, error) {
		return expectedID, nil
	}

	ctx := context.Background()
	jobID, _ := mock.Enqueue(ctx, "test", nil)

	if jobID != expectedID {
		t.Errorf("expected job ID %s, got %s", expectedID.Hex(), jobID.Hex())
	}
}

func TestMockJobsService_ErrorBehavior(t *testing.T) {
	mock := NewMockJobsService()

	expectedErr := context.DeadlineExceeded
	mock.EnqueueFunc = func(ctx context.Context, jobType string, payload map[string]any, opts ...jobs.Option) (primitive.ObjectID, error) {
		return primitive.NilObjectID, expectedErr
	}

	ctx := context.Background()
	_, err := mock.Enqueue(ctx, "test", nil)

	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestMockJobsService_RegisteredTypes(t *testing.T) {
	mock := NewMockJobsService()

	mock.RegisterHandler("job1", jobs.HandlerFunc(func(ctx context.Context, job *jobs.Job) error { return nil }))
	mock.RegisterHandler("job2", jobs.HandlerFunc(func(ctx context.Context, job *jobs.Job) error { return nil }))

	types := mock.RegisteredTypes()

	if len(types) != 2 {
		t.Errorf("expected 2 registered types, got %d", len(types))
	}
}
