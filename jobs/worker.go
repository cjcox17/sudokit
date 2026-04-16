package jobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type Worker struct {
	id             string
	queue          *Queue
	registry       *Registry
	eventPublisher EventPublisher
	stopCh         chan struct{}
	wg             sync.WaitGroup
}

type WorkerPool struct {
	workers        []*Worker
	queue          *Queue
	registry       *Registry
	eventPublisher EventPublisher
	stopCh         chan struct{}
	wg             sync.WaitGroup
	mu             sync.RWMutex
	started        bool
}

func NewWorkerPool(numWorkers int, queue *Queue, registry *Registry) *WorkerPool {
	workers := make([]*Worker, numWorkers)
	for i := 0; i < numWorkers; i++ {
		workers[i] = &Worker{
			id:       fmt.Sprintf("worker-%d", i),
			queue:    queue,
			registry: registry,
			stopCh:   make(chan struct{}),
		}
	}

	return &WorkerPool{
		workers:  workers,
		queue:    queue,
		registry: registry,
		stopCh:   make(chan struct{}),
		started:  false,
	}
}

func (p *WorkerPool) SetEventPublisher(publisher EventPublisher) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.eventPublisher = publisher
	for _, worker := range p.workers {
		worker.eventPublisher = publisher
	}
}

func (p *WorkerPool) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.started {
		return fmt.Errorf("worker pool already started")
	}

	if err := p.registry.Validate(); err != nil {
		return fmt.Errorf("registry validation failed: %w", err)
	}

	p.started = true

	for _, worker := range p.workers {
		p.wg.Add(1)
		go worker.run(&p.wg, p.stopCh)
	}

	p.wg.Add(1)
	go p.runMaintenance()

	slog.Info("Worker pool started", "workers", len(p.workers))
	return nil
}

func (p *WorkerPool) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.started {
		return nil
	}

	slog.Info("Stopping worker pool...")

	close(p.stopCh)

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		p.started = false
		slog.Info("Worker pool stopped")
		return nil
	case <-ctx.Done():
		p.started = false
		slog.Warn("Worker pool stop timed out")
		return ctx.Err()
	}
}

func (p *WorkerPool) runMaintenance() {
	defer p.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopCh:
			return
		case <-ticker.C:
			count, err := p.queue.ReleaseStaleLocks(context.Background(), 5*time.Minute)
			if err != nil {
				slog.Error("Failed to release stale locks", "error", err)
			} else if count > 0 {
				slog.Info("Released stale job locks", "count", count)
			}
		}
	}
}

func (w *Worker) run(wg *sync.WaitGroup, stopCh chan struct{}) {
	defer wg.Done()

	pollInterval := 2 * time.Second
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			job, err := w.queue.Dequeue(context.Background(), w.id)
			if err != nil {
				slog.Error("Failed to dequeue job", "worker", w.id, "error", err)
				continue
			}

			if job == nil {
				select {
				case <-stopCh:
					return
				case <-time.After(pollInterval):
					continue
				}
			}

			w.processJob(job)
		}
	}
}

func (w *Worker) processJob(job *Job) {
	timeout := 30 * time.Minute
	if job.Timeout > 0 {
		timeout = job.Timeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	w.queue.RegisterRunning(job.ID, cancel)
	defer w.queue.UnregisterRunning(job.ID)

	slog.Debug("Processing job", "worker", w.id, "job_id", job.ID, "type", job.JobType, "timeout", timeout)

	if w.eventPublisher != nil {
		_ = w.eventPublisher.PublishJobEvent(ctx, "job.started", job)
	}

	handler, ok := w.registry.GetHandler(job.JobType)
	if !ok {
		err := fmt.Errorf("no handler registered for job type: %s", job.JobType)
		slog.Error("Handler not found", "job_id", job.ID, "type", job.JobType)
		w.queue.Fail(ctx, job.ID, err.Error())
		if w.eventPublisher != nil {
			job.Error = err.Error()
			_ = w.eventPublisher.PublishJobEvent(ctx, "job.failed", job)
		}
		return
	}

	err := handler.Handle(ctx, job)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			slog.Info("Job cancelled", "job_id", job.ID, "type", job.JobType, "error", err)
			w.queue.Cancel(ctx, job.ID)
			if w.eventPublisher != nil {
				job.Status = StatusCancelled
				_ = w.eventPublisher.PublishJobEvent(ctx, "job.cancelled", job)
			}
			return
		}
		slog.Error("Job failed", "job_id", job.ID, "type", job.JobType, "error", err)
		w.queue.Fail(ctx, job.ID, err.Error())
		if w.eventPublisher != nil {
			job.Error = err.Error()
			_ = w.eventPublisher.PublishJobEvent(ctx, "job.failed", job)
		}
		return
	}

	if err := w.queue.Acknowledge(ctx, job.ID, job.Result); err != nil {
		slog.Error("Failed to acknowledge job", "job_id", job.ID, "error", err)
		return
	}
	slog.Debug("Job completed", "job_id", job.ID, "type", job.JobType)

	if w.eventPublisher != nil {
		_ = w.eventPublisher.PublishJobEvent(ctx, "job.completed", job)
	}
}

type JobWithProgress struct {
	*Job
	Queue *Queue
}

func (j *JobWithProgress) ReportProgress(ctx context.Context, progress int, message string) error {
	return j.Queue.UpdateProgress(ctx, j.ID, progress, message)
}
