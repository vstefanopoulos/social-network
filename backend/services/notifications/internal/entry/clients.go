/*
Establish connections to other services
*/

package entry

// import (
// 	"fmt"
// 	"social-network/services/notifications/internal/client"
// 	chatpb "social-network/shared/gen-go/chat"
// 	postspb "social-network/shared/gen-go/posts"
// 	userpb "social-network/shared/gen-go/users"
// 	"social-network/shared/ports"
// 	"time"

// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/backoff"
// 	"google.golang.org/grpc/credentials/insecure"
// )

// func InitClients() *client.Clients {
// 	c := &client.Clients{}
// 	dialOpts := []grpc.DialOption{
// 		grpc.WithTransportCredentials(insecure.NewCredentials()),
// 		grpc.WithDefaultServiceConfig(`{
//         "loadBalancingConfig": [{"round_robin":{}}]
//     	}`),
// 		grpc.WithConnectParams(grpc.ConnectParams{
// 			MinConnectTimeout: 2 * time.Second,
// 			Backoff: backoff.Config{
// 				BaseDelay:  1 * time.Second,
// 				Multiplier: 1.2,
// 				Jitter:     0.5,
// 				MaxDelay:   5 * time.Second,
// 			},
// 		}),
// 	}

// 	// List of initializer functions
// 	initializers := []func(opts []grpc.DialOption, c *client.Clients) error{
// 		InitUserClient,
// 		InitChatClient,
// 		InitPostsClient,
// 		// Add more here as you add more clients
// 	}

// 	for _, initFn := range initializers {
// 		if err := initFn(dialOpts, c); err != nil {
// 			fmt.Println(err)
// 		}
// 	}
// 	return c
// }

// func InitUserClient(opts []grpc.DialOption, c *client.Clients) (err error) {
// 	conn, err := grpc.NewClient(ports.Users, opts...)
// 	if err != nil {
// 		err = fmt.Errorf("failed to dial user service: %v", err)
// 	}
// 	c.UsersClient = userpb.NewUserServiceClient(conn)
// 	return err
// }

// func InitChatClient(opts []grpc.DialOption, c *client.Clients) (err error) {
// 	conn, err := grpc.NewClient(ports.Chat, opts...)
// 	if err != nil {
// 		err = fmt.Errorf("failed to dial chat service: %v", err)
// 	}
// 	c.ChatClient = chatpb.NewChatServiceClient(conn)
// 	return err
// }

// func InitPostsClient(opts []grpc.DialOption, c *client.Clients) (err error) {
// 	conn, err := grpc.NewClient(ports.Posts, opts...)
// 	if err != nil {
// 		err = fmt.Errorf("failed to dial posts service: %v", err)
// 	}
// 	c.PostsClient = postspb.NewPostsServiceClient(conn)
// 	return err
// }
