package handler

import (
	"social-network/services/chat/internal/application"
	pb "social-network/shared/gen-go/chat"
)

type ChatHandler struct {
	pb.UnimplementedChatServiceServer
	Application *application.ChatService
}

func NewChatHandler(app *application.ChatService) *ChatHandler {
	return &ChatHandler{
		Application: app,
	}
}
