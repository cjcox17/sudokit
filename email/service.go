package email

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type EmailService struct {
	sender    Sender
	templates map[string]*Template
	mu        sync.RWMutex
}

func NewService(cfg *Config) (*EmailService, error) {
	var sender Sender

	switch cfg.Provider {
	case ProviderMailgun:
		if cfg.MailgunDomain == "" || cfg.MailgunAPIKey == "" {
			return nil, ErrProviderNotConfig
		}
		sender = NewMailgunSender(cfg.MailgunDomain, cfg.MailgunAPIKey, cfg.FromAddress, cfg.FromName)
	case ProviderSMTP:
		if cfg.SMTPHost == "" {
			return nil, ErrProviderNotConfig
		}
		sender = NewSMTPSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, cfg.FromAddress, cfg.FromName, cfg.SMTPUseTLS)
	case ProviderMock:
		sender = NewMockSender()
	default:
		return nil, ErrProviderNotConfig
	}

	return &EmailService{
		sender:    sender,
		templates: make(map[string]*Template),
	}, nil
}

func (s *EmailService) Send(ctx context.Context, input *SendInput) (*SendOutput, error) {
	return s.sender.Send(ctx, input)
}

func (s *EmailService) SendTemplate(ctx context.Context, templateID string, to []string, data map[string]any) (*SendOutput, error) {
	rendered, err := s.Render(ctx, &RenderInput{
		TemplateID: templateID,
		Data:       data,
	})
	if err != nil {
		return nil, err
	}

	tmpl, err := s.GetTemplate(ctx, templateID)
	if err != nil {
		return nil, err
	}

	return s.Send(ctx, &SendInput{
		From:     tmpl.From,
		FromName: tmpl.FromName,
		To:       to,
		Subject:  rendered.Subject,
		HTMLBody: rendered.HTMLBody,
		TextBody: rendered.TextBody,
	})
}

func (s *EmailService) Render(ctx context.Context, input *RenderInput) (*RenderOutput, error) {
	tmpl, err := s.GetTemplate(ctx, input.TemplateID)
	if err != nil {
		return nil, err
	}

	return &RenderOutput{
		Subject:  tmpl.Subject,
		HTMLBody: tmpl.HTMLBody,
		TextBody: tmpl.TextBody,
	}, nil
}

func (s *EmailService) GetTemplate(ctx context.Context, templateID string) (*Template, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tmpl, ok := s.templates[templateID]
	if !ok {
		return nil, ErrTemplateNotFound
	}

	return tmpl, nil
}

func (s *EmailService) RegisterTemplate(tmpl *Template) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.templates[tmpl.ID] = tmpl
}

func (s *EmailService) HealthCheck(ctx context.Context) error {
	return s.sender.HealthCheck(ctx)
}

type MockSender struct {
	emails []*SendOutput
	mu     sync.Mutex
}

func NewMockSender() *MockSender {
	return &MockSender{
		emails: make([]*SendOutput, 0),
	}
}

func (s *MockSender) Send(ctx context.Context, input *SendInput) (*SendOutput, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	output := &SendOutput{
		MessageID: fmt.Sprintf("mock-%d", time.Now().UnixNano()),
		From:      input.From,
		To:        input.To,
		SentAt:    time.Now(),
		Provider:  ProviderMock,
	}

	s.mu.Lock()
	s.emails = append(s.emails, output)
	s.mu.Unlock()

	return output, nil
}

func (s *MockSender) HealthCheck(ctx context.Context) error {
	return nil
}

func (s *MockSender) GetEmails() []*SendOutput {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.emails
}

func (s *MockSender) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.emails = make([]*SendOutput, 0)
}
