package kernel

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/cjcox17/sudokit/cache"
	"github.com/cjcox17/sudokit/email"
	"github.com/cjcox17/sudokit/events"
	"github.com/cjcox17/sudokit/jobs"
	"github.com/cjcox17/sudokit/notifications"
	"github.com/cjcox17/sudokit/sms"
	"github.com/cjcox17/sudokit/storage"
	"github.com/cjcox17/sudokit/websocket"
)

var (
	app  *Kernel
	once sync.Once
	mu   sync.RWMutex
)

type Kernel struct {
	jobs          jobs.ServiceInterface
	events        events.ServiceInterface
	notifications notifications.ServiceInterface
	webSocket     websocket.Hub
	storage       storage.Service
	email         email.Service
	sms           sms.Service
	cache         cache.Service
	config        *Config
	started       bool
	mu            sync.RWMutex
}

type Services struct {
	Jobs          jobs.ServiceInterface
	Events        events.ServiceInterface
	Notifications notifications.ServiceInterface
	WebSocket     websocket.Hub
	Storage       storage.Service
	Email         email.Service
	SMS           sms.Service
	Cache         cache.Service
}

func Boot(cfg *Config, services *Services) error {
	var initErr error
	once.Do(func() {
		app = &Kernel{
			config:        cfg,
			jobs:          services.Jobs,
			events:        services.Events,
			notifications: services.Notifications,
			webSocket:     services.WebSocket,
			storage:       services.Storage,
			email:         services.Email,
			sms:           services.SMS,
			cache:         services.Cache,
			started:       false,
		}

		if err := app.startServices(); err != nil {
			initErr = err
			return
		}

		app.started = true
		slog.Info("SudoKit kernel booted successfully")
	})

	return initErr
}

func (k *Kernel) startServices() error {
	// Only start jobs service - events service should be started before kernel.Boot()
	// to ensure handlers are attached before job events are published
	if k.jobs != nil {
		slog.Info("Kernel: starting jobs service...")
		if err := k.jobs.Start(); err != nil {
			return fmt.Errorf("failed to start jobs service: %w", err)
		}
		slog.Info("Jobs service started")
	}

	return nil
}

func Shutdown(ctx context.Context) error {
	mu.Lock()
	defer mu.Unlock()

	if app == nil {
		return nil
	}

	slog.Info("Shutting down SudoKit kernel...")

	var errs []error

	if app.jobs != nil {
		if err := app.jobs.Stop(ctx); err != nil {
			errs = append(errs, fmt.Errorf("jobs: %w", err))
		}
	}

	if app.events != nil {
		if err := app.events.Stop(ctx); err != nil {
			errs = append(errs, fmt.Errorf("events: %w", err))
		}
	}

	app = nil
	once = sync.Once{}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	slog.Info("SudoKit kernel shutdown complete")
	return nil
}

func App() *Kernel {
	mu.RLock()
	defer mu.RUnlock()
	return app
}

func (k *Kernel) Jobs() jobs.ServiceInterface {
	return k.jobs
}

func (k *Kernel) Events() events.ServiceInterface {
	return k.events
}

func (k *Kernel) Notifications() notifications.ServiceInterface {
	return k.notifications
}

func (k *Kernel) WebSocket() websocket.Hub {
	return k.webSocket
}

func (k *Kernel) Storage() storage.Service {
	return k.storage
}

func (k *Kernel) Email() email.Service {
	return k.email
}

func (k *Kernel) SMS() sms.Service {
	return k.sms
}

func (k *Kernel) Cache() cache.Service {
	return k.cache
}

func (k *Kernel) IsStarted() bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.started
}

func NewWithServices(services *Services) *Kernel {
	return &Kernel{
		jobs:          services.Jobs,
		events:        services.Events,
		notifications: services.Notifications,
		webSocket:     services.WebSocket,
		storage:       services.Storage,
		email:         services.Email,
		sms:           services.SMS,
		cache:         services.Cache,
		config:        &Config{},
		started:       true,
	}
}
