package payloads

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PaymentReceivedPayload struct {
	CaseFileID     primitive.ObjectID `json:"case_file_id"`
	CaseFileNumber string             `json:"case_file_number"`
	PaymentID      primitive.ObjectID `json:"payment_id"`
	Amount         float64            `json:"amount"`
	PaymentMethod  string             `json:"payment_method"`
	ReceivedAt     time.Time          `json:"received_at"`
	CollectorID    primitive.ObjectID `json:"collector_id"`
	CollectorName  string             `json:"collector_name"`
}

type PaymentFailedPayload struct {
	CaseFileID     primitive.ObjectID `json:"case_file_id"`
	CaseFileNumber string             `json:"case_file_number"`
	Amount         float64            `json:"amount"`
	PaymentMethod  string             `json:"payment_method"`
	Reason         string             `json:"reason"`
}

type PaymentPlanCreatedPayload struct {
	CaseFileID     primitive.ObjectID `json:"case_file_id"`
	CaseFileNumber string             `json:"case_file_number"`
	PlanID         primitive.ObjectID `json:"plan_id"`
	TotalAmount    float64            `json:"total_amount"`
	Installments   int                `json:"installments"`
	StartDate      time.Time          `json:"start_date"`
	EndDate        time.Time          `json:"end_date"`
}
