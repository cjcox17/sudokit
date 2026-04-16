package testing

import (
	"context"
	"testing"
	"time"

	"github.com/cjcox17/sudokit/notifications"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMockNotificationsService_Create(t *testing.T) {
	mock := NewMockNotificationsService()
	ctx := context.Background()
	userID := primitive.NewObjectID()
	orgID := primitive.NewObjectID()

	input := notifications.CreateInput{
		Type:     notifications.TypeInfo,
		Category: notifications.CategorySystem,
		Title:    "Test",
		Message:  "Test message",
	}

	notif, err := mock.Create(ctx, userID, orgID, input)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if notif == nil {
		t.Fatal("expected notification, got nil")
	}

	if len(mock.CreateCalls) != 1 {
		t.Errorf("expected 1 create call, got %d", len(mock.CreateCalls))
	}

	call := mock.CreateCalls[0]
	if call.UserID != userID {
		t.Error("expected user ID to match")
	}

	if call.Input.Title != "Test" {
		t.Errorf("expected title 'Test', got '%s'", call.Input.Title)
	}
}

func TestMockNotificationsService_CreateForUsers(t *testing.T) {
	mock := NewMockNotificationsService()
	ctx := context.Background()
	userIDs := []primitive.ObjectID{primitive.NewObjectID(), primitive.NewObjectID()}
	orgID := primitive.NewObjectID()

	input := notifications.CreateInput{
		Type:     notifications.TypeSuccess,
		Category: notifications.CategoryImport,
		Title:    "Import Complete",
		Message:  "Your import has completed",
	}

	err := mock.CreateForUsers(ctx, userIDs, orgID, input)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(mock.CreateForUsersCalls) != 1 {
		t.Errorf("expected 1 CreateForUsers call, got %d", len(mock.CreateForUsersCalls))
	}

	call := mock.CreateForUsersCalls[0]
	if len(call.UserIDs) != 2 {
		t.Errorf("expected 2 user IDs, got %d", len(call.UserIDs))
	}
}

func TestMockNotificationsService_List(t *testing.T) {
	mock := NewMockNotificationsService()
	ctx := context.Background()
	userID := primitive.NewObjectID()

	opts := notifications.ListOptions{
		Limit:  10,
		Skip:   0,
		Unread: true,
	}

	_, _, err := mock.List(ctx, userID, opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(mock.ListCalls) != 1 {
		t.Errorf("expected 1 list call, got %d", len(mock.ListCalls))
	}

	call := mock.ListCalls[0]
	if call.UserID != userID {
		t.Error("expected user ID to match")
	}

	if call.Opts.Limit != 10 {
		t.Errorf("expected limit 10, got %d", call.Opts.Limit)
	}
}

func TestMockNotificationsService_GetUnreadCount(t *testing.T) {
	mock := NewMockNotificationsService()
	ctx := context.Background()
	userID := primitive.NewObjectID()

	count, err := mock.GetUnreadCount(ctx, userID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}
}

func TestMockNotificationsService_MarkAsRead(t *testing.T) {
	mock := NewMockNotificationsService()
	ctx := context.Background()
	notificationID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	err := mock.MarkAsRead(ctx, notificationID, userID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(mock.MarkAsReadCalls) != 1 {
		t.Errorf("expected 1 mark as read call, got %d", len(mock.MarkAsReadCalls))
	}

	call := mock.MarkAsReadCalls[0]
	if call.NotificationID != notificationID {
		t.Error("expected notification ID to match")
	}

	if call.UserID != userID {
		t.Error("expected user ID to match")
	}
}

func TestMockNotificationsService_MarkAllAsRead(t *testing.T) {
	mock := NewMockNotificationsService()
	ctx := context.Background()
	userID := primitive.NewObjectID()

	count, err := mock.MarkAllAsRead(ctx, userID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}
}

func TestMockNotificationsService_Delete(t *testing.T) {
	mock := NewMockNotificationsService()
	ctx := context.Background()
	notificationID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	err := mock.Delete(ctx, notificationID, userID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(mock.DeleteCalls) != 1 {
		t.Errorf("expected 1 delete call, got %d", len(mock.DeleteCalls))
	}

	call := mock.DeleteCalls[0]
	if call.NotificationID != notificationID {
		t.Error("expected notification ID to match")
	}
}

func TestMockNotificationsService_Reset(t *testing.T) {
	mock := NewMockNotificationsService()
	ctx := context.Background()
	userID := primitive.NewObjectID()
	orgID := primitive.NewObjectID()

	mock.Create(ctx, userID, orgID, notifications.CreateInput{})
	mock.List(ctx, userID, notifications.ListOptions{})
	mock.MarkAsRead(ctx, primitive.NewObjectID(), userID)

	mock.Reset()

	if len(mock.CreateCalls) != 0 {
		t.Error("expected create calls to be reset")
	}

	if len(mock.ListCalls) != 0 {
		t.Error("expected list calls to be reset")
	}

	if len(mock.MarkAsReadCalls) != 0 {
		t.Error("expected mark as read calls to be reset")
	}
}

func TestMockBroadcaster_Broadcast(t *testing.T) {
	mock := NewMockBroadcaster()
	userID := primitive.NewObjectID()

	mock.Broadcast(userID, "notification_new", map[string]any{"title": "Test"})

	if len(mock.BroadcastCalls) != 1 {
		t.Errorf("expected 1 broadcast call, got %d", len(mock.BroadcastCalls))
	}

	call := mock.BroadcastCalls[0]
	if call.UserID != userID {
		t.Error("expected user ID to match")
	}

	if call.Event != "notification_new" {
		t.Errorf("expected event 'notification_new', got '%s'", call.Event)
	}
}

func TestMockBroadcaster_GetCallsByEvent(t *testing.T) {
	mock := NewMockBroadcaster()
	userID := primitive.NewObjectID()

	mock.Broadcast(userID, "notification_new", nil)
	mock.Broadcast(userID, "notification_unread_count", nil)
	mock.Broadcast(userID, "notification_new", nil)

	calls := mock.GetCallsByEvent("notification_new")

	if len(calls) != 2 {
		t.Errorf("expected 2 calls for notification_new, got %d", len(calls))
	}
}

func TestMockBroadcaster_Reset(t *testing.T) {
	mock := NewMockBroadcaster()

	mock.Broadcast(primitive.NewObjectID(), "event", nil)
	mock.Reset()

	if len(mock.BroadcastCalls) != 0 {
		t.Error("expected broadcast calls to be reset")
	}
}

func TestAssertNotificationCreated(t *testing.T) {
	mock := NewMockNotificationsService()
	ctx := context.Background()
	userID := primitive.NewObjectID()
	orgID := primitive.NewObjectID()

	mock.Create(ctx, userID, orgID, notifications.CreateInput{})

	AssertNotificationCreated(t, mock, userID)
}

func TestAssertNotificationMarkedAsRead(t *testing.T) {
	mock := NewMockNotificationsService()
	ctx := context.Background()
	notificationID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	mock.MarkAsRead(ctx, notificationID, userID)

	AssertNotificationMarkedAsRead(t, mock, notificationID)
}

func TestCreateTestNotification(t *testing.T) {
	userID := primitive.NewObjectID()
	notif := CreateTestNotification(userID, notifications.TypeInfo, notifications.CategorySystem)

	if notif.UserID != userID {
		t.Error("expected user ID to match")
	}

	if notif.Type != notifications.TypeInfo {
		t.Errorf("expected type '%s', got '%s'", notifications.TypeInfo, notif.Type)
	}

	if notif.Category != notifications.CategorySystem {
		t.Errorf("expected category '%s', got '%s'", notifications.CategorySystem, notif.Category)
	}

	if notif.Title != "Test Notification" {
		t.Errorf("expected title 'Test Notification', got '%s'", notif.Title)
	}
}

func TestNotificationTypes(t *testing.T) {
	tests := []struct {
		name     string
		nt       notifications.Type
		expected string
	}{
		{"info", notifications.TypeInfo, "info"},
		{"warning", notifications.TypeWarning, "warning"},
		{"success", notifications.TypeSuccess, "success"},
		{"error", notifications.TypeError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.nt) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.nt)
			}

			if !tt.nt.IsValid() {
				t.Errorf("expected type %s to be valid", tt.nt)
			}
		})
	}
}

