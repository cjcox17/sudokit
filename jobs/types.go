package jobs

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusRunning, StatusCompleted, StatusFailed, StatusCancelled:
		return true
	default:
		return false
	}
}

type Job struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrganizationID primitive.ObjectID `bson:"organization_id,omitempty" json:"organization_id,omitempty"`
	UserID         primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
	JobType        string             `bson:"job_type" json:"job_type"`
	Payload        map[string]any     `bson:"payload" json:"payload"`
	Status         Status             `bson:"status" json:"status"`
	Priority       int                `bson:"priority" json:"priority"`
	Attempts       int                `bson:"attempts" json:"attempts"`
	MaxAttempts    int                `bson:"max_attempts" json:"max_attempts"`
	Progress       int                `bson:"progress" json:"progress"`
	Message        string             `bson:"message,omitempty" json:"message,omitempty"`
	Result         map[string]any     `bson:"result,omitempty" json:"result,omitempty"`
	Error          string             `bson:"error,omitempty" json:"error,omitempty"`
	LockedAt       *time.Time         `bson:"locked_at,omitempty" json:"locked_at,omitempty"`
	LockedBy       string             `bson:"locked_by,omitempty" json:"locked_by,omitempty"`
	RunAt          *time.Time         `bson:"run_at,omitempty" json:"run_at,omitempty"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	StartedAt      *time.Time         `bson:"started_at,omitempty" json:"started_at,omitempty"`
	CompletedAt    *time.Time         `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
	LongRunning    bool               `bson:"long_running" json:"long_running"`
	Timeout        time.Duration      `bson:"timeout" json:"timeout"`
}

type Options struct {
	Priority    int
	MaxAttempts int
	Delay       time.Duration
	RunAt       *time.Time
	UserID      primitive.ObjectID
	OrgID       primitive.ObjectID
	LongRunning bool
	Timeout     time.Duration
}

type Option func(*Options)

func WithPriority(priority int) Option {
	return func(o *Options) {
		o.Priority = priority
	}
}

func WithMaxAttempts(maxAttempts int) Option {
	return func(o *Options) {
		o.MaxAttempts = maxAttempts
	}
}

func WithDelay(delay time.Duration) Option {
	return func(o *Options) {
		o.Delay = delay
	}
}

func WithRunAt(runAt time.Time) Option {
	return func(o *Options) {
		o.RunAt = &runAt
	}
}

func WithUser(userID primitive.ObjectID) Option {
	return func(o *Options) {
		o.UserID = userID
	}
}

func WithOrganization(orgID primitive.ObjectID) Option {
	return func(o *Options) {
		o.OrgID = orgID
	}
}

func WithLongRunning() Option {
	return func(o *Options) {
		o.LongRunning = true
		o.MaxAttempts = 1
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.Timeout = timeout
	}
}

type Handler interface {
	Handle(ctx context.Context, job *Job) error
}

type HandlerFunc func(ctx context.Context, job *Job) error

func (f HandlerFunc) Handle(ctx context.Context, job *Job) error {
	return f(ctx, job)
}

type ProgressReporter interface {
	ReportProgress(ctx context.Context, jobID primitive.ObjectID, progress int, message string) error
}

type EventPublisher interface {
	PublishJobEvent(ctx context.Context, eventType string, job *Job) error
}

type JobBroadcaster interface {
	BroadcastJobUpdate(ctx context.Context, userID primitive.ObjectID, job *Job) error
}

type ServiceInterface interface {
	Enqueue(ctx context.Context, jobType string, payload map[string]any, opts ...Option) (primitive.ObjectID, error)
	Get(ctx context.Context, jobID primitive.ObjectID) (*Job, error)
	GetByUserID(ctx context.Context, userID primitive.ObjectID, limit int) ([]*Job, error)
	GetByOrganizationID(ctx context.Context, orgID primitive.ObjectID, limit, skip int) ([]*Job, int, error)
	Cancel(ctx context.Context, jobID primitive.ObjectID) error
	RegisterHandler(jobType string, handler Handler)
	RegisterHandlerFunc(jobType string, handler HandlerFunc)
	UpdateProgress(ctx context.Context, jobID primitive.ObjectID, progress int, message string) error
	BroadcastUpdate(ctx context.Context, jobID primitive.ObjectID, progress int, message string) error
	KeepAlive(ctx context.Context, jobID primitive.ObjectID) error
	Start() error
	Stop(ctx context.Context) error
	IsStarted() bool
	RegisteredTypes() []string
}
