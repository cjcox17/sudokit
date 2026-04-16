package testing

import (
	"context"
	"testing"
	"time"

	"github.com/cjcox17/sudokit/events"
	"github.com/cjcox17/sudokit/events/payloads"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMockEventsService_Publish(t *testing.T) {
	mock := NewMockEventsService()
	ctx := context.Background()

	err := mock.Publish(ctx, events.EventCaseFileCreated, map[string]any{
		"case_file_number": "CF-123",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(mock.PublishCalls) != 1 {
		t.Errorf("expected 1 publish call, got %d", len(mock.PublishCalls))
	}

	call := mock.PublishCalls[0]
	if call.EventType != events.EventCaseFileCreated {
		t.Errorf("expected event type '%s', got '%s'", events.EventCaseFileCreated, call.EventType)
	}
}

func TestMockEventsService_PublishTypedPayload(t *testing.T) {
	mock := NewMockEventsService()
	ctx := context.Background()

	payload := payloads.CaseFileCreatedPayload{
		CaseFileNumber: "CF-123",
		Status:         "NEW",
		CreditorID:     primitive.NewObjectID(),
		DebtorCount:    2,
	}

	err := mock.Publish(ctx, events.EventCaseFileCreated, payload)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(mock.PublishCalls) != 1 {
		t.Errorf("expected 1 publish call, got %d", len(mock.PublishCalls))
	}
}

func TestMockEventsService_Subscribe(t *testing.T) {
	mock := NewMockEventsService()

	err := mock.Subscribe(events.EventCaseFileCreated, func(ctx context.Context, e *events.BaseEvent) error {
		return nil
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(mock.SubscribeCalls) != 1 {
		t.Errorf("expected 1 subscribe call, got %d", len(mock.SubscribeCalls))
	}

	call := mock.SubscribeCalls[0]
	if call.EventType != events.EventCaseFileCreated {
		t.Errorf("expected event type '%s', got '%s'", events.EventCaseFileCreated, call.EventType)
	}
}

func TestMockEventsService_GetLastPublishPayload(t *testing.T) {
	mock := NewMockEventsService()
	ctx := context.Background()

	mock.Publish(ctx, "event1", map[string]any{"key": "value1"})
	mock.Publish(ctx, "event2", map[string]any{"key": "value2"})

	eventType, payload := mock.GetLastPublishPayload()

	if eventType != "event2" {
		t.Errorf("expected event type 'event2', got '%s'", eventType)
	}

	if payload["key"] != "value2" {
		t.Errorf("expected payload key 'value2', got '%v'", payload["key"])
	}
}

func TestMockEventsService_GetPublishCallsByType(t *testing.T) {
	mock := NewMockEventsService()
	ctx := context.Background()

	mock.Publish(ctx, "event1", nil)
	mock.Publish(ctx, "event2", nil)
	mock.Publish(ctx, "event1", nil)

	calls := mock.GetPublishCallsByType("event1")

	if len(calls) != 2 {
		t.Errorf("expected 2 calls for event1, got %d", len(calls))
	}
}

func TestMockEventsService_Reset(t *testing.T) {
	mock := NewMockEventsService()
	ctx := context.Background()

	mock.Publish(ctx, "event1", nil)
	mock.Subscribe("event1", nil)

	mock.Reset()

	if len(mock.PublishCalls) != 0 {
		t.Error("expected publish calls to be reset")
	}

	if len(mock.SubscribeCalls) != 0 {
		t.Error("expected subscribe calls to be reset")
	}
}

func TestMockBroadcaster_BroadcastToUser(t *testing.T) {
	mock := NewMockBroadcaster()
	userID := primitive.NewObjectID()

	mock.BroadcastToUser(userID, "test.event", map[string]any{"foo": "bar"})

	if len(mock.BroadcastToUserCalls) != 1 {
		t.Errorf("expected 1 broadcast call, got %d", len(mock.BroadcastToUserCalls))
	}

	call := mock.BroadcastToUserCalls[0]
	if call.UserID != userID {
		t.Error("expected user ID to match")
	}

	if call.Event != "test.event" {
		t.Errorf("expected event 'test.event', got '%s'", call.Event)
	}
}

func TestMockBroadcaster_BroadcastToOrg(t *testing.T) {
	mock := NewMockBroadcaster()
	orgID := primitive.NewObjectID()

	mock.BroadcastToOrg(orgID, "test.event", map[string]any{"foo": "bar"})

	if len(mock.BroadcastToOrgCalls) != 1 {
		t.Errorf("expected 1 broadcast call, got %d", len(mock.BroadcastToOrgCalls))
	}

	call := mock.BroadcastToOrgCalls[0]
	if call.OrgID != orgID {
		t.Error("expected org ID to match")
	}

	if call.Event != "test.event" {
		t.Errorf("expected event 'test.event', got '%s'", call.Event)
	}
}

func TestMockBroadcaster_Reset(t *testing.T) {
	mock := NewMockBroadcaster()

	mock.BroadcastToUser(primitive.NewObjectID(), "event", nil)
	mock.BroadcastToOrg(primitive.NewObjectID(), "event", nil)

	mock.Reset()

	if len(mock.BroadcastToUserCalls) != 0 {
		t.Error("expected broadcast to user calls to be reset")
	}

	if len(mock.BroadcastToOrgCalls) != 0 {
		t.Error("expected broadcast to org calls to be reset")
	}
}

func TestAssertEventPublished(t *testing.T) {
	mock := NewMockEventsService()
	ctx := context.Background()

	mock.Publish(ctx, events.EventCaseFileCreated, nil)

	AssertEventPublished(t, mock, events.EventCaseFileCreated)
}

func TestAssertCaseFileCreated(t *testing.T) {
	mock := NewMockEventsService()
	ctx := context.Background()

	mock.Publish(ctx, events.EventCaseFileCreated, nil)

	AssertCaseFileCreated(t, mock)
}

func TestAssertImportCompleted(t *testing.T) {
	mock := NewMockEventsService()
	ctx := context.Background()

	mock.Publish(ctx, events.EventImportCompleted, nil)

	AssertImportCompleted(t, mock)
}

func TestGetPayload(t *testing.T) {
	mock := NewMockEventsService()
	ctx := context.Background()

	expectedPayload := payloads.CaseFileCreatedPayload{
		CaseFileNumber: "CF-123",
		Status:         "NEW",
		CreditorID:     primitive.NewObjectID(),
		DebtorCount:    2,
	}

	mock.Publish(ctx, events.EventCaseFileCreated, expectedPayload)

	payload, ok := GetPayload[payloads.CaseFileCreatedPayload](mock)
	if !ok {
		t.Error("expected to get payload")
	}

	if payload.CaseFileNumber != "CF-123" {
		t.Errorf("expected case file number 'CF-123', got '%s'", payload.CaseFileNumber)
	}

	if payload.Status != "NEW" {
		t.Errorf("expected status 'NEW', got '%s'", payload.Status)
	}
}

func TestEventTypes(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"casefile.created", events.EventCaseFileCreated, "casefile.created"},
		{"casefile.updated", events.EventCaseFileUpdated, "casefile.updated"},
		{"casefile.deleted", events.EventCaseFileDeleted, "casefile.deleted"},
		{"payment.received", events.EventPaymentReceived, "payment.received"},
		{"user.created", events.EventUserCreated, "user.created"},
		{"import.completed", events.EventImportCompleted, "import.completed"},
		{"job.started", events.EventJobStarted, "job.started"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.constant)
			}
		})
	}
}

