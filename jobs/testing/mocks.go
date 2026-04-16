package testing

import (
	"context"
	"sync"
	"time"

	"github.com/cjcox17/sudokit/jobs"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MockJobsService struct {
	mu sync.RWMutex

	EnqueueFunc         func(ctx context.Context, jobType string, payload map[string]any, opts ...jobs.Option) (primitive.ObjectID, error)
	GetFunc             func(ctx context.Context, jobID primitive.ObjectID) (*jobs.Job, error)
	GetByUserIDFunc     func(ctx context.Context, userID primitive.ObjectID, limit int) ([]*jobs.Job, error)
	CancelFunc          func(ctx context.Context, jobID primitive.ObjectID) error
	OnRegisterHandler   func(jobType string, handler jobs.Handler)
	StartFunc           func() error
	StopFunc            func(ctx context.Context) error
	UpdateProgressFunc  func(ctx context.Context, jobID primitive.ObjectID, progress int, message string) error
	BroadcastUpdateFunc func(ctx context.Context, jobID primitive.ObjectID, progress int, message string) error
	KeepAliveFunc       func(ctx context.Context, jobID primitive.ObjectID) error

	EnqueueCalls         []EnqueueCall
	GetCalls             []GetCall
	GetByUserIDCalls     []GetByUserIDCall
	CancelCalls          []CancelCall
	UpdateProgressCalls  []UpdateProgressCall
	BroadcastUpdateCalls []BroadcastUpdateCall
	KeepAliveCalls       []KeepAliveCall
	RegisteredHandlers   map[string]jobs.Handler
}

type EnqueueCall struct {
	JobType string
	Payload map[string]any
	Options []jobs.Option
}

type GetCall struct {
	JobID primitive.ObjectID
}

type GetByUserIDCall struct {
	UserID primitive.ObjectID
	Limit  int
}

type CancelCall struct {
	JobID primitive.ObjectID
}

type UpdateProgressCall struct {
	JobID    primitive.ObjectID
	Progress int
	Message  string
}

type BroadcastUpdateCall struct {
	JobID    primitive.ObjectID
	Progress int
	Message  string
}

type KeepAliveCall struct {
	JobID primitive.ObjectID
}

func NewMockJobsService() *MockJobsService {
	return &MockJobsService{
		EnqueueCalls:         make([]EnqueueCall, 0),
		GetCalls:             make([]GetCall, 0),
		GetByUserIDCalls:     make([]GetByUserIDCall, 0),
		CancelCalls:          make([]CancelCall, 0),
		UpdateProgressCalls:  make([]UpdateProgressCall, 0),
		BroadcastUpdateCalls: make([]BroadcastUpdateCall, 0),
		KeepAliveCalls:       make([]KeepAliveCall, 0),
		RegisteredHandlers:   make(map[string]jobs.Handler),
	}
}

func (m *MockJobsService) Enqueue(ctx context.Context, jobType string, payload map[string]any, opts ...jobs.Option) (primitive.ObjectID, error) {
	m.mu.Lock()
	m.EnqueueCalls = append(m.EnqueueCalls, EnqueueCall{
		JobType: jobType,
		Payload: payload,
		Options: opts,
	})
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
	return &jobs.Job{
		ID:     jobID,
		Status: jobs.StatusPending,
	}, nil
}

func (m *MockJobsService) GetByUserID(ctx context.Context, userID primitive.ObjectID, limit int) ([]*jobs.Job, error) {
	m.mu.Lock()
	m.GetByUserIDCalls = append(m.GetByUserIDCalls, GetByUserIDCall{UserID: userID, Limit: limit})
	m.mu.Unlock()

	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID, limit)
	}
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

func (m *MockJobsService) RegisterHandler(jobType string, handler jobs.Handler) {
	m.mu.Lock()
	m.RegisteredHandlers[jobType] = handler
	m.mu.Unlock()

	if m.OnRegisterHandler != nil {
		m.OnRegisterHandler(jobType, handler)
	}
}

func (m *MockJobsService) RegisterHandlerFunc(jobType string, handler jobs.HandlerFunc) {
	m.RegisterHandler(jobType, handler)
}

