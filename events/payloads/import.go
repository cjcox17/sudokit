package payloads

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ImportStartedPayload struct {
	JobID        string             `json:"job_id"`
	FileName     string             `json:"file_name"`
	TotalRows    int                `json:"total_rows"`
	CreditorID   primitive.ObjectID `json:"creditor_id"`
	CreditorName string             `json:"creditor_name"`
}

type ImportProgressPayload struct {
	JobID          string `json:"job_id"`
	ProcessedRows  int    `json:"processed_rows"`
	TotalRows      int    `json:"total_rows"`
	Progress       int    `json:"progress"`
	CurrentMessage string `json:"current_message"`
}

type ImportCompletedPayload struct {
	JobID         string             `json:"job_id"`
	FileName      string             `json:"file_name"`
	TotalRows     int                `json:"total_rows"`
	ImportedCount int                `json:"imported_count"`
	ErrorCount    int                `json:"error_count"`
	CreditorID    primitive.ObjectID `json:"creditor_id"`
	CreditorName  string             `json:"creditor_name"`
	HasFailedRows bool               `json:"has_failed_rows"`
}

type ImportFailedPayload struct {
	JobID    string `json:"job_id"`
	FileName string `json:"file_name"`
	Error    string `json:"error"`
}

type JobStartedPayload struct {
	JobID   string `json:"job_id"`
	JobType string `json:"job_type"`
}

type JobProgressPayload struct {
	JobID    string `json:"job_id"`
	JobType  string `json:"job_type"`
	Progress int    `json:"progress"`
	Message  string `json:"message"`
}

type JobCompletedPayload struct {
	JobID   string         `json:"job_id"`
	JobType string         `json:"job_type"`
	Result  map[string]any `json:"result"`
}

type JobFailedPayload struct {
	JobID   string `json:"job_id"`
	JobType string `json:"job_type"`
	Error   string `json:"error"`
}
