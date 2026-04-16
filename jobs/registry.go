package jobs

import (
	"fmt"
	"sync"
)

type Registry struct {
	handlers map[string]Handler
	mu       sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{
		handlers: make(map[string]Handler),
	}
}

func (r *Registry) Register(jobType string, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[jobType] = handler
}

func (r *Registry) RegisterFunc(jobType string, handler HandlerFunc) {
	r.Register(jobType, handler)
}

func (r *Registry) GetHandler(jobType string) (Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	handler, ok := r.handlers[jobType]
	return handler, ok
}

func (r *Registry) HasHandler(jobType string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.handlers[jobType]
	return ok
}

func (r *Registry) RegisteredTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.handlers))
	for t := range r.handlers {
		types = append(types, t)
	}
	return types
}

func (r *Registry) Validate() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.handlers) == 0 {
		return fmt.Errorf("no handlers registered")
	}

	return nil
}
