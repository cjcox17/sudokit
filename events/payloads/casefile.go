package payloads

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CaseFileCreatedPayload struct {
	CaseFileNumber string             `json:"case_file_number"`
	Status         string             `json:"status"`
	CreditorID     primitive.ObjectID `json:"creditor_id"`
	DebtorCount    int                `json:"debtor_count"`
}

type CaseFileUpdatedPayload struct {
	CaseFileID     primitive.ObjectID     `json:"case_file_id"`
	CaseFileNumber string                 `json:"case_file_number"`
	Changes        map[string]FieldChange `json:"changes"`
}

type CaseFileDeletedPayload struct {
	CaseFileID     primitive.ObjectID `json:"case_file_id"`
	CaseFileNumber string             `json:"case_file_number"`
}

type CaseFileViewedPayload struct {
	CaseFileID     primitive.ObjectID `json:"case_file_id"`
	CaseFileNumber string             `json:"case_file_number"`
	ViewerID       primitive.ObjectID `json:"viewer_id"`
	ViewerName     string             `json:"viewer_name"`
}

type CaseFileAssignedPayload struct {
	CaseFileID       primitive.ObjectID  `json:"case_file_id"`
	CaseFileNumber   string              `json:"case_file_number"`
	OldCollectorID   *primitive.ObjectID `json:"old_collector_id,omitempty"`
	NewCollectorID   primitive.ObjectID  `json:"new_collector_id"`
	OldCollectorName string              `json:"old_collector_name,omitempty"`
	NewCollectorName string              `json:"new_collector_name"`
}

type CaseFileStatusChangedPayload struct {
	CaseFileID     primitive.ObjectID `json:"case_file_id"`
	CaseFileNumber string             `json:"case_file_number"`
	OldStatus      string             `json:"old_status"`
	NewStatus      string             `json:"new_status"`
}

type FieldChange struct {
	Old any `json:"old"`
	New any `json:"new"`
}

type DebtorCreatedPayload struct {
	CaseFileID primitive.ObjectID `json:"case_file_id"`
	DebtorID   primitive.ObjectID `json:"debtor_id"`
	FirstName  string             `json:"first_name"`
	LastName   string             `json:"last_name"`
}

type DebtorUpdatedPayload struct {
	CaseFileID primitive.ObjectID     `json:"case_file_id"`
	DebtorID   primitive.ObjectID     `json:"debtor_id"`
	Changes    map[string]FieldChange `json:"changes"`
}

type DebtorDeletedPayload struct {
	CaseFileID primitive.ObjectID `json:"case_file_id"`
	DebtorID   primitive.ObjectID `json:"debtor_id"`
}
