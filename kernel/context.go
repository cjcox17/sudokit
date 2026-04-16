package kernel

import (
	"context"

	"github.com/cjcox17/sudokit/cache"
	"github.com/cjcox17/sudokit/email"
	"github.com/cjcox17/sudokit/events"
	"github.com/cjcox17/sudokit/jobs"
	"github.com/cjcox17/sudokit/notifications"
	"github.com/cjcox17/sudokit/sms"
	"github.com/cjcox17/sudokit/storage"
	"github.com/cjcox17/sudokit/websocket"
)

type ctxKey struct{}

func WithServices(ctx context.Context, k *Kernel) context.Context {
	return context.WithValue(ctx, ctxKey{}, k)
}

func FromContext(ctx context.Context) *Kernel {
	if k, ok := ctx.Value(ctxKey{}).(*Kernel); ok {
		return k
	}
	return App()
}

func Jobs(ctx context.Context) jobs.ServiceInterface {
	return FromContext(ctx).Jobs()
}

func Events(ctx context.Context) events.ServiceInterface {
	return FromContext(ctx).Events()
}

func Notify(ctx context.Context) notifications.ServiceInterface {
	return FromContext(ctx).Notifications()
}

func Broadcast(ctx context.Context) websocket.Hub {
	return FromContext(ctx).WebSocket()
}

func Storage(ctx context.Context) storage.Service {
	return FromContext(ctx).Storage()
}

func Email(ctx context.Context) email.Service {
	return FromContext(ctx).Email()
}

func SMS(ctx context.Context) sms.Service {
	return FromContext(ctx).SMS()
}

func Cache(ctx context.Context) cache.Service {
	return FromContext(ctx).Cache()
}

func GetJobs(ctx context.Context) jobs.ServiceInterface {
	return Jobs(ctx)
}

func GetEvents(ctx context.Context) events.ServiceInterface {
	return Events(ctx)
}

func GetNotifications(ctx context.Context) notifications.ServiceInterface {
	return Notify(ctx)
}

func GetWebSocket(ctx context.Context) websocket.Hub {
	return Broadcast(ctx)
}

func GetStorage(ctx context.Context) storage.Service {
	return Storage(ctx)
}

func GetEmail(ctx context.Context) email.Service {
	return Email(ctx)
}

func GetSMS(ctx context.Context) sms.Service {
	return SMS(ctx)
}

func GetCache(ctx context.Context) cache.Service {
	return Cache(ctx)
}