func (m *MockJobsService) UpdateProgress(ctx context.Context, jobID primitive.ObjectID, progress int, message string) error {
	m.mu.Lock()
	m.UpdateProgressCalls = append(m.UpdateProgressCalls, UpdateProgressCall{
		JobID:    jobID,
		Progress: progress,
		Message:  message,
	})
	m.mu.Unlock()

	if m.UpdateProgressFunc != nil {
		return m.UpdateProgressFunc(ctx, jobID, progress, message)
	}
	return nil
}

func (m *MockJobsService) BroadcastUpdate(ctx context.Context, jobID primitive.ObjectID, progress int, message string) error {
	m.mu.Lock()
	m.BroadcastUpdateCalls = append(m.BroadcastUpdateCalls, BroadcastUpdateCall{
		JobID:    jobID,
		Progress: progress,
		Message:  message,
	})
	m.mu.Unlock()

	if m.BroadcastUpdateFunc != nil {
		return m.BroadcastUpdateFunc(ctx, jobID, progress, message)
	}
	return nil
}

func (m *MockJobsService) KeepAlive(ctx context.Context, jobID primitive.ObjectID) error {
	m.mu.Lock()
	m.KeepAliveCalls = append(m.KeepAliveCalls, KeepAliveCall{
		JobID: jobID,
	})
	m.mu.Unlock()

	if m.KeepAliveFunc != nil {
		return m.KeepAliveFunc(ctx, jobID)
	}
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

func (m *MockJobsService) IsStarted() bool {
	return true
}

func (m *MockJobsService) RegisteredTypes() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	types := make([]string, 0, len(m.RegisteredHandlers))
	for t := range m.RegisteredHandlers {
		types = append(types, t)
	}
	return types
}

func (m *MockJobsService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.EnqueueCalls = make([]EnqueueCall, 0)
	m.GetCalls = make([]GetCall, 0)
	m.GetByUserIDCalls = make([]GetByUserIDCall, 0)
	m.CancelCalls = make([]CancelCall, 0)
	m.UpdateProgressCalls = make([]UpdateProgressCall, 0)
	m.BroadcastUpdateCalls = make([]BroadcastUpdateCall, 0)
	m.KeepAliveCalls = make([]KeepAliveCall, 0)
}

func (m *MockJobsService) GetLastEnqueuedJob() (jobType string, payload map[string]any) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.EnqueueCalls) == 0 {
		return "", nil
	}

	call := m.EnqueueCalls[len(m.EnqueueCalls)-1]
	return call.JobType, call.Payload
}

func (m *MockJobsService) GetEnqueueCallsByType(jobType string) []EnqueueCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []EnqueueCall
	for _, call := range m.EnqueueCalls {
		if call.JobType == jobType {
			result = append(result, call)
		}
	}
	return result
}

func AssertJobEnqueued(t interface{ Errorf(string, ...any) }, mock *MockJobsService, expectedJobType string) {
	found := false
	for _, call := range mock.EnqueueCalls {
		if call.JobType == expectedJobType {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected job of type '%s' to be enqueued, but it was not", expectedJobType)
	}
}

func AssertJobCancelled(t interface{ Errorf(string, ...any) }, mock *MockJobsService, jobID primitive.ObjectID) {
	found := false
	for _, call := range mock.CancelCalls {
		if call.JobID == jobID {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected job '%s' to be cancelled, but it was not", jobID.Hex())
	}
}

func CreateTestJob(jobType string, status jobs.Status) *jobs.Job {
	return &jobs.Job{
		ID:          primitive.NewObjectID(),
		JobType:     jobType,
		Status:      status,
		Priority:    0,
		Attempts:    0,
		MaxAttempts: 3,
		Progress:    0,
		CreatedAt:   time.Now(),
	}
}

func CreateTestJobWithID(id primitive.ObjectID, jobType string, status jobs.Status) *jobs.Job {
	return &jobs.Job{
		ID:          id,
		JobType:     jobType,
		Status:      status,
		Priority:    0,
		Attempts:    0,
		MaxAttempts: 3,
		Progress:    0,
		CreatedAt:   time.Now(),
	}
}
