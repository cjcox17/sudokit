package testing

import (
	"context"
	"testing"
	"time"

	"github.com/cjcox17/sudokit/email"
)

func TestMockEmailService_Send(t *testing.T) {
	svc := NewMockEmailService()

	output, err := svc.Send(context.Background(), &email.SendInput{
		From:     "test@example.com",
		To:       []string{"recipient@example.com"},
		Subject:  "Test Subject",
		HTMLBody: "<p>Test body</p>",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if output.MessageID == "" {
		t.Error("expected message ID")
	}

	if output.From != "test@example.com" {
		t.Errorf("expected from test@example.com, got %s", output.From)
	}

	emails := svc.GetEmails()
	if len(emails) != 1 {
		t.Errorf("expected 1 email, got %d", len(emails))
	}
}

func TestMockEmailService_SendTemplate(t *testing.T) {
	svc := NewMockEmailService()

	svc.RegisterTemplate(&email.Template{
		ID:       "welcome",
		Subject:  "Welcome!",
		HTMLBody: "<p>Welcome to our service</p>",
		From:     "noreply@example.com",
	})

	output, err := svc.SendTemplate(context.Background(), "welcome", []string{"user@example.com"}, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(svc.SendTemplateCalls) != 1 {
		t.Errorf("expected 1 SendTemplate call, got %d", len(svc.SendTemplateCalls))
	}

	if output.From != "noreply@example.com" {
		t.Errorf("expected from noreply@example.com, got %s", output.From)
	}
}

func TestMockEmailService_Validation(t *testing.T) {
	svc := NewMockEmailService()

	_, err := svc.Send(context.Background(), &email.SendInput{
		From: "",
		To:   []string{"recipient@example.com"},
	})
	if err != email.ErrInvalidFrom {
		t.Errorf("expected ErrInvalidFrom, got %v", err)
	}

	_, err = svc.Send(context.Background(), &email.SendInput{
		From: "test@example.com",
		To:   []string{},
	})
	if err != email.ErrInvalidTo {
		t.Errorf("expected ErrInvalidTo, got %v", err)
	}

	_, err = svc.Send(context.Background(), &email.SendInput{
		From:    "test@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "",
	})
	if err != email.ErrInvalidSubject {
		t.Errorf("expected ErrInvalidSubject, got %v", err)
	}
}

func TestMockEmailService_Assertions(t *testing.T) {
	svc := NewMockEmailService()

	svc.Send(context.Background(), &email.SendInput{
		From:     "test@example.com",
		To:       []string{"user1@example.com", "user2@example.com"},
		Subject:  "Test",
		HTMLBody: "<p>Body</p>",
	})

	svc.AssertEmailSent(t, "user1@example.com")
	svc.AssertEmailCount(t, 1)
}

func TestMockEmailService_Reset(t *testing.T) {
	svc := NewMockEmailService()

	svc.Send(context.Background(), &email.SendInput{
		From:     "test@example.com",
		To:       []string{"user@example.com"},
		Subject:  "Test",
		HTMLBody: "<p>Body</p>",
	})

	if len(svc.GetEmails()) != 1 {
		t.Error("expected 1 email before reset")
	}

	svc.Reset()

	if len(svc.GetEmails()) != 0 {
		t.Error("expected 0 emails after reset")
	}
}

func TestMockEmailService_CustomBehavior(t *testing.T) {
	svc := NewMockEmailService()

	svc.SendFunc = func(ctx context.Context, input *email.SendInput) (*email.SendOutput, error) {
		return &email.SendOutput{
			MessageID: "custom-id",
			From:      input.From,
			To:        input.To,
			SentAt:    time.Now(),
			Provider:  email.ProviderMock,
		}, nil
	}

	output, err := svc.Send(context.Background(), &email.SendInput{
		From:     "test@example.com",
		To:       []string{"user@example.com"},
		Subject:  "Test",
		HTMLBody: "<p>Body</p>",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if output.MessageID != "custom-id" {
		t.Errorf("expected custom-id, got %s", output.MessageID)
	}
}

func TestSendInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   *email.SendInput
		wantErr error
	}{
		{
			name: "valid input",
			input: &email.SendInput{
				From:     "test@example.com",
				To:       []string{"recipient@example.com"},
				Subject:  "Test",
				HTMLBody: "<p>Body</p>",
			},
			wantErr: nil,
		},
		{
			name: "missing from",
			input: &email.SendInput{
				To:       []string{"recipient@example.com"},
				Subject:  "Test",
				HTMLBody: "<p>Body</p>",
			},
			wantErr: email.ErrInvalidFrom,
		},
		{
			name: "missing to",
			input: &email.SendInput{
				From:     "test@example.com",
				Subject:  "Test",
				HTMLBody: "<p>Body</p>",
			},
			wantErr: email.ErrInvalidTo,
		},
		{
			name: "missing subject",
			input: &email.SendInput{
				From:     "test@example.com",
				To:       []string{"recipient@example.com"},
				HTMLBody: "<p>Body</p>",
			},
			wantErr: email.ErrInvalidSubject,
		},
		{
			name: "missing body",
			input: &email.SendInput{
				From:    "test@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test",
			},
			wantErr: email.ErrInvalidBody,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if err != tt.wantErr {
				t.Errorf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestProviderConstants(t *testing.T) {
	if email.ProviderMailgun != "mailgun" {
		t.Error("expected ProviderMailgun to be 'mailgun'")
	}
	if email.ProviderSMTP != "smtp" {
		t.Error("expected ProviderSMTP to be 'smtp'")
	}
	if email.ProviderMock != "mock" {
		t.Error("expected ProviderMock to be 'mock'")
	}
}

func TestAttachment(t *testing.T) {
	att := email.Attachment{
		Filename:    "test.pdf",
		ContentType: "application/pdf",
		Data:        []byte("test data"),
	}

	if att.Filename != "test.pdf" {
		t.Error("expected filename test.pdf")
	}
}

func TestTemplate(t *testing.T) {
	tmpl := email.Template{
		ID:        "welcome",
		Name:      "Welcome Email",
		Subject:   "Welcome!",
		HTMLBody:  "<p>Welcome</p>",
		TextBody:  "Welcome",
		From:      "noreply@example.com",
		FromName:  "My App",
		Variables: []string{"name", "email"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if tmpl.ID != "welcome" {
		t.Error("expected template ID welcome")
	}
}

func TestMockEmailService_TemplateNotFound(t *testing.T) {
	svc := NewMockEmailService()

	_, err := svc.GetTemplate(context.Background(), "nonexistent")
	if err != email.ErrTemplateNotFound {
		t.Errorf("expected ErrTemplateNotFound, got %v", err)
	}
}