func TestNotificationCategories(t *testing.T) {
	tests := []struct {
		name     string
		nc       notifications.Category
		expected string
	}{
		{"system", notifications.CategorySystem, "system"},
		{"task", notifications.CategoryTask, "task"},
		{"mention", notifications.CategoryMention, "mention"},
		{"deadline", notifications.CategoryDeadline, "deadline"},
		{"esign", notifications.CategoryESign, "esign"},
		{"import", notifications.CategoryImport, "import"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.nc) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.nc)
			}

			if !tt.nc.IsValid() {
				t.Errorf("expected category %s to be valid", tt.nc)
			}
		})
	}
}

func TestTypeIsValid(t *testing.T) {
	validTypes := []notifications.Type{
		notifications.TypeInfo,
		notifications.TypeWarning,
		notifications.TypeSuccess,
		notifications.TypeError,
	}

	for _, nt := range validTypes {
		if !nt.IsValid() {
			t.Errorf("expected type '%s' to be valid", nt)
		}
	}

	invalidType := notifications.Type("invalid")
	if invalidType.IsValid() {
		t.Error("expected invalid type to be invalid")
	}
}

func TestCategoryIsValid(t *testing.T) {
	validCategories := []notifications.Category{
		notifications.CategorySystem,
		notifications.CategoryTask,
		notifications.CategoryMention,
		notifications.CategoryDeadline,
		notifications.CategoryESign,
		notifications.CategoryImport,
	}

	for _, nc := range validCategories {
		if !nc.IsValid() {
			t.Errorf("expected category '%s' to be valid", nc)
		}
	}

	invalidCategory := notifications.Category("invalid")
	if invalidCategory.IsValid() {
		t.Error("expected invalid category to be invalid")
	}
}

func TestCreateInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   notifications.CreateInput
		wantErr error
	}{
		{
			name: "valid input",
			input: notifications.CreateInput{
				Type:     notifications.TypeInfo,
				Category: notifications.CategorySystem,
				Title:    "Test",
				Message:  "Test message",
			},
			wantErr: nil,
		},
		{
			name: "invalid type",
			input: notifications.CreateInput{
				Type:     notifications.Type("invalid"),
				Category: notifications.CategorySystem,
				Title:    "Test",
				Message:  "Test message",
			},
			wantErr: notifications.ErrInvalidType,
		},
		{
			name: "invalid category",
			input: notifications.CreateInput{
				Type:     notifications.TypeInfo,
				Category: notifications.Category("invalid"),
				Title:    "Test",
				Message:  "Test message",
			},
			wantErr: notifications.ErrInvalidCategory,
		},
		{
			name: "empty title",
			input: notifications.CreateInput{
				Type:     notifications.TypeInfo,
				Category: notifications.CategorySystem,
				Title:    "",
				Message:  "Test message",
			},
			wantErr: notifications.ErrTitleEmpty,
		},
		{
			name: "empty message",
			input: notifications.CreateInput{
				Type:     notifications.TypeInfo,
				Category: notifications.CategorySystem,
				Title:    "Test",
				Message:  "",
			},
			wantErr: notifications.ErrMessageEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()

			if tt.wantErr == nil && err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if tt.wantErr != nil && err != tt.wantErr {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestNotificationErrors(t *testing.T) {
	tests := []struct {
		name string
		err  *notifications.Error
		code string
		msg  string
	}{
		{"invalid type", notifications.ErrInvalidType, "INVALID_TYPE", "invalid notification type"},
		{"invalid category", notifications.ErrInvalidCategory, "INVALID_CATEGORY", "invalid notification category"},
		{"title empty", notifications.ErrTitleEmpty, "TITLE_EMPTY", "title cannot be empty"},
		{"message empty", notifications.ErrMessageEmpty, "MESSAGE_EMPTY", "message cannot be empty"},
		{"not found", notifications.ErrNotFound, "NOT_FOUND", "notification not found"},
		{"not owner", notifications.ErrNotOwner, "NOT_OWNER", "you do not own this notification"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.code {
				t.Errorf("expected code '%s', got '%s'", tt.code, tt.err.Code)
			}

			if tt.err.Message != tt.msg {
				t.Errorf("expected message '%s', got '%s'", tt.msg, tt.err.Message)
			}

			if tt.err.Error() != tt.msg {
				t.Errorf("expected Error() to return '%s', got '%s'", tt.msg, tt.err.Error())
			}
		})
	}
}

func TestMockNotificationsService_GetLastCreateCall(t *testing.T) {
	mock := NewMockNotificationsService()
	ctx := context.Background()
	userID1 := primitive.NewObjectID()
	userID2 := primitive.NewObjectID()
	orgID := primitive.NewObjectID()

	mock.Create(ctx, userID1, orgID, notifications.CreateInput{Title: "First"})
	mock.Create(ctx, userID2, orgID, notifications.CreateInput{Title: "Second"})

	call := mock.GetLastCreateCall()
	if call == nil {
		t.Fatal("expected call, got nil")
	}

	if call.UserID != userID2 {
		t.Error("expected last user ID to match")
	}

	if call.Input.Title != "Second" {
		t.Errorf("expected title 'Second', got '%s'", call.Input.Title)
	}
}

func TestMockNotificationsService_CustomBehavior(t *testing.T) {
	mock := NewMockNotificationsService()

	expectedID := primitive.NewObjectID()
	mock.CreateFunc = func(ctx context.Context, userID, orgID primitive.ObjectID, input notifications.CreateInput) (*notifications.Notification, error) {
		return &notifications.Notification{ID: expectedID}, nil
	}

	ctx := context.Background()
	notif, _ := mock.Create(ctx, primitive.NewObjectID(), primitive.NewObjectID(), notifications.CreateInput{})

	if notif.ID != expectedID {
		t.Errorf("expected ID %s, got %s", expectedID.Hex(), notif.ID.Hex())
	}
}

func TestListOptions(t *testing.T) {
	opts := notifications.ListOptions{
		Limit:  50,
		Skip:   10,
		Unread: true,
	}

	if opts.Limit != 50 {
		t.Errorf("expected limit 50, got %d", opts.Limit)
	}

	if opts.Skip != 10 {
		t.Errorf("expected skip 10, got %d", opts.Skip)
	}

	if !opts.Unread {
		t.Error("expected unread to be true")
	}
}

func TestNotificationStruct(t *testing.T) {
	now := time.Now()
	actionURL := "/test/path"

	notif := &notifications.Notification{
		ID:             primitive.NewObjectID(),
		OrganizationID: primitive.NewObjectID(),
		UserID:         primitive.NewObjectID(),
		Type:           notifications.TypeSuccess,
		Category:       notifications.CategoryImport,
		Title:          "Import Complete",
		Message:        "Your import has completed successfully",
		ActionURL:      &actionURL,
		Metadata:       map[string]any{"job_id": "123"},
		IsRead:         false,
		CreatedAt:      now,
	}

	if notif.Type != notifications.TypeSuccess {
		t.Errorf("expected type '%s', got '%s'", notifications.TypeSuccess, notif.Type)
	}

	if notif.Category != notifications.CategoryImport {
		t.Errorf("expected category '%s', got '%s'", notifications.CategoryImport, notif.Category)
	}

	if notif.Title != "Import Complete" {
		t.Errorf("expected title 'Import Complete', got '%s'", notif.Title)
	}

	if *notif.ActionURL != "/test/path" {
		t.Errorf("expected action URL '/test/path', got '%s'", *notif.ActionURL)
	}

	if notif.Metadata["job_id"] != "123" {
		t.Error("expected metadata job_id to be '123'")
	}

	if notif.IsRead {
		t.Error("expected IsRead to be false")
	}

	if notif.CreatedAt != now {
		t.Error("expected CreatedAt to match")
	}
}
