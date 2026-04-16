package sms

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidTo          = errors.New("invalid to phone number")
	ErrInvalidFrom        = errors.New("invalid from phone number")
	ErrInvalidMessage     = errors.New("invalid message")
	ErrSendFailed         = errors.New("failed to send SMS")
	ErrProviderNotConfig  = errors.New("SMS provider not configured")
	ErrTemplateNotFound   = errors.New("template not found")
	ErrTemplateRender     = errors.New("template rendering failed")
	ErrInvalidPhoneNumber = errors.New("invalid phone number format")
)

type Provider string

const (
	ProviderTwilio Provider = "twilio"
	ProviderMock   Provider = "mock"
)

type SendInput struct {
	From         string
	To           string
	Message      string
	TemplateID   string
	TemplateData map[string]any
	MediaURLs    []string
}

type SendOutput struct {
	MessageID string
	From      string
	To        string
	Body      string
	SentAt    time.Time
	Provider  Provider
	Status    string
}

type Template struct {
	ID        string
	Name      string
	Body      string
	Variables []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Config struct {
	Provider      Provider
	FromNumber    string
	TwilioSID     string
	TwilioToken   string
	TwilioFrom    string
	MaxMessageLen int
}

type Sender interface {
	Send(ctx context.Context, input *SendInput) (*SendOutput, error)
	HealthCheck(ctx context.Context) error
}

type Service interface {
	Send(ctx context.Context, input *SendInput) (*SendOutput, error)
	SendTemplate(ctx context.Context, templateID string, to string, data map[string]any) (*SendOutput, error)
	GetTemplate(ctx context.Context, templateID string) (*Template, error)
	HealthCheck(ctx context.Context) error
}

func (i *SendInput) Validate() error {
	if i.To == "" {
		return ErrInvalidTo
	}
	if i.Message == "" && i.TemplateID == "" {
		return ErrInvalidMessage
	}
	return nil
}
