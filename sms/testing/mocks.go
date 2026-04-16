package testing

import (
	"context"
	"sync"
	"time"

	"github.com/cjcox17/sudokit/sms"
)

type MockSMSService struct {
	mu sync.Mutex

	SendFunc         func(ctx context.Context, input *sms.SendInput) (*sms.SendOutput, error)
	SendTemplateFunc func(ctx context.Context, templateID string, to string, data map[string]any) (*sms.SendOutput, error)
	GetTemplateFunc  func(ctx context.Context, templateID string) (*sms.Template, error)
	HealthCheckFunc  func(ctx context.Context) error

	SendCalls         []SendCall
	SendTemplateCalls []SendTemplateCall
	GetTemplateCalls  []GetTemplateCall

	messages  []*sms.SendOutput
	templates map[string]*sms.Template
}

type SendCall struct {
	Input *sms.SendInput
}

type SendTemplateCall struct {
	TemplateID string
	To         string
	Data       map[string]any
}

type GetTemplateCall struct {
	TemplateID string
}

func NewMockSMSService() *MockSMSService {
	return &MockSMSService{
		messages:  make([]*sms.SendOutput, 0),
		templates: make(map[string]*sms.Template),
	}
}

func (s *MockSMSService) Send(ctx context.Context, input *sms.SendInput) (*sms.SendOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.sendLocked(ctx, input)
}

func (s *MockSMSService) sendLocked(ctx context.Context, input *sms.SendInput) (*sms.SendOutput, error) {
	s.SendCalls = append(s.SendCalls, SendCall{Input: input})

	if s.SendFunc != nil {
		return s.SendFunc(ctx, input)
	}

	if err := input.Validate(); err != nil {
		return nil, err
	}

	output := &sms.SendOutput{
		MessageID: "mock-" + time.Now().Format("20060102150405"),
		From:      input.From,
		To:        input.To,
		Body:      input.Message,
		SentAt:    time.Now(),
		Provider:  sms.ProviderMock,
		Status:    "sent",
	}

	s.messages = append(s.messages, output)
	return output, nil
}

func (s *MockSMSService) SendTemplate(ctx context.Context, templateID string, to string, data map[string]any) (*sms.SendOutput, error) {
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

	return s.sendLocked(ctx, &sms.SendInput{
		To:      to,
		Message: tmpl.Body,
	})
}

func (s *MockSMSService) GetTemplate(ctx context.Context, templateID string) (*sms.Template, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.GetTemplateCalls = append(s.GetTemplateCalls, GetTemplateCall{TemplateID: templateID})

	if s.GetTemplateFunc != nil {
		return s.GetTemplateFunc(ctx, templateID)
	}

	return s.getTemplateLocked(templateID)
}

func (s *MockSMSService) getTemplateLocked(templateID string) (*sms.Template, error) {
	tmpl, ok := s.templates[templateID]
	if !ok {
		return nil, sms.ErrTemplateNotFound
	}
	return tmpl, nil
}

func (s *MockSMSService) HealthCheck(ctx context.Context) error {
	if s.HealthCheckFunc != nil {
		return s.HealthCheckFunc(ctx)
	}
	return nil
}

func (s *MockSMSService) RegisterTemplate(tmpl *sms.Template) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.templates[tmpl.ID] = tmpl
}

func (s *MockSMSService) GetMessages() []*sms.SendOutput {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.messages
}

func (s *MockSMSService) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = make([]*sms.SendOutput, 0)
	s.templates = make(map[string]*sms.Template)
	s.SendCalls = nil
	s.SendTemplateCalls = nil
	s.GetTemplateCalls = nil
}

func (s *MockSMSService) AssertSMSSent(t assertT, to string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, msg := range s.messages {
		if msg.To == to {
			return
		}
	}
	t.Errorf("expected SMS to be sent to %s, but it was not", to)
}

func (s *MockSMSService) AssertSMSCount(t assertT, count int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.messages) != count {
		t.Errorf("expected %d SMS messages, got %d", count, len(s.messages))
	}
}

type assertT interface {
	Errorf(format string, args ...interface{})
}
