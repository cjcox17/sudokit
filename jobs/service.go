package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Service struct {
	queue          *Queue
	registry       *Registry
	pool           *WorkerPool
	eventPublisher EventPublisher
	broadcaster    JobBroadcaster
	mu             sync.RWMutex
	started        bool
}

func NewService(db *mongo.Database, numWorkers int) *Service {
	queue := NewQueue(db)
	registry := NewRegistry()
	pool := NewWorkerPool(numWorkers, queue, registry)

	return &Service{
		queue:    queue,
		registry: registry,
		pool:     pool,
		started:  false,
	}
}

func NewServiceWithEvents(db *mongo.Database, numWorkers int, eventPublisher EventPublisher) *Service {
	s := NewService(db, numWorkers)
	s.eventPublisher = eventPublisher
	s.pool.SetEventPublisher(eventPublisher)
	return s
}

func (s *Service) SetEventPublisher(publisher EventPublisher) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.eventPublisher = publisher
	s.pool.SetEventPublisher(publisher)
}

func (s *Service) SetBroadcaster(broadcaster JobBroadcaster) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.broadcaster = broadcaster
}

func (s *Service) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return fmt.Errorf("service already started")
	}

	slog.Info("Jobs service: ensuring indexes...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.queue.EnsureIndexes(ctx); err != nil {
		return fmt.Errorf("failed to ensure indexes: %w", err)
	}
	slog.Info("Jobs service: indexes ensured")

	slog.Info("Jobs service: starting worker pool...")
	if err := s.pool.Start(); err != nil {
		return fmt.Errorf("failed to start worker pool: %w", err)
	}
	slog.Info("Jobs service: worker pool started")

	s.started = true
	slog.Info("Jobs service started")
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return nil
	}

	if err := s.pool.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop worker pool: %w", err)
	}

	s.started = false
	slog.Info("Jobs service stopped")
	return nil
}

func (s *Service) Enqueue(ctx context.Context, jobType string, payload map[string]any, opts ...Option) (primitive.ObjectID, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, err := s.queue.Enqueue(ctx, jobType, payload, opts...)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return job.ID, nil
}

func (s *Service) Get(ctx context.Context, jobID primitive.ObjectID) (*Job, error) {
	return s.queue.GetByID(ctx, jobID)
}

func (s *Service) GetByUserID(ctx context.Context, userID primitive.ObjectID, limit int) ([]*Job, error) {
	return s.queue.GetByUserID(ctx, userID, limit)
}

func (s *Service) GetByOrganizationID(ctx context.Context, orgID primitive.ObjectID, limit, skip int) ([]*Job, int, error) {
	return s.queue.GetByOrganizationID(ctx, orgID, limit, skip)
}

func (s *Service) Cancel(ctx context.Context, jobID primitive.ObjectID) error {
	job, _ := s.queue.GetByID(ctx, jobID)

	if err := s.queue.Cancel(ctx, jobID); err != nil {
		return err
	}

	if s.eventPublisher != nil && job != nil {
		job.Status = StatusCancelled
		_ = s.eventPublisher.PublishJobEvent(ctx, "job.cancelled", job)
	}

	return nil
}

func (s *Service) RegisterHandler(jobType string, handler Handler) {
	s.registry.Register(jobType, handler)
}

func (s *Service) RegisterHandlerFunc(jobType string, handler HandlerFunc) {
	s.registry.RegisterFunc(jobType, handler)
}

func (s *Service) UpdateProgress(ctx context.Context, jobID primitive.ObjectID, progress int, message string) error {
	err := s.queue.UpdateProgress(ctx, jobID, progress, message)
	if err != nil {
		return err
	}

	if s.eventPublisher != nil {
		job, _ := s.queue.GetByID(ctx, jobID)
		if job != nil {
			job.Progress = progress
			job.Message = message
			_ = s.eventPublisher.PublishJobEvent(ctx, "job.progress", job)
		}
	}

	return nil
}

func (s *Service) BroadcastUpdate(ctx context.Context, jobID primitive.ObjectID, progress int, message string) error {
	err := s.queue.UpdateProgress(ctx, jobID, progress, message)
	if err != nil {
		return err
	}

	if s.broadcaster != nil {
		job, _ := s.queue.GetByID(ctx, jobID)
		if job != nil {
			job.Progress = progress
			job.Message = message
			_ = s.broadcaster.BroadcastJobUpdate(ctx, job.UserID, job)
		}
	}

	return nil
}

func (s *Service) KeepAlive(ctx context.Context, jobID primitive.ObjectID) error {
	return s.queue.KeepAlive(ctx, jobID)
}

func (s *Service) IsStarted() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.started
}

func (s *Service) RegisteredTypes() []string {
	return s.registry.RegisteredTypes()
}

var (
	globalService *Service
	serviceOnce   sync.Once
)

func InitService(db *mongo.Database, numWorkers int) *Service {
	serviceOnce.Do(func() {
		globalService = NewService(db, numWorkers)
	})
	return globalService
}

func GetService() *Service {
	return globalService
}
