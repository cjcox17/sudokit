package email

import (
	"context"
	"errors"
	"io"
	"time"
)

var (
	ErrInvalidFrom        = errors.New("invalid from address")
	ErrInvalidTo          = errors.New("invalid to address")
	ErrInvalidSubject     = errors.New("invalid subject")
	ErrInvalidBody        = errors.New("invalid body")
	ErrSendFailed         = errors.New("failed to send email")
	ErrTemplateNotFound   = errors.New("template not found")
	ErrTemplateRender     = errors.New("template rendering failed")
	ErrProviderNotConfig  = errors.New("email provider not configured")
	ErrAttachmentTooLarge = errors.New("attachment too large")
)

type Provider string

const (
	ProviderMailgun  Provider = "mailgun"
	ProviderSendGrid Provider = "sendgrid"
	ProviderSMTP     Provider = "smtp"
	ProviderMock     Provider = "mock"
)

type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

type AttachmentReader struct {
	Filename    string
	ContentType string
	Reader      io.Reader
	Size        int64
}

type SendInput struct {
	From         string
	FromName     string
	To           []string
	CC           []string
	BCC          []string
	Subject      string
	HTMLBody     string
	TextBody     string
	Attachments  []Attachment
	Headers      map[string]string
	TemplateID   string
	TemplateData map[string]any
}

type SendOutput struct {
	MessageID string
	From      string
	To        []string
	SentAt    time.Time
	Provider  Provider
}

type Template struct {
	ID        string
	Name      string
	Subject   string
	HTMLBody  string
	TextBody  string
	From      string
	FromName  string
	Variables []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RenderInput struct {
	TemplateID string
	Data       map[string]any
}

type RenderOutput struct {
	Subject  string
	HTMLBody string
	TextBody string
}

type Config struct {
	Provider          Provider
	FromAddress       string
	FromName          string
	ReplyTo           string
	MailgunDomain     string
	MailgunAPIKey     string
	SendGridAPIKey    string
	SMTPHost          string
	SMTPPort          int
	SMTPUsername      string
	SMTPPassword      string
	SMTPUseTLS        bool
	MaxAttachmentSize int64
}

type Sender interface {
	Send(ctx context.Context, input *SendInput) (*SendOutput, error)
	HealthCheck(ctx context.Context) error
}

type Service interface {
	Send(ctx context.Context, input *SendInput) (*SendOutput, error)
	SendTemplate(ctx context.Context, templateID string, to []string, data map[string]any) (*SendOutput, error)
	Render(ctx context.Context, input *RenderInput) (*RenderOutput, error)
	GetTemplate(ctx context.Context, templateID string) (*Template, error)
	HealthCheck(ctx context.Context) error
}

func (i *SendInput) Validate() error {
	if i.From == "" {
		return ErrInvalidFrom
	}
	if len(i.To) == 0 {
		return ErrInvalidTo
	}
	for _, addr := range i.To {
		if addr == "" {
			return ErrInvalidTo
		}
	}
	if i.Subject == "" {
		return ErrInvalidSubject
	}
	if i.HTMLBody == "" && i.TextBody == "" {
		return ErrInvalidBody
	}
	return nil
}
