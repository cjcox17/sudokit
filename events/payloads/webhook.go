package payloads

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ESignatureCompletedPayload struct {
	DocumentID     primitive.ObjectID `json:"document_id"`
	DocumentTitle  string             `json:"document_title"`
	CaseFileID     primitive.ObjectID `json:"case_file_id"`
	CaseFileNumber string             `json:"case_file_number"`
	SignerID       primitive.ObjectID `json:"signer_id"`
	SignerName     string             `json:"signer_name"`
	SignerEmail    string             `json:"signer_email"`
	CompletedAt    time.Time          `json:"completed_at"`
	Provider       string             `json:"provider"`
}

type ESignatureDeclinedPayload struct {
	DocumentID     primitive.ObjectID `json:"document_id"`
	DocumentTitle  string             `json:"document_title"`
	CaseFileID     primitive.ObjectID `json:"case_file_id"`
	CaseFileNumber string             `json:"case_file_number"`
	SignerID       primitive.ObjectID `json:"signer_id"`
	SignerName     string             `json:"signer_name"`
	SignerEmail    string             `json:"signer_email"`
	Reason         string             `json:"reason"`
	DeclinedAt     time.Time          `json:"declined_at"`
	Provider       string             `json:"provider"`
}

type CallReceivedPayload struct {
	CallID         primitive.ObjectID  `json:"call_id"`
	FromNumber     string              `json:"from_number"`
	ToNumber       string              `json:"to_number"`
	Duration       int                 `json:"duration"`
	RecordingURL   string              `json:"recording_url,omitempty"`
	Transcription  string              `json:"transcription,omitempty"`
	CaseFileID     *primitive.ObjectID `json:"case_file_id,omitempty"`
	CaseFileNumber string              `json:"case_file_number,omitempty"`
	ReceivedAt     time.Time           `json:"received_at"`
	Provider       string              `json:"provider"`
}

type SMSReceivedPayload struct {
	SMSID          primitive.ObjectID  `json:"sms_id"`
	FromNumber     string              `json:"from_number"`
	ToNumber       string              `json:"to_number"`
	Message        string              `json:"message"`
	CaseFileID     *primitive.ObjectID `json:"case_file_id,omitempty"`
	CaseFileNumber string              `json:"case_file_number,omitempty"`
	ReceivedAt     time.Time           `json:"received_at"`
	Provider       string              `json:"provider"`
}

type BulkActionPayload struct {
	ActionID        primitive.ObjectID `json:"action_id"`
	ActionType      string             `json:"action_type"`
	TotalItems      int                `json:"total_items"`
	ProcessedItems  int                `json:"processed_items"`
	SuccessCount    int                `json:"success_count"`
	ErrorCount      int                `json:"error_count"`
	Progress        int                `json:"progress"`
	Status          string             `json:"status"`
	InitiatedBy     primitive.ObjectID `json:"initiated_by"`
	InitiatedByName string             `json:"initiated_by_name"`
}

type ErrorPayload struct {
	ErrorID        primitive.ObjectID  `json:"error_id"`
	ErrorCode      string              `json:"error_code"`
	ErrorMessage   string              `json:"error_message"`
	ErrorDetails   map[string]any      `json:"error_details,omitempty"`
	EntityID       *primitive.ObjectID `json:"entity_id,omitempty"`
	EntityType     string              `json:"entity_type"`
	UserID         *primitive.ObjectID `json:"user_id,omitempty"`
	OrganizationID primitive.ObjectID  `json:"organization_id"`
	OccurredAt     time.Time           `json:"occurred_at"`
	Severity       string              `json:"severity"`
}
