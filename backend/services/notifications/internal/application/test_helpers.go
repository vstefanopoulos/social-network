package application

import (
	"context"

	"social-network/services/notifications/internal/db/sqlc"

	"github.com/stretchr/testify/mock"
)

// MockDB is a mock implementation of the database queries
type MockDB struct {
	mock.Mock
}

func (m *MockDB) CreateNotification(ctx context.Context, arg sqlc.CreateNotificationParams) (sqlc.Notification, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(sqlc.Notification), args.Error(1)
}

func (m *MockDB) GetNotificationByID(ctx context.Context, id int64) (sqlc.Notification, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(sqlc.Notification), args.Error(1)
}

func (m *MockDB) GetUserNotifications(ctx context.Context, arg sqlc.GetUserNotificationsParams) ([]sqlc.Notification, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]sqlc.Notification), args.Error(1)
}

func (m *MockDB) GetUserNotificationsCount(ctx context.Context, userID int64) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDB) GetUserUnreadNotificationsCount(ctx context.Context, userID int64) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDB) MarkNotificationAsRead(ctx context.Context, arg sqlc.MarkNotificationAsReadParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockDB) MarkAllAsRead(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockDB) DeleteNotification(ctx context.Context, arg sqlc.DeleteNotificationParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockDB) CreateNotificationType(ctx context.Context, arg sqlc.CreateNotificationTypeParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockDB) GetNotificationType(ctx context.Context, notifType string) (sqlc.NotificationType, error) {
	args := m.Called(ctx, notifType)
	return args.Get(0).(sqlc.NotificationType), args.Error(1)
}

func (m *MockDB) UpdateNotificationCount(ctx context.Context, arg sqlc.UpdateNotificationCountParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockDB) GetUnreadNotificationByTypeAndEntity(ctx context.Context, arg sqlc.GetUnreadNotificationByTypeAndEntityParams) (sqlc.Notification, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(sqlc.Notification), args.Error(1)
}

func (m *MockDB) GetNotificationByTypeAndEntity(ctx context.Context, arg sqlc.GetNotificationByTypeAndEntityParams) (sqlc.Notification, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(sqlc.Notification), args.Error(1)
}

func (m *MockDB) MarkNotificationAsActed(ctx context.Context, arg sqlc.MarkNotificationAsActedParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

// NewApplicationWithMocks creates a new application service with mocked dependencies for testing
func NewApplicationWithMocks(db DBInterface) *Application {
	return &Application{
		DB:      db,
		Clients: nil, // nil for tests
		NatsConn: nil, // nil for tests
	}
}