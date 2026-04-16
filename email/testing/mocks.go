package testing

import (
	"context"
	"sync"
	"time"

	"github.com/cjcox17/sudokit/email"
)

type MockEmailService struct {
	mu sync.Mutex

	SendFunc         func(ctx context.Context, input *email.SendInput) (*email.SendOutput, error)
	SendTemplateFunc func(ctx context.Context, templateID string, to []string, data map[string]any) (*email.SendOutput, error)
	RenderFunc       func(ctx context.Context, input *email.RenderInput) (*email.RenderOutput, error)
	GetTemplateFunc  func(ctx context.Context, templateID string) (*email.Template, error)
	HealthCheckFunc  func(ctx context.Context) error

	SendCalls         []SendCall
	SendTemplateCalls []SendTemplateCall
	RenderCalls       []RenderCall
	GetTemplateCalls  []GetTemplateCall

	emails    []*email.SendOutput
	templates map[string]*email.Template
}

type SendCall struct {
	Input *email.SendInput
}

type SendTemplateCall struct {
	TemplateID string
	To         []string
	Data       map[string]any
}

type RenderCall struct {
	Input *email.RenderInput
}

type GetTemplateCall struct {
	TemplateID string
}

func NewMockEmailService() *MockEmailService {
	return &MockEmailService{
		emails:    make([]*email.SendOutput, 0),
		templates: make(map[string]*email.Template),
	}
}

func (s *MockEmailService) Send(ctx context.Context, input *email.SendInput) (*email.SendOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.sendLocked(ctx, input)
}

func (s *MockEmailService) sendLocked(ctx context.Context, input *email.SendInput) (*email.SendOutput, error) {
	s.SendCalls = append(s.SendCalls, SendCall{Input: input})

	if s.SendFunc != nil {
		return s.SendFunc(ctx, input)
	}

	if err := input.Validate(); err != nil {
		return nil, err
	}

	output := &email.SendOutput{
		MessageID: "mock-" + time.Now().Format("20060102150405"),
		From:      input.From,
		To:        input.To,
		SentAt:    time.Now(),
		Provider:  email.ProviderMock,
	}

	s.emails = append(s.emails, output)
	return output, nil
}

func (s *MockEmailService) SendTemplate(ctx context.Context, templateID string, to []string, data map[string]any) (*email.SendOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.SendTemplateCalls = append(s.SendTemplateCalls, SendTemplateCall{
		TemplateID: templateID,
		To:         to,
		Data:       data,
	})

	if s.SendTemplateFunc != nil {
		return s.SendTemplateFunc(ctx, templateID, to, data)
	}

	tmpl, err := s.getTemplateLocked(templateID)
	if err != nil {
		return nil, err
	}

	return s.sendLocked(ctx, &email.SendInput{
		From:     tmpl.From,
		FromName: tmpl.FromName,
		To:       to,
		Subject:  tmpl.Subject,
		HTMLBody: tmpl.HTMLBody,
		TextBody: tmpl.TextBody,
	})
}

func (s *MockEmailService) Render(ctx context.Context, input *email.RenderInput) (*email.RenderOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.RenderCalls = append(s.RenderCalls, RenderCall{Input: input})

	if s.RenderFunc != nil {
		return s.RenderFunc(ctx, input)
	}

	tmpl, err := s.getTemplateLocked(input.TemplateID)
	if err != nil {
		return nil, err
	}

	return &email.RenderOutput{
		Subject:  tmpl.Subject,
		HTMLBody: tmpl.HTMLBody,
		TextBody: tmpl.TextBody,
	}, nil
}

func (s *MockEmailService) GetTemplate(ctx context.Context, templateID string) (*email.Template, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.GetTemplateCalls = append(s.GetTemplateCalls, GetTemplateCall{TemplateID: templateID})

	if s.GetTemplateFunc != nil {
		return s.GetTemplateFunc(ctx, templateID)
	}

	return s.getTemplateLocked(templateID)
}

func (s *MockEmailService) getTemplateLocked(templateID string) (*email.Template, error) {
	tmpl, ok := s.templates[templateID]
	if !ok {
		return nil, email.ErrTemplateNotFound
	}
	return tmpl, nil
}

func (s *MockEmailService) HealthCheck(ctx context.Context) error {
	if s.HealthCheckFunc != nil {
		return s.HealthCheckFunc(ctx)
	}
	return nil
}

func (s *MockEmailService) RegisterTemplate(tmpl *email.Template) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.templates[tmpl.ID] = tmpl
}

func (s *MockEmailService) GetEmails() []*email.SendOutput {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.emails
}

func (s *MockEmailService) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.emails = make([]*email.SendOutput, 0)
	s.templates = make(map[string]*email.Template)
	s.SendCalls = nil
	s.SendTemplateCalls = nil
	s.RenderCalls = nil
	s.GetTemplateCalls = nil
}

func (s *MockEmailService) AssertEmailSent(t assertT, to string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, email := range s.emails {
		for _, recipient := range email.To {
			if recipient == to {
				return
			}
		}
	}
	t.Errorf("expected email to be sent to %s, but it was not", to)
}

func (s *MockEmailService) AssertEmailCount(t assertT, count int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.emails) != count {
		t.Errorf("expected %d emails, got %d", count, len(s.emails))
	}
}

type assertT interface {
	Errorf(format string, args ...interface{})
}
