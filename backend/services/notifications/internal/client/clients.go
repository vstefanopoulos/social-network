package client

import (
	chatPb "social-network/shared/gen-go/chat"
	postsPb "social-network/shared/gen-go/posts"
	usersPb "social-network/shared/gen-go/users"
)

// Clients holds connections to other services
type Clients struct {
	UsersClient usersPb.UserServiceClient
	ChatClient  chatPb.ChatServiceClient
	PostsClient postsPb.PostsServiceClient
}

// NewClients creates a new Clients instance
func NewClients(usersClient usersPb.UserServiceClient, chatClient chatPb.ChatServiceClient, postsClient postsPb.PostsServiceClient) *Clients {
	return &Clients{
		UsersClient: usersClient,
		ChatClient:  chatClient,
		PostsClient: postsClient,
	}
}
