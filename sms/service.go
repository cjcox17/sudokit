package sms

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type SMSService struct {
	sender    Sender
	templates map[string]*Template
	mu        sync.RWMutex
}

func NewService(cfg *Config) (*SMSService, error) {
	var sender Sender

	switch cfg.Provider {
	case ProviderTwilio:
		if cfg.TwilioSID == "" || cfg.TwilioToken == "" {
			return nil, ErrProviderNotConfig
		}
		from := cfg.TwilioFrom
		if from == "" {
			from = cfg.FromNumber
		}
		sender = NewTwilioSender(cfg.TwilioSID, cfg.TwilioToken, from)
	case ProviderMock:
		sender = NewMockSender()
	default:
		return nil, ErrProviderNotConfig
	}

	return &SMSService{
		sender:    sender,
		templates: make(map[string]*Template),
	}, nil
}

func (s *SMSService) Send(ctx context.Context, input *SendInput) (*SendOutput, error) {
	if input.From == "" {
		input.From = s.sender.(*TwilioSender).from
	}
	return s.sender.Send(ctx, input)
}

func (s *SMSService) SendTemplate(ctx context.Context, templateID string, to string, data map[string]any) (*SendOutput, error) {
	tmpl, err := s.GetTemplate(ctx, templateID)
	if err != nil {
		return nil, err
	}

	return s.Send(ctx, &SendInput{
		To:           to,
		TemplateID:   templateID,
		TemplateData: data,
		Message:      tmpl.Body,
	})
}

func (s *SMSService) GetTemplate(ctx context.Context, templateID string) (*Template, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tmpl, ok := s.templates[templateID]
	if !ok {
		return nil, ErrTemplateNotFound
	}

	return tmpl, nil
}

func (s *SMSService) RegisterTemplate(tmpl *Template) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.templates[tmpl.ID] = tmpl
}

func (s *SMSService) HealthCheck(ctx context.Context) error {
	return s.sender.HealthCheck(ctx)
}

type MockSender struct {
	messages []*SendOutput
	mu       sync.Mutex
}

func NewMockSender() *MockSender {
	return &MockSender{
		messages: make([]*SendOutput, 0),
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
		Body:      input.Message,
		SentAt:    time.Now(),
		Provider:  ProviderMock,
		Status:    "sent",
	}

	s.mu.Lock()
	s.messages = append(s.messages, output)
	s.mu.Unlock()

	return output, nil
}

func (s *MockSender) HealthCheck(ctx context.Context) error {
	return nil
}

func (s *MockSender) GetMessages() []*SendOutput {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.messages
}

func (s *MockSender) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = make([]*SendOutput, 0)
}