func TestNewEvent(t *testing.T) {
	payload := payloads.CaseFileCreatedPayload{
		CaseFileNumber: "CF-123",
		Status:         "NEW",
	}

	event := events.NewEvent(events.EventCaseFileCreated, payload)

	if event.EventType != events.EventCaseFileCreated {
		t.Errorf("expected type '%s', got '%s'", events.EventCaseFileCreated, event.EventType)
	}

	if event.Payload.CaseFileNumber != "CF-123" {
		t.Errorf("expected case file number 'CF-123', got '%s'", event.Payload.CaseFileNumber)
	}

	if event.Timestamp.IsZero() {
		t.Error("expected timestamp to be set")
	}
}

func TestEventWithUser(t *testing.T) {
	userID := primitive.NewObjectID().Hex()
	payload := payloads.CaseFileCreatedPayload{CaseFileNumber: "CF-123"}

	event := events.NewEvent(events.EventCaseFileCreated, payload,
		events.WithUser(userID, "testuser"))

	if event.UserID != userID {
		t.Error("expected user ID to be set")
	}
}

func TestEventWithOrganization(t *testing.T) {
	orgID := primitive.NewObjectID().Hex()
	payload := payloads.CaseFileCreatedPayload{CaseFileNumber: "CF-123"}

	event := events.NewEvent(events.EventCaseFileCreated, payload,
		events.WithOrganization(orgID))

	if event.OrganizationID != orgID {
		t.Error("expected organization ID to be set")
	}
}

