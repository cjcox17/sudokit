package notifications

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service struct {
	db         *mongo.Database
	collection *mongo.Collection
	broadcast  BroadcastFunc
	mu         sync.RWMutex
}

func NewService(db *mongo.Database, broadcast BroadcastFunc) *Service {
	return &Service{
		db:         db,
		collection: db.Collection("notifications"),
		broadcast:  broadcast,
	}
}

func (s *Service) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "is_read", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "organization_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
	}

	_, err := s.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	slog.Info("Notifications indexes created")
	return nil
}

func (s *Service) Create(ctx context.Context, userID, orgID primitive.ObjectID, input CreateInput) (*Notification, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	notification := &Notification{
		ID:             primitive.NewObjectID(),
		OrganizationID: orgID,
		UserID:         userID,
		Type:           input.Type,
		Category:       input.Category,
		Title:          input.Title,
		Message:        input.Message,
		ActionURL:      input.ActionURL,
		Metadata:       input.Metadata,
		IsRead:         false,
		CreatedAt:      now,
	}

	_, err := s.collection.InsertOne(ctx, notification)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	go s.broadcastNotification(notification)
	go s.broadcastUnreadCountUpdate(ctx, userID)

	slog.Debug("Notification created", "id", notification.ID, "user_id", userID)
	return notification, nil
}

func (s *Service) CreateForUsers(ctx context.Context, userIDs []primitive.ObjectID, orgID primitive.ObjectID, input CreateInput) error {
	if err := input.Validate(); err != nil {
		return err
	}

	now := time.Now()
	var notifications []interface{}

	for _, userID := range userIDs {
		notification := &Notification{
			ID:             primitive.NewObjectID(),
			OrganizationID: orgID,
			UserID:         userID,
			Type:           input.Type,
			Category:       input.Category,
			Title:          input.Title,
			Message:        input.Message,
			ActionURL:      input.ActionURL,
			Metadata:       input.Metadata,
			IsRead:         false,
			CreatedAt:      now,
		}
		notifications = append(notifications, notification)
	}

	_, err := s.collection.InsertMany(ctx, notifications)
	if err != nil {
		return fmt.Errorf("failed to create notifications: %w", err)
	}

	go func() {
		for _, userID := range userIDs {
			s.broadcastUnreadCountUpdate(ctx, userID)
		}
	}()

	slog.Debug("Notifications created for users", "count", len(userIDs))
	return nil
}

func (s *Service) CreateForOrg(ctx context.Context, orgID primitive.ObjectID, input CreateInput) error {
	return fmt.Errorf("CreateForOrg requires user list - use CreateForUsers instead")
}

func (s *Service) List(ctx context.Context, userID primitive.ObjectID, opts ListOptions) ([]*Notification, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if opts.Limit <= 0 {
		opts.Limit = 20
	}
	if opts.Limit > 100 {
		opts.Limit = 100
	}

	filter := bson.M{"user_id": userID}
	if opts.Unread {
		filter["is_read"] = false
	}

	findOpts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(opts.Limit)).
		SetSkip(int64(opts.Skip))

	cursor, err := s.collection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find notifications: %w", err)
	}
	defer cursor.Close(ctx)

	var notifications []*Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, 0, fmt.Errorf("failed to decode notifications: %w", err)
	}

	if notifications == nil {
		notifications = []*Notification{}
	}

	total, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	return notifications, int(total), nil
}

func (s *Service) GetUnreadCount(ctx context.Context, userID primitive.ObjectID) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filter := bson.M{
		"user_id": userID,
		"is_read": false,
	}

	count, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	return int(count), nil
}

func (s *Service) MarkAsRead(ctx context.Context, notificationID, userID primitive.ObjectID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var notification Notification
	err := s.collection.FindOne(ctx, bson.M{"_id": notificationID}).Decode(&notification)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrNotFound
		}
		return fmt.Errorf("failed to find notification: %w", err)
	}

	if notification.UserID != userID {
		return ErrNotOwner
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"is_read": true,
			"read_at": now,
		},
	}

	_, err = s.collection.UpdateByID(ctx, notificationID, update)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	go s.broadcastUnreadCountUpdate(ctx, userID)

	slog.Debug("Notification marked as read", "id", notificationID)
	return nil
}

func (s *Service) MarkAllAsRead(ctx context.Context, userID primitive.ObjectID) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	filter := bson.M{
		"user_id": userID,
		"is_read": false,
	}
	update := bson.M{
		"$set": bson.M{
			"is_read": true,
			"read_at": now,
		},
	}

	result, err := s.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, fmt.Errorf("failed to mark all as read: %w", err)
	}

	go s.broadcastUnreadCountUpdate(ctx, userID)

	slog.Debug("All notifications marked as read", "user_id", userID, "count", result.ModifiedCount)
	return int(result.ModifiedCount), nil
}

func (s *Service) Delete(ctx context.Context, notificationID, userID primitive.ObjectID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var notification Notification
	err := s.collection.FindOne(ctx, bson.M{"_id": notificationID}).Decode(&notification)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrNotFound
		}
		return fmt.Errorf("failed to find notification: %w", err)
	}

	if notification.UserID != userID {
		return ErrNotOwner
	}

	_, err = s.collection.DeleteOne(ctx, bson.M{"_id": notificationID})
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	if !notification.IsRead {
		go s.broadcastUnreadCountUpdate(ctx, userID)
	}

	slog.Debug("Notification deleted", "id", notificationID)
	return nil
}

func (s *Service) broadcastNotification(notification *Notification) {
	if s.broadcast == nil {
		return
	}

	s.broadcast(notification.UserID, "notification_new", map[string]any{
		"id":         notification.ID.Hex(),
		"type":       notification.Type,
		"category":   notification.Category,
		"title":      notification.Title,
		"message":    notification.Message,
		"action_url": notification.ActionURL,
		"metadata":   notification.Metadata,
		"is_read":    notification.IsRead,
		"created_at": notification.CreatedAt,
	})
}

func (s *Service) broadcastUnreadCountUpdate(ctx context.Context, userID primitive.ObjectID) {
	if s.broadcast == nil {
		return
	}

	count, err := s.GetUnreadCount(ctx, userID)
	if err != nil {
		slog.Error("Failed to get unread count", "error", err)
		return
	}

	s.broadcast(userID, "notification_unread_count", map[string]any{
		"unread_count": count,
	})
}

var (
	globalService *Service
	serviceOnce   sync.Once
)

func InitService(db *mongo.Database, broadcast BroadcastFunc) *Service {
	serviceOnce.Do(func() {
		globalService = NewService(db, broadcast)
	})
	return globalService
}

func GetService() *Service {
	return globalService
}
