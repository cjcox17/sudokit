package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Queue struct {
	db         *mongo.Database
	collection *mongo.Collection
	mu         sync.RWMutex
	running    map[primitive.ObjectID]context.CancelFunc
	runningMu  sync.RWMutex
}

func NewQueue(db *mongo.Database) *Queue {
	return &Queue{
		db:         db,
		collection: db.Collection("jobs"),
		running:    make(map[primitive.ObjectID]context.CancelFunc),
	}
}

func (q *Queue) RegisterRunning(jobID primitive.ObjectID, cancel context.CancelFunc) {
	q.runningMu.Lock()
	defer q.runningMu.Unlock()
	q.running[jobID] = cancel
}

func (q *Queue) UnregisterRunning(jobID primitive.ObjectID) {
	q.runningMu.Lock()
	defer q.runningMu.Unlock()
	delete(q.running, jobID)
}

func (q *Queue) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "priority", Value: -1},
				{Key: "created_at", Value: 1},
			},
		},
		{
			Keys:    bson.D{{Key: "locked_at", Value: 1}},
			Options: options.Index().SetSparse(true),
		},
		{
			Keys: bson.D{{Key: "organization_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "run_at", Value: 1}},
		},
	}

	_, err := q.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	slog.Info("Job queue indexes created")
	return nil
}

func (q *Queue) Enqueue(ctx context.Context, jobType string, payload map[string]any, opts ...Option) (*Job, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	options := &Options{
		Priority:    0,
		MaxAttempts: 3,
	}

	for _, opt := range opts {
		opt(options)
	}

	now := time.Now()
	job := &Job{
		ID:          primitive.NewObjectID(),
		JobType:     jobType,
		Payload:     payload,
		Status:      StatusPending,
		Priority:    options.Priority,
		Attempts:    0,
		MaxAttempts: options.MaxAttempts,
		Progress:    0,
		CreatedAt:   now,
		LongRunning: options.LongRunning,
		Timeout:     options.Timeout,
	}

	if options.UserID != primitive.NilObjectID {
		job.UserID = options.UserID
	}

	if options.OrgID != primitive.NilObjectID {
		job.OrganizationID = options.OrgID
	}

	if options.Delay > 0 {
		runAt := now.Add(options.Delay)
		job.RunAt = &runAt
	} else if options.RunAt != nil {
		job.RunAt = options.RunAt
	}

	_, err := q.collection.InsertOne(ctx, job)
	if err != nil {
		return nil, fmt.Errorf("failed to insert job: %w", err)
	}

	slog.Debug("Job enqueued", "id", job.ID, "type", jobType, "long_running", job.LongRunning)
	return job, nil
}

func (q *Queue) Dequeue(ctx context.Context, workerID string) (*Job, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now()

	filter := bson.M{
		"status": StatusPending,
		"$or": []bson.M{
			{"run_at": bson.M{"$exists": false}},
			{"run_at": nil},
			{"run_at": bson.M{"$lte": now}},
		},
	}

	update := bson.M{
		"$set": bson.M{
			"status":     StatusRunning,
			"locked_at":  now,
			"locked_by":  workerID,
			"started_at": now,
		},
		"$inc": bson.M{"attempts": 1},
	}

	findOptions := options.FindOneAndUpdate().
		SetSort(bson.D{
			{Key: "priority", Value: -1},
			{Key: "created_at", Value: 1},
		}).
		SetReturnDocument(options.After)

	var job Job
	err := q.collection.FindOneAndUpdate(ctx, filter, update, findOptions).Decode(&job)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	return &job, nil
}

func (q *Queue) Acknowledge(ctx context.Context, jobID primitive.ObjectID, result map[string]any) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":       StatusCompleted,
			"completed_at": now,
			"progress":     100,
			"result":       result,
			"locked_at":    nil,
			"locked_by":    "",
		},
	}

	_, err := q.collection.UpdateByID(ctx, jobID, update)
	if err != nil {
		return fmt.Errorf("failed to acknowledge job: %w", err)
	}

	slog.Debug("Job acknowledged", "id", jobID)
	return nil
}

