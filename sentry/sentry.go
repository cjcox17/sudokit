package sentry

import (
	"github.com/getsentry/sentry-go"
	sentryfiber "github.com/getsentry/sentry-go/fiber"
	"github.com/gofiber/fiber/v2"
)

var enabled bool

func Init(dsn string) {
	if dsn == "" {
		enabled = false
		return
	}
	enabled = true
}

func CaptureError(c *fiber.Ctx, err error) {
	if !enabled || err == nil {
		return
	}

	if hub := sentryfiber.GetHubFromContext(c); hub != nil {
		hub.CaptureException(err)
	}
}

func CaptureMessage(c *fiber.Ctx, message string) {
	if !enabled {
		return
	}

	if hub := sentryfiber.GetHubFromContext(c); hub != nil {
		hub.CaptureMessage(message)
	}
}

func CaptureErrorWithContext(c *fiber.Ctx, err error, extras map[string]any) {
	if !enabled || err == nil {
		return
	}

	if hub := sentryfiber.GetHubFromContext(c); hub != nil {
		hub.WithScope(func(scope *sentry.Scope) {
			for key, value := range extras {
				scope.SetExtra(key, value)
			}
			hub.CaptureException(err)
		})
	}
}

func CaptureMessageWithContext(c *fiber.Ctx, message string, extras map[string]any) {
	if !enabled {
		return
	}

	if hub := sentryfiber.GetHubFromContext(c); hub != nil {
		hub.WithScope(func(scope *sentry.Scope) {
			for key, value := range extras {
				scope.SetExtra(key, value)
			}
			hub.CaptureMessage(message)
		})
	}
}

func SetSentryUser(c *fiber.Ctx, userID, email, organizationID string) {
	if !enabled {
		return
	}

	if hub := sentryfiber.GetHubFromContext(c); hub != nil {
		hub.Scope().SetUser(sentry.User{
			ID:    userID,
			Email: email,
			Data: map[string]string{
				"organization_id": organizationID,
			},
		})
	}
}

func SetSentryTag(c *fiber.Ctx, key, value string) {
	if !enabled {
		return
	}

	if hub := sentryfiber.GetHubFromContext(c); hub != nil {
		hub.Scope().SetTag(key, value)
	}
}