func TestEventWithAggregate(t *testing.T) {
	aggregateID := primitive.NewObjectID().Hex()
	payload := payloads.CaseFileCreatedPayload{CaseFileNumber: "CF-123"}

	event := events.NewEvent(events.EventCaseFileCreated, payload,
		events.WithAggregate("CaseFile", aggregateID))

	if event.AggregateType != "CaseFile" {
		t.Errorf("expected aggregate type 'CaseFile', got '%s'", event.AggregateType)
	}

	if event.AggregateID != aggregateID {
		t.Error("expected aggregate ID to be set")
	}
}

func TestEventWithMetadata(t *testing.T) {
	metadata := map[string]any{"source": "test"}
	payload := payloads.CaseFileCreatedPayload{CaseFileNumber: "CF-123"}

	event := events.NewEvent(events.EventCaseFileCreated, payload,
		events.WithMetadata(metadata))

	if event.Metadata["source"] != "test" {
		t.Error("expected metadata to be set")
	}
}

func TestEventWithRequestID(t *testing.T) {
	requestID := "req-123"
	payload := payloads.CaseFileCreatedPayload{CaseFileNumber: "CF-123"}

	event := events.NewEvent(events.EventCaseFileCreated, payload,
		events.WithRequestID(requestID))

	if event.RequestID != requestID {
		t.Error("expected request_id to be set")
	}
}

func TestBaseEventWithUser(t *testing.T) {
	userID := primitive.NewObjectID().Hex()

	event := events.NewBaseEvent(
		events.EventCaseFileCreated,
		"CaseFile",
		primitive.NewObjectID().Hex(),
		map[string]any{},
		events.WithUser(userID, "testuser"),
	)

	if event.UserID != userID {
		t.Error("expected user ID to be set")
	}
}

func TestBaseEventWithPayload(t *testing.T) {
	payload := map[string]any{"case_file_number": "CF-123"}

	event := events.NewBaseEvent(
		events.EventCaseFileCreated,
		"CaseFile",
		primitive.NewObjectID().Hex(),
		payload,
	)

	if event.Payload["case_file_number"] != "CF-123" {
		t.Error("expected payload to be set")
	}
}

func TestEventToBaseEvent(t *testing.T) {
	userID := primitive.NewObjectID().Hex()
	orgID := primitive.NewObjectID().Hex()
	aggregateID := primitive.NewObjectID().Hex()

	payload := payloads.CaseFileCreatedPayload{
		CaseFileNumber: "CF-123",
		Status:         "NEW",
	}

	typedEvent := events.NewEvent(events.EventCaseFileCreated, payload,
		events.WithUser(userID, "testuser"),
		events.WithOrganization(orgID),
		events.WithAggregate("CaseFile", aggregateID),
	)

	baseEvent := typedEvent.ToBaseEvent()

	if baseEvent.EventType != events.EventCaseFileCreated {
		t.Errorf("expected type '%s', got '%s'", events.EventCaseFileCreated, baseEvent.EventType)
	}

	if baseEvent.UserID != userID {
		t.Error("expected user ID to match")
	}

	if baseEvent.OrganizationID != orgID {
		t.Error("expected organization ID to match")
	}
}

