package events

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type HandlerRegistry struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
	service  *Service
}

func NewHandlerRegistry(service *Service) *HandlerRegistry {
	return &HandlerRegistry{
		handlers: make(map[string][]Handler),
		service:  service,
	}
}

func (hr *HandlerRegistry) Register(eventType string, handler Handler) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	if hr.handlers[eventType] == nil {
		hr.handlers[eventType] = []Handler{}
	}
	hr.handlers[eventType] = append(hr.handlers[eventType], handler)
	slog.Debug("Registered handler for event type", "event_type", eventType)
}

func (hr *HandlerRegistry) RegisterFunc(eventType string, handler func(ctx context.Context, event *BaseEvent) error) {
	hr.Register(eventType, handler)
}

func (hr *HandlerRegistry) Attach() error {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	if hr.service == nil {
		slog.Warn("Service not set, handlers not attached")
		return nil
	}

	for eventType, handlers := range hr.handlers {
		for _, handler := range handlers {
			err := hr.service.Subscribe(eventType, handler)
			if err != nil {
				slog.Error("Failed to attach handler", "event_type", eventType, "error", err)
				return err
			}
		}
		slog.Debug("Attached handlers for event type", "count", len(handlers), "event_type", eventType)
	}

	return nil
}

func (hr *HandlerRegistry) RegisteredTypes() []string {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	types := make([]string, 0, len(hr.handlers))
	for t := range hr.handlers {
		types = append(types, t)
	}
	return types
}

func RetryableHandler(handler Handler, maxRetries int, backoff time.Duration) Handler {
	return func(ctx context.Context, event *BaseEvent) error {
		var err error
		for i := 0; i < maxRetries; i++ {
			err = handler(ctx, event)
			if err == nil {
				return nil
			}

			slog.Warn("Handler failed for event", "event_type", event.EventType, "attempt", i+1, "max_retries", maxRetries, "error", err)

			if i < maxRetries-1 {
				time.Sleep(backoff * time.Duration(i+1))
			}
		}

		slog.Error("Handler failed permanently for event", "event_type", event.EventType, "max_retries", maxRetries, "error", err)
		return err
	}
}

func FilteredHandler(handler Handler, filter func(*BaseEvent) bool) Handler {
	return func(ctx context.Context, event *BaseEvent) error {
		if filter(event) {
			return handler(ctx, event)
		}
		return nil
	}
}

type BatchHandler struct {
	mu           sync.Mutex
	batchSize    int
	flushTimeout time.Duration
	handler      func(ctx context.Context, events []*BaseEvent) error
	events       []*BaseEvent
	ticker       *time.Ticker
	done         chan bool
	service      *Service
}

func NewBatchHandler(service *Service, batchSize int, flushTimeout time.Duration, handler func(ctx context.Context, events []*BaseEvent) error) *BatchHandler {
	bh := &BatchHandler{
		batchSize:    batchSize,
		flushTimeout: flushTimeout,
		handler:      handler,
		events:       make([]*BaseEvent, 0, batchSize),
		ticker:       time.NewTicker(flushTimeout),
		done:         make(chan bool),
		service:      service,
	}

	go bh.run()
	return bh
}

func (bh *BatchHandler) run() {
	for {
		select {
		case <-bh.ticker.C:
			bh.flush()
		case <-bh.done:
			bh.flush()
			return
		}
	}
}

func (bh *BatchHandler) Handle(ctx context.Context, event *BaseEvent) error {
	bh.mu.Lock()
	defer bh.mu.Unlock()

	bh.events = append(bh.events, event)

	if len(bh.events) >= bh.batchSize {
		return bh.flushUnlocked()
	}

	return nil
}

func (bh *BatchHandler) flush() error {
	bh.mu.Lock()
	defer bh.mu.Unlock()
	return bh.flushUnlocked()
}

func (bh *BatchHandler) flushUnlocked() error {
	if len(bh.events) == 0 {
		return nil
	}

	events := bh.events
	bh.events = make([]*BaseEvent, 0, bh.batchSize)

	return bh.handler(context.Background(), events)
}

func (bh *BatchHandler) Stop() {
	bh.ticker.Stop()
	bh.done <- true
}

func (bh *BatchHandler) AsHandler() Handler {
	return bh.Handle
}
