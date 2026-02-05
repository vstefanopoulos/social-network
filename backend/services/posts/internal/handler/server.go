package handler

import (
	"social-network/services/posts/internal/application"
	pb "social-network/shared/gen-go/posts"
)

// Holds Client conns, services and handler funcs
type PostsHandler struct {
	pb.UnimplementedPostsServiceServer
	Application *application.Application
}

func NewPostsHandler(service *application.Application) *PostsHandler {
	return &PostsHandler{
		Application: service,
	}
}