func TestImportPayloads(t *testing.T) {
	t.Run("ImportStartedPayload", func(t *testing.T) {
		payload := payloads.ImportStartedPayload{
			JobID:      "job-123",
			FileName:   "test.csv",
			TotalRows:  1000,
			CreditorID: primitive.NewObjectID(),
		}

		if payload.JobID != "job-123" {
			t.Error("expected job ID to match")
		}
	})

	t.Run("ImportProgressPayload", func(t *testing.T) {
		payload := payloads.ImportProgressPayload{
			JobID:         "job-123",
			ProcessedRows: 500,
			TotalRows:     1000,
			Progress:      50,
		}

		if payload.Progress != 50 {
			t.Errorf("expected progress 50, got %d", payload.Progress)
		}
	})

	t.Run("ImportCompletedPayload", func(t *testing.T) {
		payload := payloads.ImportCompletedPayload{
			JobID:         "job-123",
			TotalRows:     1000,
			ImportedCount: 950,
			ErrorCount:    50,
			HasFailedRows: true,
		}

		if payload.ImportedCount != 950 {
			t.Errorf("expected imported count 950, got %d", payload.ImportedCount)
		}
	})
}

func TestPaymentPayloads(t *testing.T) {
	t.Run("PaymentReceivedPayload", func(t *testing.T) {
		payload := payloads.PaymentReceivedPayload{
			CaseFileNumber: "CF-123",
			Amount:         100.50,
			PaymentMethod:  "credit_card",
			ReceivedAt:     time.Now(),
		}

		if payload.Amount != 100.50 {
			t.Errorf("expected amount 100.50, got %f", payload.Amount)
		}
	})
}

func TestUserPayloads(t *testing.T) {
	t.Run("UserCreatedPayload", func(t *testing.T) {
		payload := payloads.UserCreatedPayload{
			UserID: primitive.NewObjectID(),
			Email:  "test@example.com",
			Name:   "Test User",
			Role:   "admin",
		}

		if payload.Email != "test@example.com" {
			t.Errorf("expected email 'test@example.com', got '%s'", payload.Email)
		}
	})

	t.Run("UserMentionPayload", func(t *testing.T) {
		payload := payloads.UserMentionPayload{
			MentionedUserID:   primitive.NewObjectID(),
			MentionedByUserID: primitive.NewObjectID(),
			Context:           "case file comment",
			EntityType:        "CaseFile",
		}

		if payload.Context != "case file comment" {
			t.Errorf("expected context 'case file comment', got '%s'", payload.Context)
		}
	})
}

func TestWebhookPayloads(t *testing.T) {
	t.Run("ESignatureCompletedPayload", func(t *testing.T) {
		payload := payloads.ESignatureCompletedPayload{
			DocumentTitle:  "Agreement",
			CaseFileNumber: "CF-123",
			SignerName:     "John Doe",
			SignerEmail:    "john@example.com",
			CompletedAt:    time.Now(),
			Provider:       "DocuSign",
		}

		if payload.Provider != "DocuSign" {
			t.Errorf("expected provider 'DocuSign', got '%s'", payload.Provider)
		}
	})

	t.Run("BulkActionPayload", func(t *testing.T) {
		payload := payloads.BulkActionPayload{
			ActionType:     "status_change",
			TotalItems:     100,
			ProcessedItems: 50,
			SuccessCount:   48,
			ErrorCount:     2,
			Progress:       50,
			Status:         "in_progress",
		}

		if payload.Progress != 50 {
			t.Errorf("expected progress 50, got %d", payload.Progress)
		}
	})
}
