package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/asaskevich/EventBus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service struct {
	bus     EventBus.Bus
	hub     Broadcaster
	mu      sync.RWMutex
	started bool
}

type Broadcaster interface {
	BroadcastToUser(userID primitive.ObjectID, event string, data any)
	BroadcastToOrg(orgID primitive.ObjectID, event string, data any)
}

func NewService(hub Broadcaster) *Service {
	return &Service{
		bus:     EventBus.New(),
		hub:     hub,
		started: false,
	}
}

func (s *Service) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.started = true
	slog.Info("Events service started")
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.started = false
	slog.Info("Events service stopped")
	return nil
}

func (s *Service) Publish(ctx context.Context, eventType string, payload any, opts ...EventOption) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.started {
		return fmt.Errorf("events service not started")
	}

	event := &BaseEvent{
		ID:        primitive.NewObjectID(),
		EventType: eventType,
		Payload:   make(map[string]any),
		Metadata:  make(map[string]any),
	}

	for _, opt := range opts {
		opt(event)
	}

	if p, ok := payload.(map[string]any); ok {
		event.Payload = p
	} else {
		data, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		if err := json.Unmarshal(data, &event.Payload); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}
	}

	s.bus.Publish("domain.event", event)
	s.bus.Publish(eventType, event)

	if s.hub != nil && event.OrganizationID != "" {
		go func() {
			orgID, _ := primitive.ObjectIDFromHex(event.OrganizationID)
			s.hub.BroadcastToOrg(orgID, "event", map[string]any{
				"type":      eventType,
				"payload":   event.Payload,
				"user_id":   event.UserID,
				"timestamp": event.Timestamp,
			})
		}()
	}

	slog.Debug("Event published", "type", eventType, "id", event.ID)

	return nil
}

func (s *Service) Subscribe(eventType string, handler Handler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.bus.SubscribeAsync(eventType, func(event *BaseEvent) {
		if err := handler(context.Background(), event); err != nil {
			slog.Error("Event handler failed", "type", eventType, "error", err)
		}
	}, false)
}

func (s *Service) SubscribeSync(eventType string, handler Handler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.bus.Subscribe(eventType, func(event *BaseEvent) {
		if err := handler(context.Background(), event); err != nil {
			slog.Error("Event handler failed", "type", eventType, "error", err)
		}
	})
}

func SubscribeTyped[T any](s *Service, eventType string, handler func(ctx context.Context, e *Event[T]) error) error {
	return s.Subscribe(eventType, func(ctx context.Context, event *BaseEvent) error {
		var typedPayload T
		if event.Payload != nil {
			data, err := json.Marshal(event.Payload)
			if err != nil {
				return fmt.Errorf("failed to marshal payload: %w", err)
			}
			if err := json.Unmarshal(data, &typedPayload); err != nil {
				return fmt.Errorf("failed to unmarshal payload: %w", err)
			}
		}

		typedEvent := &Event[T]{
			ID:             event.ID,
			EventType:      event.EventType,
			AggregateType:  event.AggregateType,
			AggregateID:    event.AggregateID,
			UserID:         event.UserID,
			UserName:       event.UserName,
			OrganizationID: event.OrganizationID,
			RequestID:      event.RequestID,
			Payload:        typedPayload,
			Metadata:       event.Metadata,
			Timestamp:      event.Timestamp,
			ProcessedAt:    event.ProcessedAt,
			ProcessedBy:    event.ProcessedBy,
		}

		return handler(ctx, typedEvent)
	})
}

type Handler func(ctx context.Context, event *BaseEvent) error

var (
	globalService *Service
	serviceOnce   sync.Once
)

func InitService(hub Broadcaster) *Service {
	serviceOnce.Do(func() {
		globalService = NewService(hub)
	})
	return globalService
}

func GetService() *Service {
	return globalService
}
