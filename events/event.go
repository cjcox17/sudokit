package events

import (
	"context"
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EventOption func(*BaseEvent)

func WithUser(userID, userName string) EventOption {
	return func(e *BaseEvent) {
		e.UserID = userID
		e.UserName = userName
	}
}

func WithOrganization(orgID string) EventOption {
	return func(e *BaseEvent) {
		e.OrganizationID = orgID
	}
}

func WithAggregate(aggregateType, aggregateID string) EventOption {
	return func(e *BaseEvent) {
		e.AggregateType = aggregateType
		e.AggregateID = aggregateID
	}
}

func WithMetadata(metadata map[string]any) EventOption {
	return func(e *BaseEvent) {
		if e.Metadata == nil {
			e.Metadata = make(map[string]any)
		}
		for k, v := range metadata {
			e.Metadata[k] = v
		}
	}
}

func WithMetadataValue(key string, value any) EventOption {
	return func(e *BaseEvent) {
		if e.Metadata == nil {
			e.Metadata = make(map[string]any)
		}
		e.Metadata[key] = value
	}
}

func WithRequestID(requestID string) EventOption {
	return func(e *BaseEvent) {
		e.RequestID = requestID
	}
}

func WithTimestamp(timestamp time.Time) EventOption {
	return func(e *BaseEvent) {
		e.Timestamp = timestamp
	}
}

type Event[T any] struct {
	ID             primitive.ObjectID `json:"id"`
	EventType      string             `json:"event_type"`
	AggregateType  string             `json:"aggregate_type"`
	AggregateID    string             `json:"aggregate_id"`
	UserID         string             `json:"user_id,omitempty"`
	UserName       string             `json:"user_name,omitempty"`
	OrganizationID string             `json:"organization_id,omitempty"`
	RequestID      string             `json:"request_id,omitempty"`
	Payload        T                  `json:"payload"`
	Metadata       map[string]any     `json:"metadata,omitempty"`
	Timestamp      time.Time          `json:"timestamp"`
	ProcessedAt    *time.Time         `json:"processed_at,omitempty"`
	ProcessedBy    []string           `json:"processed_by,omitempty"`
}

type BaseEvent struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	EventType      string             `bson:"event_type" json:"event_type"`
	AggregateType  string             `bson:"aggregate_type" json:"aggregate_type"`
	AggregateID    string             `bson:"aggregate_id" json:"aggregate_id"`
	UserID         string             `bson:"user_id,omitempty" json:"user_id,omitempty"`
	UserName       string             `bson:"user_name,omitempty" json:"user_name,omitempty"`
	OrganizationID string             `bson:"organization_id,omitempty" json:"organization_id,omitempty"`
	RequestID      string             `bson:"request_id,omitempty" json:"request_id,omitempty"`
	Payload        map[string]any     `bson:"payload" json:"payload"`
	Metadata       map[string]any     `bson:"metadata,omitempty" json:"metadata,omitempty"`
	Timestamp      time.Time          `bson:"timestamp" json:"timestamp"`
	ProcessedAt    *time.Time         `bson:"processed_at,omitempty" json:"processed_at,omitempty"`
	ProcessedBy    []string           `bson:"processed_by,omitempty" json:"processed_by,omitempty"`
}

func NewEvent[T any](eventType string, payload T, opts ...EventOption) *Event[T] {
	e := &Event[T]{
		ID:        primitive.NewObjectID(),
		EventType: eventType,
		Payload:   payload,
		Metadata:  make(map[string]any),
		Timestamp: time.Now(),
	}
	base := &BaseEvent{
		ID:        e.ID,
		EventType: e.EventType,
		Metadata:  e.Metadata,
		Timestamp: e.Timestamp,
	}
	for _, opt := range opts {
		opt(base)
	}
	e.AggregateType = base.AggregateType
	e.AggregateID = base.AggregateID
	e.UserID = base.UserID
	e.UserName = base.UserName
	e.OrganizationID = base.OrganizationID
	e.RequestID = base.RequestID
	return e
}