func (q *Queue) Fail(ctx context.Context, jobID primitive.ObjectID, errMsg string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	var job Job
	err := q.collection.FindOne(ctx, bson.M{"_id": jobID}).Decode(&job)
	if err != nil {
		return fmt.Errorf("failed to find job: %w", err)
	}

	now := time.Now()

	if job.Attempts >= job.MaxAttempts {
		update := bson.M{
			"$set": bson.M{
				"status":       StatusFailed,
				"error":        errMsg,
				"completed_at": now,
				"locked_at":    nil,
				"locked_by":    "",
			},
		}
		_, err = q.collection.UpdateByID(ctx, jobID, update)
		if err != nil {
			return fmt.Errorf("failed to mark job as failed: %w", err)
		}
		slog.Debug("Job permanently failed", "id", jobID, "attempts", job.Attempts)
	} else {
		backoff := q.calculateBackoff(job.Attempts)
		runAt := now.Add(backoff)

		update := bson.M{
			"$set": bson.M{
				"status":    StatusPending,
				"error":     errMsg,
				"run_at":    runAt,
				"locked_at": nil,
				"locked_by": "",
			},
		}
		_, err = q.collection.UpdateByID(ctx, jobID, update)
		if err != nil {
			return fmt.Errorf("failed to requeue job: %w", err)
		}
		slog.Debug("Job requeued for retry", "id", jobID, "attempt", job.Attempts, "backoff", backoff)
	}

	return nil
}

func (q *Queue) calculateBackoff(attempts int) time.Duration {
	backoff := time.Duration(1<<uint(attempts)) * time.Second
	maxBackoff := 5 * time.Minute
	if backoff > maxBackoff {
		backoff = maxBackoff
	}
	return backoff
}

func (q *Queue) UpdateProgress(ctx context.Context, jobID primitive.ObjectID, progress int, message string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}

	update := bson.M{
		"$set": bson.M{
			"progress": progress,
			"message":  message,
		},
	}

	_, err := q.collection.UpdateByID(ctx, jobID, update)
	if err != nil {
		return fmt.Errorf("failed to update progress: %w", err)
	}

	return nil
}

func (q *Queue) GetByID(ctx context.Context, jobID primitive.ObjectID) (*Job, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var job Job
	err := q.collection.FindOne(ctx, bson.M{"_id": jobID}).Decode(&job)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	return &job, nil
}

func (q *Queue) GetByUserID(ctx context.Context, userID primitive.ObjectID, limit int) ([]*Job, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if limit <= 0 {
		limit = 20
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := q.collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find jobs: %w", err)
	}
	defer cursor.Close(ctx)

	var jobs []*Job
	if err := cursor.All(ctx, &jobs); err != nil {
		return nil, fmt.Errorf("failed to decode jobs: %w", err)
	}

	return jobs, nil
}

func (q *Queue) GetByOrganizationID(ctx context.Context, orgID primitive.ObjectID, limit, skip int) ([]*Job, int, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if limit <= 0 {
		limit = 20
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(skip))

	cursor, err := q.collection.Find(ctx, bson.M{"organization_id": orgID}, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find jobs: %w", err)
	}
	defer cursor.Close(ctx)

	var jobs []*Job
	if err := cursor.All(ctx, &jobs); err != nil {
		return nil, 0, fmt.Errorf("failed to decode jobs: %w", err)
	}

	total, err := q.collection.CountDocuments(ctx, bson.M{"organization_id": orgID})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count jobs: %w", err)
	}

	return jobs, int(total), nil
}

func (q *Queue) Cancel(ctx context.Context, jobID primitive.ObjectID) error {
	q.runningMu.RLock()
	cancel, exists := q.running[jobID]
	q.runningMu.RUnlock()

	if exists && cancel != nil {
		cancel()
		slog.Debug("Sent cancel signal to running job", "id", jobID)
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":       StatusCancelled,
			"completed_at": now,
			"locked_at":    nil,
			"locked_by":    "",
		},
	}

	_, err := q.collection.UpdateByID(ctx, jobID, update)
	if err != nil {
		return fmt.Errorf("failed to cancel job: %w", err)
	}

	slog.Debug("Job cancelled", "id", jobID)
	return nil
}

func (q *Queue) KeepAlive(ctx context.Context, jobID primitive.ObjectID) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"locked_at": now,
		},
	}

	_, err := q.collection.UpdateByID(ctx, jobID, update)
	if err != nil {
		return fmt.Errorf("failed to keep job alive: %w", err)
	}

	return nil
}

func (q *Queue) ReleaseStaleLocks(ctx context.Context, timeout time.Duration) (int, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	cutoff := time.Now().Add(-timeout)

	filter := bson.M{
		"status":       StatusRunning,
		"locked_at":    bson.M{"$lt": cutoff},
		"long_running": bson.M{"$ne": true},
	}

	update := bson.M{
		"$set": bson.M{
			"status":    StatusPending,
			"locked_at": nil,
			"locked_by": "",
		},
	}

	result, err := q.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, fmt.Errorf("failed to release stale locks: %w", err)
	}

	if result.ModifiedCount > 0 {
		slog.Debug("Released stale job locks", "count", result.ModifiedCount)
	}

	return int(result.ModifiedCount), nil
}
