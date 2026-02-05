package handler

import (
	"social-network/services/users/internal/application"
	"social-network/shared/gen-go/users"
)

// Holds Client conns, services and handler funcs
type UsersHandler struct {
	users.UnimplementedUserServiceServer
	Application *application.Application
}

func NewUsersHanlder(service *application.Application) *UsersHandler {
	return &UsersHandler{
		Application: service,
	}
}
