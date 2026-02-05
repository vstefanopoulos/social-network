package handler

import (
	"social-network/services/notifications/internal/application"
	pb "social-network/shared/gen-go/notifications"
)

// Holds Client conns, services and handler funcs
type Server struct {
	pb.UnimplementedNotificationServiceServer
	Application *application.Application
}

func NewNotificationsHandler(app *application.Application) *Server {
	return &Server{
		Application: app,
	}
}
