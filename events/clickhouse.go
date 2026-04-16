package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ClickHouseConsumer struct {
	mu         sync.RWMutex
	conn       driver.Conn
	service    *Service
	batchSize  int
	batchQueue chan *BaseEvent
	stopCh     chan struct{}
	started    bool
}

func NewClickHouseConsumer(conn driver.Conn, service *Service, batchSize int) *ClickHouseConsumer {
	if batchSize <= 0 {
		batchSize = 1000
	}

	return &ClickHouseConsumer{
		conn:       conn,
		service:    service,
		batchSize:  batchSize,
		batchQueue: make(chan *BaseEvent, batchSize*2),
		stopCh:     make(chan struct{}),
		started:    false,
	}
}

func (ec *ClickHouseConsumer) Start() error {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	if ec.started {
		return nil
	}

	if ec.service == nil {
		slog.Warn("Events service not set, ClickHouse consumer not started")
		return nil
	}

	err := ec.service.Subscribe("domain.event", ec.handleEvent)
	if err != nil {
		slog.Error("Failed to subscribe to domain events", "error", err)
		return err
	}

	go ec.processBatches()

	ec.started = true
	slog.Info("ClickHouse event consumer started")
	return nil
}

func (ec *ClickHouseConsumer) handleEvent(ctx context.Context, event *BaseEvent) error {
	start := time.Now()
	select {
	case ec.batchQueue <- event:
		duration := time.Since(start)
		if duration > 100*time.Millisecond {
			slog.Warn("SLOW EVENT QUEUE", "event_type", event.EventType, "duration", duration)
		}
		return nil
	case <-time.After(5 * time.Second):
		duration := time.Since(start)
		slog.Error("TIMEOUT queueing event", "event_type", event.EventType, "event_id", event.ID.Hex(), "duration", duration)
		return fmt.Errorf("timeout queueing event")
	}
}

func (ec *ClickHouseConsumer) processBatches() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	batch := make([]*BaseEvent, 0, ec.batchSize)

	flushBatch := func() {
		if len(batch) == 0 {
			return
		}

		if err := ec.insertBatch(batch); err != nil {
			slog.Error("Failed to insert batch of events", "batch_size", len(batch), "error", err)
		} else {
			slog.Debug("Successfully inserted batch of events to ClickHouse", "batch_size", len(batch))
		}

		batch = batch[:0]
	}

	for {
		select {
		case event := <-ec.batchQueue:
			batch = append(batch, event)

			if len(batch) >= ec.batchSize {
				flushBatch()
			}

		case <-ticker.C:
			flushBatch()

		case <-ec.stopCh:
			flushBatch()
			return
		}
	}
}

