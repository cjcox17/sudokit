package payloads

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserCreatedPayload struct {
	UserID         primitive.ObjectID `json:"user_id"`
	Email          string             `json:"email"`
	Name           string             `json:"name"`
	Role           string             `json:"role"`
	OrganizationID primitive.ObjectID `json:"organization_id"`
}

type UserUpdatedPayload struct {
	UserID  primitive.ObjectID     `json:"user_id"`
	Email   string                 `json:"email"`
	Name    string                 `json:"name"`
	Changes map[string]FieldChange `json:"changes"`
}

type UserLoggedInPayload struct {
	UserID    primitive.ObjectID `json:"user_id"`
	Email     string             `json:"email"`
	IPAddress string             `json:"ip_address"`
	UserAgent string             `json:"user_agent"`
	LoginAt   time.Time          `json:"login_at"`
}

type UserLoggedOutPayload struct {
	UserID   primitive.ObjectID `json:"user_id"`
	Email    string             `json:"email"`
	LogoutAt time.Time          `json:"logout_at"`
}

type UserWentAwayPayload struct {
	UserID primitive.ObjectID `json:"user_id"`
	AwayAt time.Time          `json:"away_at"`
}

type UserCameBackPayload struct {
	UserID primitive.ObjectID `json:"user_id"`
	BackAt time.Time          `json:"back_at"`
}

type OrganizationCreatedPayload struct {
	OrganizationID primitive.ObjectID `json:"organization_id"`
	Name           string             `json:"name"`
	OwnerID        primitive.ObjectID `json:"owner_id"`
	OwnerEmail     string             `json:"owner_email"`
}

type OrganizationUpdatedPayload struct {
	OrganizationID primitive.ObjectID     `json:"organization_id"`
	Name           string                 `json:"name"`
	Changes        map[string]FieldChange `json:"changes"`
}

type UserMentionPayload struct {
	MentionedUserID     primitive.ObjectID `json:"mentioned_user_id"`
	MentionedByUserID   primitive.ObjectID `json:"mentioned_by_user_id"`
	MentionedByUserName string             `json:"mentioned_by_user_name"`
	Context             string             `json:"context"`
	EntityID            primitive.ObjectID `json:"entity_id"`
	EntityType          string             `json:"entity_type"`
}

type TaskAssignedPayload struct {
	TaskID         primitive.ObjectID `json:"task_id"`
	TaskTitle      string             `json:"task_title"`
	AssigneeID     primitive.ObjectID `json:"assignee_id"`
	AssigneeName   string             `json:"assignee_name"`
	AssignedByID   primitive.ObjectID `json:"assigned_by_id"`
	AssignedByName string             `json:"assigned_by_name"`
	DueDate        *time.Time         `json:"due_date,omitempty"`
}

type DeadlineApproachingPayload struct {
	EntityID       primitive.ObjectID `json:"entity_id"`
	EntityType     string             `json:"entity_type"`
	EntityTitle    string             `json:"entity_title"`
	Deadline       time.Time          `json:"deadline"`
	HoursRemaining int                `json:"hours_remaining"`
	UserID         primitive.ObjectID `json:"user_id"`
	UserName       string             `json:"user_name"`
}
