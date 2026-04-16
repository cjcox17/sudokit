package testing

import (
	"context"
	"testing"

	"github.com/cjcox17/sudokit/sms"
)

func TestMockSMSService_Send(t *testing.T) {
	svc := NewMockSMSService()

	output, err := svc.Send(context.Background(), &sms.SendInput{
		From:    "+15555551234",
		To:      "+15555559876",
		Message: "Test message",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if output.MessageID == "" {
		t.Error("expected message ID")
	}

	if output.To != "+15555559876" {
		t.Errorf("expected to +15555559876, got %s", output.To)
	}

	messages := svc.GetMessages()
	if len(messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(messages))
	}
}

func TestMockSMSService_SendTemplate(t *testing.T) {
	svc := NewMockSMSService()

	svc.RegisterTemplate(&sms.Template{
		ID:   "welcome",
		Body: "Welcome to our service!",
	})

	output, err := svc.SendTemplate(context.Background(), "welcome", "+15555559876", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(svc.SendTemplateCalls) != 1 {
		t.Errorf("expected 1 SendTemplate call, got %d", len(svc.SendTemplateCalls))
	}

	if output.Body != "Welcome to our service!" {
		t.Errorf("expected body 'Welcome to our service!', got %s", output.Body)
	}
}

func TestMockSMSService_Validation(t *testing.T) {
	svc := NewMockSMSService()

	_, err := svc.Send(context.Background(), &sms.SendInput{
		To:      "",
		Message: "Test",
	})
	if err != sms.ErrInvalidTo {
		t.Errorf("expected ErrInvalidTo, got %v", err)
	}

	_, err = svc.Send(context.Background(), &sms.SendInput{
		To: "+15555559876",
	})
	if err != sms.ErrInvalidMessage {
		t.Errorf("expected ErrInvalidMessage, got %v", err)
	}
}

func TestMockSMSService_Assertions(t *testing.T) {
	svc := NewMockSMSService()

	svc.Send(context.Background(), &sms.SendInput{
		To:      "+15555551111",
		Message: "Test 1",
	})

	svc.Send(context.Background(), &sms.SendInput{
		To:      "+15555552222",
		Message: "Test 2",
	})

	svc.AssertSMSSent(t, "+15555551111")
	svc.AssertSMSCount(t, 2)
}

func TestMockSMSService_Reset(t *testing.T) {
	svc := NewMockSMSService()

	svc.Send(context.Background(), &sms.SendInput{
		To:      "+15555559876",
		Message: "Test",
	})

	if len(svc.GetMessages()) != 1 {
		t.Error("expected 1 message before reset")
	}

	svc.Reset()

	if len(svc.GetMessages()) != 0 {
		t.Error("expected 0 messages after reset")
	}
}

func TestMockSMSService_TemplateNotFound(t *testing.T) {
	svc := NewMockSMSService()

	_, err := svc.GetTemplate(context.Background(), "nonexistent")
	if err != sms.ErrTemplateNotFound {
		t.Errorf("expected ErrTemplateNotFound, got %v", err)
	}
}

func TestSMSProviderConstants(t *testing.T) {
	if sms.ProviderTwilio != "twilio" {
		t.Error("expected ProviderTwilio to be 'twilio'")
	}
	if sms.ProviderMock != "mock" {
		t.Error("expected ProviderMock to be 'mock'")
	}
}

func TestSendInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   *sms.SendInput
		wantErr error
	}{
		{
			name: "valid input",
			input: &sms.SendInput{
				To:      "+15555559876",
				Message: "Test message",
			},
			wantErr: nil,
		},
		{
			name: "missing to",
			input: &sms.SendInput{
				Message: "Test message",
			},
			wantErr: sms.ErrInvalidTo,
		},
		{
			name: "missing message",
			input: &sms.SendInput{
				To: "+15555559876",
			},
			wantErr: sms.ErrInvalidMessage,
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