func (ec *ClickHouseConsumer) insertBatch(events []*BaseEvent) error {
	if len(events) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	batch, err := ec.conn.PrepareBatch(ctx, "INSERT INTO events (id, event_type, aggregate_type, aggregate_id, organization_id, user_id, user_name, request_id, payload, metadata, timestamp, processed_by)")
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, event := range events {
		timestamp := event.Timestamp.UTC()

		payloadJSON, err := json.Marshal(event.Payload)
		if err != nil {
			slog.Warn("Failed to marshal payload for event", "event_id", event.ID.Hex(), "error", err)
			continue
		}

		metadataJSON, err := json.Marshal(event.Metadata)
		if err != nil {
			slog.Warn("Failed to marshal metadata for event", "event_id", event.ID.Hex(), "error", err)
			continue
		}

		err = batch.Append(
			event.ID.Hex(),
			event.EventType,
			event.AggregateType,
			event.AggregateID,
			event.OrganizationID,
			event.UserID,
			event.UserName,
			event.RequestID,
			string(payloadJSON),
			string(metadataJSON),
			timestamp,
			event.ProcessedBy,
		)

		if err != nil {
			slog.Warn("Failed to append event to batch", "event_id", event.ID.Hex(), "error", err)
			continue
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	return nil
}

func (ec *ClickHouseConsumer) Stop() error {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	if !ec.started {
		return nil
	}

	close(ec.stopCh)
	ec.started = false

	slog.Info("ClickHouse event consumer stopped")
	return nil
}

func (ec *ClickHouseConsumer) IsStarted() bool {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	return ec.started
}

func (ec *ClickHouseConsumer) GetEvents(ctx context.Context, filter map[string]any, limit int64, offset int64) ([]*BaseEvent, error) {
	query := "SELECT id, event_type, aggregate_type, aggregate_id, organization_id, user_id, user_name, request_id, payload, metadata, timestamp, processed_by FROM events"

	whereClauses := make([]string, 0)
	args := make([]interface{}, 0)

	if eventType, ok := filter["event_type"].(string); ok {
		whereClauses = append(whereClauses, "event_type = ?")
		args = append(args, eventType)
	}

	if aggregateType, ok := filter["aggregate_type"].(string); ok {
		whereClauses = append(whereClauses, "aggregate_type = ?")
		args = append(args, aggregateType)
	}

	if aggregateID, ok := filter["aggregate_id"].(string); ok {
		whereClauses = append(whereClauses, "aggregate_id = ?")
		args = append(args, aggregateID)
	}

	if orgID, ok := filter["organization_id"].(string); ok {
		whereClauses = append(whereClauses, "organization_id = ?")
		args = append(args, orgID)
	}

	if len(whereClauses) > 0 {
		query += " WHERE "
		for i, clause := range whereClauses {
			if i > 0 {
				query += " AND "
			}
			query += clause
		}
	}

	query += " ORDER BY timestamp DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	rows, err := ec.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	events := make([]*BaseEvent, 0)

	for rows.Next() {
		var (
			id             string
			eventType      string
			aggregateType  string
			aggregateID    string
			organizationID string
			userID         string
			userName       string
			requestID      string
			payloadJSON    []byte
			metadataJSON   []byte
			timestamp      time.Time
			processedBy    []string
		)

		err := rows.Scan(
			&id,
			&eventType,
			&aggregateType,
			&aggregateID,
			&organizationID,
			&userID,
			&userName,
			&requestID,
			&payloadJSON,
			&metadataJSON,
			&timestamp,
			&processedBy,
		)

		if err != nil {
			slog.Warn("Failed to scan event row", "error", err)
			continue
		}

		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			slog.Warn("Failed to parse ObjectID", "id", id, "error", err)
			continue
		}

		var payload map[string]any
		if err := json.Unmarshal(payloadJSON, &payload); err != nil {
			payload = make(map[string]any)
		}

		var metadata map[string]any
		if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
			metadata = make(map[string]any)
		}

		if userName != "" {
			if metadata == nil {
				metadata = make(map[string]any)
			}
			metadata["user_name"] = userName
		}

		if requestID != "" {
			if metadata == nil {
				metadata = make(map[string]any)
			}
			metadata["request_id"] = requestID
		}

		event := &BaseEvent{
			ID:             objectID,
			EventType:      eventType,
			AggregateType:  aggregateType,
			AggregateID:    aggregateID,
			UserID:         userID,
			UserName:       userName,
			OrganizationID: organizationID,
			RequestID:      requestID,
			Payload:        payload,
			Metadata:       metadata,
			Timestamp:      timestamp,
			ProcessedBy:    processedBy,
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return events, nil
}

func (ec *ClickHouseConsumer) GetEventsByType(ctx context.Context, eventType string, limit int64) ([]*BaseEvent, error) {
	filter := map[string]any{"event_type": eventType}
	return ec.GetEvents(ctx, filter, limit, 0)
}

func (ec *ClickHouseConsumer) GetEventsByAggregate(ctx context.Context, aggregateType, aggregateID string) ([]*BaseEvent, error) {
	filter := map[string]any{
		"aggregate_type": aggregateType,
		"aggregate_id":   aggregateID,
	}
	return ec.GetEvents(ctx, filter, 0, 0)
}

func (ec *ClickHouseConsumer) GetEventsByAggregateWithPagination(ctx context.Context, aggregateType, aggregateID string, limit int64, skip int64) ([]*BaseEvent, int64, error) {
	filter := map[string]any{
		"aggregate_type": aggregateType,
		"aggregate_id":   aggregateID,
	}

	countQuery := "SELECT count() FROM events WHERE aggregate_type = ? AND aggregate_id = ?"
	var total uint64
	err := ec.conn.QueryRow(ctx, countQuery, aggregateType, aggregateID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count events: %w", err)
	}

	events, err := ec.GetEvents(ctx, filter, limit, skip)
	if err != nil {
		return nil, 0, err
	}

	return events, int64(total), nil
}

func (ec *ClickHouseConsumer) GetEventsByOrganization(ctx context.Context, orgID string, limit int64) ([]*BaseEvent, error) {
	filter := map[string]any{"organization_id": orgID}
	return ec.GetEvents(ctx, filter, limit, 0)
}

func (ec *ClickHouseConsumer) GetRecentEvents(ctx context.Context, limit int64) ([]*BaseEvent, error) {
	return ec.GetEvents(ctx, map[string]any{}, limit, 0)
}

func (ec *ClickHouseConsumer) CountEvents(ctx context.Context, filter map[string]any) (int64, error) {
	query := "SELECT count() FROM events"

	whereClauses := make([]string, 0)
	args := make([]interface{}, 0)

	if eventType, ok := filter["event_type"].(string); ok {
		whereClauses = append(whereClauses, "event_type = ?")
		args = append(args, eventType)
	}

	if aggregateType, ok := filter["aggregate_type"].(string); ok {
		whereClauses = append(whereClauses, "aggregate_type = ?")
		args = append(args, aggregateType)
	}

	if aggregateID, ok := filter["aggregate_id"].(string); ok {
		whereClauses = append(whereClauses, "aggregate_id = ?")
		args = append(args, aggregateID)
	}

	if orgID, ok := filter["organization_id"].(string); ok {
		whereClauses = append(whereClauses, "organization_id = ?")
		args = append(args, orgID)
	}

	if len(whereClauses) > 0 {
		query += " WHERE "
		for i, clause := range whereClauses {
			if i > 0 {
				query += " AND "
			}
			query += clause
		}
	}

	var count uint64
	err := ec.conn.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count events: %w", err)
	}

	return int64(count), nil
}

func (ec *ClickHouseConsumer) GetEventsByAggregateWithPaginationAndType(ctx context.Context, aggregateType, aggregateID string, eventType string, limit, offset int64) ([]*BaseEvent, int64, error) {
	filter := map[string]any{
		"aggregate_type": aggregateType,
		"aggregate_id":   aggregateID,
	}
	if eventType != "" {
		filter["event_type"] = eventType
	}

	events, err := ec.GetEvents(ctx, filter, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := ec.CountEvents(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

var GlobalClickHouseConsumer *ClickHouseConsumer

func InitGlobalClickHouseConsumer(conn driver.Conn, service *Service, batchSize int) error {
	GlobalClickHouseConsumer = NewClickHouseConsumer(conn, service, batchSize)
	return GlobalClickHouseConsumer.Start()
}
