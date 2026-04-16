package notifications

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Notification struct {
	ID             primitive.ObjectID `bson:"_id" json:"id"`
	OrganizationID primitive.ObjectID `bson:"organization_id" json:"organization_id"`
	UserID         primitive.ObjectID `bson:"user_id" json:"user_id"`
	Type           Type               `bson:"type" json:"type"`
	Category       Category           `bson:"category" json:"category"`
	Title          string             `bson:"title" json:"title"`
	Message        string             `bson:"message" json:"message"`
	ActionURL      *string            `bson:"action_url,omitempty" json:"action_url,omitempty"`
	Metadata       map[string]any     `bson:"metadata,omitempty" json:"metadata,omitempty"`
	IsRead         bool               `bson:"is_read" json:"is_read"`
	ReadAt         *time.Time         `bson:"read_at,omitempty" json:"read_at,omitempty"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
}

type Type string

const (
	TypeInfo    Type = "info"
	TypeWarning Type = "warning"
	TypeSuccess Type = "success"
	TypeError   Type = "error"
)

func (t Type) IsValid() bool {
	switch t {
	case TypeInfo, TypeWarning, TypeSuccess, TypeError:
		return true
	default:
		return false
	}
}

type Category string

const (
	CategorySystem   Category = "system"
	CategoryTask     Category = "task"
	CategoryMention  Category = "mention"
	CategoryDeadline Category = "deadline"
	CategoryESign    Category = "esign"
	CategoryImport   Category = "import"
	CategoryExport   Category = "export"
	CategoryShare    Category = "share"
)

func (c Category) IsValid() bool {
	switch c {
	case CategorySystem, CategoryTask, CategoryMention, CategoryDeadline, CategoryESign, CategoryImport, CategoryExport, CategoryShare:
		return true
	default:
		return false
	}
}

type CreateInput struct {
	Type      Type
	Category  Category
	Title     string
	Message   string
	ActionURL *string
	Metadata  map[string]any
}

func (i CreateInput) Validate() error {
	if !i.Type.IsValid() {
		return ErrInvalidType
	}
	if !i.Category.IsValid() {
		return ErrInvalidCategory
	}
	if i.Title == "" {
		return ErrTitleEmpty
	}
	if i.Message == "" {
		return ErrMessageEmpty
	}
	return nil
}

type ListOptions struct {
	Limit  int
	Skip   int
	Unread bool
}

type BroadcastFunc func(userID primitive.ObjectID, event string, data any)

var (
	ErrInvalidType     = &Error{Code: "INVALID_TYPE", Message: "invalid notification type"}
	ErrInvalidCategory = &Error{Code: "INVALID_CATEGORY", Message: "invalid notification category"}
	ErrTitleEmpty      = &Error{Code: "TITLE_EMPTY", Message: "title cannot be empty"}
	ErrMessageEmpty    = &Error{Code: "MESSAGE_EMPTY", Message: "message cannot be empty"}
	ErrNotFound        = &Error{Code: "NOT_FOUND", Message: "notification not found"}
	ErrNotOwner        = &Error{Code: "NOT_OWNER", Message: "you do not own this notification"}
)

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

type ServiceInterface interface {
	EnsureIndexes(ctx context.Context) error
	Create(ctx context.Context, userID, orgID primitive.ObjectID, input CreateInput) (*Notification, error)
	CreateForUsers(ctx context.Context, userIDs []primitive.ObjectID, orgID primitive.ObjectID, input CreateInput) error
	CreateForOrg(ctx context.Context, orgID primitive.ObjectID, input CreateInput) error
	List(ctx context.Context, userID primitive.ObjectID, opts ListOptions) ([]*Notification, int, error)
	GetUnreadCount(ctx context.Context, userID primitive.ObjectID) (int, error)
	MarkAsRead(ctx context.Context, notificationID, userID primitive.ObjectID) error
	MarkAllAsRead(ctx context.Context, userID primitive.ObjectID) (int, error)
	Delete(ctx context.Context, notificationID, userID primitive.ObjectID) error
}