func NewBaseEvent(eventType, aggregateType, aggregateID string, payload map[string]any, opts ...EventOption) *BaseEvent {
	e := &BaseEvent{
		ID:            primitive.NewObjectID(),
		EventType:     eventType,
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		Payload:       payload,
		Metadata:      make(map[string]any),
		Timestamp:     time.Now(),
		ProcessedBy:   []string{},
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *BaseEvent) Publish(svc *Service) error {
	if svc == nil {
		return nil
	}
	return svc.Publish(nil, e.EventType, e.Payload,
		WithAggregate(e.AggregateType, e.AggregateID),
		WithUser(e.UserID, e.UserName),
		WithOrganization(e.OrganizationID),
		WithRequestID(e.RequestID),
		WithMetadata(e.Metadata),
	)
}

func (e *BaseEvent) PublishGlobal() error {
	svc := GetService()
	if svc == nil {
		return nil
	}
	return e.Publish(svc)
}

func (e *BaseEvent) MarshalJSON() ([]byte, error) {
	type Alias BaseEvent
	return json.Marshal(&struct {
		*Alias
		ID string `json:"id"`
	}{
		Alias: (*Alias)(e),
		ID:    e.ID.Hex(),
	})
}

func (e *Event[T]) ToBaseEvent() *BaseEvent {
	var payloadMap map[string]any
	if p, ok := any(e.Payload).(map[string]any); ok {
		payloadMap = p
	}

	return &BaseEvent{
		ID:             e.ID,
		EventType:      e.EventType,
		AggregateType:  e.AggregateType,
		AggregateID:    e.AggregateID,
		UserID:         e.UserID,
		UserName:       e.UserName,
		OrganizationID: e.OrganizationID,
		RequestID:      e.RequestID,
		Payload:        payloadMap,
		Metadata:       e.Metadata,
		Timestamp:      e.Timestamp,
		ProcessedAt:    e.ProcessedAt,
		ProcessedBy:    e.ProcessedBy,
	}
}

type ServiceInterface interface {
	Start() error
	Stop(ctx context.Context) error
	Publish(ctx context.Context, eventType string, payload any, opts ...EventOption) error
	Subscribe(eventType string, handler Handler) error
	SubscribeSync(eventType string, handler Handler) error
}

const (
	EventUserCreated       = "user.created"
	EventUserUpdated       = "user.updated"
	EventUserDeleted       = "user.deleted"
	EventUserDisabled      = "user.disabled"
	EventUserEnabled       = "user.enabled"
	EventUserPasswordReset = "user.password_reset"
	EventUserLoggedIn      = "user.logged_in"
	EventUserLoggedOut     = "user.logged_out"
	EventUserWentAway      = "user.went_away"
	EventUserCameBack      = "user.came_back"
	EventUserIPAdded       = "user.ip.added"
	EventUserIPRemoved     = "user.ip.removed"
	EventUserRestored      = "user.restored"

	EventOrganizationCreated = "organization.created"
	EventOrganizationUpdated = "organization.updated"
	EventOrganizationDeleted = "organization.deleted"

	EventCompanyCreated = "company.created"
	EventCompanyUpdated = "company.updated"
	EventCompanyDeleted = "company.deleted"

	EventCreditorCreated = "creditor.created"
	EventCreditorUpdated = "creditor.updated"
	EventCreditorDeleted = "creditor.deleted"

	EventMerchantCreated = "merchant.created"
	EventMerchantUpdated = "merchant.updated"
	EventMerchantDeleted = "merchant.deleted"

	EventCaseFileCreated = "casefile.created"
	EventCaseFileUpdated = "casefile.updated"
	EventCaseFileDeleted = "casefile.deleted"
	EventCaseFileViewed  = "casefile.viewed"

	EventCaseFileDebtorAdded          = "casefile.debtor.added"
	EventCaseFileDebtorUpdated        = "casefile.debtor.updated"
	EventCaseFileDebtorDeleted        = "casefile.debtor.deleted"
	EventCaseFileDebtorAddressAdded   = "casefile.debtor.address.added"
	EventCaseFileDebtorAddressUpdated = "casefile.debtor.address.updated"
	EventCaseFileDebtorAddressDeleted = "casefile.debtor.address.deleted"
	EventCaseFileDebtorPhoneAdded     = "casefile.debtor.phone.added"
	EventCaseFileDebtorPhoneUpdated   = "casefile.debtor.phone.updated"
	EventCaseFileDebtorPhoneDeleted   = "casefile.debtor.phone.deleted"
	EventCaseFileDebtorEmailAdded     = "casefile.debtor.email.added"
	EventCaseFileDebtorEmailUpdated   = "casefile.debtor.email.updated"
	EventCaseFileDebtorEmailDeleted   = "casefile.debtor.email.deleted"
	EventCaseFileDebtorTagAdded       = "casefile.debtor.tag.added"
	EventCaseFileDebtorTagRemoved     = "casefile.debtor.tag.removed"

	EventCaseFilePhoneAdded         = "casefile.phone.added"
	EventCaseFilePhoneStatusUpdated = "casefile.phone.status.updated"
	EventCaseFilePhoneRemoved       = "casefile.phone.removed"
	EventCaseFilePhoneMovedToDebtor = "casefile.phone.moved_to_debtor"

	EventCaseFileAddressAdded         = "casefile.address.added"
	EventCaseFileAddressUpdated       = "casefile.address.updated"
	EventCaseFileAddressRemoved       = "casefile.address.removed"
	EventCaseFileAddressMovedToDebtor = "casefile.address.moved_to_debtor"

	EventCaseFileEmailAdded         = "casefile.email.added"
	EventCaseFileEmailUpdated       = "casefile.email.updated"
	EventCaseFileEmailVerified      = "casefile.email.verified"
	EventCaseFileEmailRemoved       = "casefile.email.removed"
	EventCaseFileEmailMovedToDebtor = "casefile.email.moved_to_debtor"

	EventCaseFileAccountAdded   = "casefile.account.added"
	EventCaseFileAccountUpdated = "casefile.account.updated"

	EventCaseFileStatusUpdated                = "casefile.updated.status"
	EventCaseFileSubStatusUpdated             = "casefile.updated.sub_status"
	EventCaseFileCollectorAssigned            = "casefile.updated.collector"
	EventCaseFileToCollectorAssigned          = "casefile.updated.to_collector"
	EventCaseFileControllingCompanyUpdated    = "casefile.updated.controlling_company"
	EventCaseFileCollectionAttemptedOnUpdated = "casefile.updated.collection_attempted_on"

	EventCaseFilePaymentMethodAdded        = "casefile.payment.method.added"
	EventCaseFilePaymentPlanAdded          = "casefile.payment.plan.added"
	EventCaseFilePaymentPlanUpdated        = "casefile.payment.plan.updated"
	EventCaseFilePaymentPlanDeleted        = "casefile.payment.plan.deleted"
	EventCaseFilePaymentPlanESignSent      = "casefile.payment.plan.esign.sent"
	EventCaseFilePaymentPlanESignCompleted = "casefile.payment.plan.esign.completed"
	EventCaseFilePaymentPlanESignDeclined  = "casefile.payment.plan.esign.declined"
	EventCaseFilePaymentPlanESignExpired   = "casefile.payment.plan.esign.expired"
	EventCaseFilePaymentPlanESignError     = "casefile.payment.plan.esign.error"
	EventCaseFilePaymentPlanVoided         = "casefile.payment.plan.voided"
	EventCaseFilePaymentProcessed          = "casefile.payment.processed"
	EventCaseFilePaymentRefunded           = "casefile.payment.refunded"
	EventCaseFilePaymentCancelled          = "casefile.payment.cancelled"
	EventCaseFilePaymentCharged            = "casefile.payment.charged"
	EventCaseFilePaymentAttemptRecorded    = "casefile.payment.attempt.recorded"
	EventCaseFilePaymentLegacyUpdated      = "casefile.payment.legacy.updated"

	EventCaseFileNoteAdded   = "casefile.note.added"
	EventCaseFileEmailSent   = "casefile.email.sent"
	EventCaseFileEmailOpened = "casefile.email.opened"
	EventCaseFileEmailFailed = "casefile.email.failed"
	EventCaseFileMailSent    = "casefile.mail.sent"
	EventCaseFileVDrop       = "casefile.vdrop"

	EventDebtorCreated = "debtor.created"
	EventDebtorUpdated = "debtor.updated"
	EventDebtorDeleted = "debtor.deleted"

	EventPaymentReceived      = "payment.received"
	EventPaymentFailed        = "payment.failed"
	EventPaymentPlanCreated   = "paymentplan.created"
	EventPaymentPlanUpdated   = "paymentplan.updated"
	EventPaymentPlanCompleted = "paymentplan.completed"

	EventMailingCampaignCreated   = "mailingcampaign.created"
	EventMailingCampaignUpdated   = "mailingcampaign.updated"
	EventMailingCampaignDeleted   = "mailingcampaign.deleted"
	EventMailingCampaignSent      = "mailingcampaign.sent"
	EventMailingCampaignCompleted = "mailingcampaign.completed"
	EventCampaignSent             = "campaign.sent"
	EventCampaignCompleted        = "campaign.completed"

	EventEmailTemplateCreated = "emailtemplate.created"
	EventEmailTemplateUpdated = "emailtemplate.updated"
	EventEmailTemplateDeleted = "emailtemplate.deleted"

	EventTeamCreated      = "team.created"
	EventTeamUpdated      = "team.updated"
	EventTeamDeleted      = "team.deleted"
	EventTeamUserAssigned = "team.user.assigned"
	EventTeamUserRemoved  = "team.user.removed"

	EventScriptCreated = "script.created"
	EventScriptUpdated = "script.updated"
	EventScriptDeleted = "script.deleted"

	EventESignatureCompleted = "esignature.completed"
	EventESignatureDeclined  = "esignature.declined"
	EventCallReceived        = "call.received"
	EventSMSReceived         = "sms.received"
	EventWebhookReceived     = "webhook.received"
	EventVDropReceived       = "vdrop.received"
	EventEmailOpened         = "email.opened"

	EventAnyEvent            = "any.event"
	EventImportantEvent      = "important.event"
	EventCriticalEvent       = "critical.event"
	EventSystemEvent         = "system.event"
	EventBulkAction          = "bulk.action"
	EventError               = "error.event"
	EventUserMention         = "user.mention"
	EventTaskAssigned        = "task.assigned"
	EventDeadlineApproaching = "deadline.approaching"

	EventImportStarted   = "import.started"
	EventImportProgress  = "import.progress"
	EventImportCompleted = "import.completed"
	EventImportFailed    = "import.failed"

	EventJobStarted   = "job.started"
	EventJobProgress  = "job.progress"
	EventJobCompleted = "job.completed"
	EventJobFailed    = "job.failed"
	EventJobCancelled = "job.cancelled"
)

const (
	AggregateUser            = "User"
	AggregateOrganization    = "Organization"
	AggregateCompany         = "Company"
	AggregateCreditor        = "Creditor"
	AggregateMerchant        = "Merchant"
	AggregateCaseFile        = "CaseFile"
	AggregateMailingCampaign = "MailingCampaign"
	AggregateEmailTemplate   = "EmailTemplate"
	AggregateTeam            = "Team"
	AggregateScript          = "Script"
	AggregateJob             = "Job"
	AggregateImport          = "Import"
)
