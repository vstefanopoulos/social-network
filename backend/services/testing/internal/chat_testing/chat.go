package chattesting

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"slices"
	"time"

	"social-network/services/testing/internal/configs"
	"social-network/services/testing/internal/utils"
	"social-network/shared/gen-go/chat"
	"social-network/shared/gen-go/common"
	"social-network/shared/gen-go/users"
	ce "social-network/shared/go/commonerrors"
	"social-network/shared/go/ct"
	"social-network/shared/go/gorpc"
	"social-network/shared/go/mapping"

	"google.golang.org/protobuf/types/known/timestamppb"
)

var UsersService users.UserServiceClient
var ChatService chat.ChatServiceClient

var fail = "FAIL TEST: err ->"
var usrA *users.RegisterUserResponse
var usrB *users.RegisterUserResponse
var usrC *users.RegisterUserResponse

func StartTest(ctx context.Context, cfgs configs.Configs) error {
	var err error
	UsersService, err = gorpc.GetGRpcClient(
		users.NewUserServiceClient,
		cfgs.UsersGRPCAddr,
		ct.CommonKeys(),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to users service: %s", err.Error())
	}
	ChatService, err = gorpc.GetGRpcClient(
		chat.NewChatServiceClient,
		cfgs.ChatGRPCAddr,
		ct.CommonKeys(),
	)
	// Set Up Users.
	// Start with a clean users db or make sure the users.json exists and has valid user A, B and C ids
	// 	[
	//   {
	//     "user_id": 1,
	//     "username": "testUserA"
	//   },
	//   {
	//     "user_id": 2,
	//     "username": "testUserB"
	//   },
	//   {
	//     "user_id": 3,
	//     "username": "testUserC"
	//   }
	// ]
	utils.HandleErr("register users", ctx, registerOrGetUsers)
	utils.HandleErr("follow each other", ctx, FollowUser)

	// Group Convsersations
	utils.HandleErr("test create group conv", ctx, TestGroupConversation)
	utils.HandleErr("test get group messages", ctx, TestGetGroupMessages)

	// Private Conversations
	utils.HandleErr("test get convs count with unread msgs", ctx, TestGetConversationsCountWithUnreadMsgs)
	utils.HandleErr("test unread conversations", ctx, TestUnreadCount)
	utils.HandleErr("send msg to each other", ctx, TestCreateMessage)
	utils.HandleErr("get conversations before", ctx, TestGetConversationsBefore)
	utils.HandleErr("get conversations", ctx, TestGetConversations)
	utils.HandleErr("get previous private messages", ctx, TestGetPMs)
	utils.HandleErr("get next private messages", ctx, TestGetNextPms)

	return nil
}

// Registers two users and saves them to users.json. If users already exist it retrieves the user Ids from file
func registerOrGetUsers(ctx context.Context) error {
	fmt.Println("users-service starting register test")
	var err error
	usrA, err = UsersService.RegisterUser(ctx, &users.RegisterUserRequest{
		Username:    "testUserA",
		FirstName:   "john",
		LastName:    "doe",
		DateOfBirth: timestamppb.New(time.Unix(rand.Int64N(1000000), 0)),
		Avatar:      0,
		About:       utils.RandomString(300, true),
		Public:      false,
		Email:       "usera@hotmail.com",
		Password:    "Hello12!",
	},
	)
	decoded := ce.DecodeProto(err)
	if decoded != nil && decoded.IsClass(ce.ErrAlreadyExists) {
		users, _ := ReadUserFromJSON()
		if len(users) > 2 {
			usrA = &users[0]
			usrB = &users[1]
			usrC = &users[2]
		} else {
			return err
		}
		return nil
	}
	if err != nil {
		return errors.New(fail + err.Error())
	}

	AppendStructToUsersJSON(usrA)
	usrB, err = UsersService.RegisterUser(ctx, &users.RegisterUserRequest{
		Username:    "testUserB",
		FirstName:   "jack",
		LastName:    "foe",
		DateOfBirth: timestamppb.New(time.Unix(rand.Int64N(1000000), 0)),
		Avatar:      0,
		About:       utils.RandomString(300, true),
		Public:      false,
		Email:       "userb@hotmail.com",
		Password:    "Hello12!",
	},
	)
	if err != nil {
		return errors.New(fail + err.Error())
	}

	AppendStructToUsersJSON(usrB)
	usrC, err = UsersService.RegisterUser(ctx, &users.RegisterUserRequest{
		Username:    "testUserC",
		FirstName:   "jack",
		LastName:    "foe",
		DateOfBirth: timestamppb.New(time.Unix(rand.Int64N(1000000), 0)),
		Avatar:      0,
		About:       utils.RandomString(300, true),
		Public:      false,
		Email:       "userc@hotmail.com",
		Password:    "Hello12!",
	},
	)
	if err != nil {
		return errors.New(fail + err.Error())
	}
	AppendStructToUsersJSON(usrC)

	return nil
}

func FollowUser(ctx context.Context) error {

	// User A to B
	_, err := UsersService.FollowUser(ctx, &users.FollowUserRequest{
		FollowerId:   usrA.UserId,
		TargetUserId: usrB.UserId,
	})
	if err != nil {
		return err
	}
	// B Accepts
	_, err = UsersService.HandleFollowRequest(ctx, &users.HandleFollowRequestRequest{
		UserId:      usrB.UserId,
		RequesterId: usrA.UserId,
		Accept:      true,
	})
	if err != nil {
		return err
	}

	// User B to A
	_, err = UsersService.FollowUser(ctx, &users.FollowUserRequest{
		FollowerId:   usrB.UserId,
		TargetUserId: usrA.UserId,
	})
	if err != nil {
		return err
	}
	// A accepts
	_, err = UsersService.HandleFollowRequest(ctx, &users.HandleFollowRequestRequest{
		UserId:      usrA.UserId,
		RequesterId: usrB.UserId,
		Accept:      true,
	})

	// User A to C
	_, err = UsersService.FollowUser(ctx, &users.FollowUserRequest{
		FollowerId:   usrA.UserId,
		TargetUserId: usrC.UserId,
	})
	if err != nil {
		return err
	}
	// C Accepts
	_, err = UsersService.HandleFollowRequest(ctx, &users.HandleFollowRequestRequest{
		UserId:      usrC.UserId,
		RequesterId: usrA.UserId,
		Accept:      true,
	})
	if err != nil {
		return err
	}

	// User C to A
	_, err = UsersService.FollowUser(ctx, &users.FollowUserRequest{
		FollowerId:   usrC.UserId,
		TargetUserId: usrA.UserId,
	})
	if err != nil {
		return err
	}
	// A accepts
	_, err = UsersService.HandleFollowRequest(ctx, &users.HandleFollowRequestRequest{
		UserId:      usrA.UserId,
		RequesterId: usrC.UserId,
		Accept:      true,
	})
	if err != nil {
		return err
	}

	return nil
}

func TestCreateMessage(ctx context.Context) error {
	count := 0
	for range 3 {
		count++
		msg, err := ChatService.CreatePrivateMessage(ctx, &chat.CreatePrivateMessageRequest{
			SenderId:       usrA.UserId,
			InterlocutorId: usrB.UserId,
			MessageText:    utils.RandomString(10, false),
		})
		if err != nil {
			return err
		}
		fmt.Printf("Message %d: %s\n", count, ce.FormatValue(mapping.MapPMFromProto(msg)))
		count++
		msg, err = ChatService.CreatePrivateMessage(ctx, &chat.CreatePrivateMessageRequest{
			SenderId:       usrB.UserId,
			InterlocutorId: usrA.UserId,
			MessageText:    utils.RandomString(10, false),
		})
		if err != nil {
			return err
		}
		fmt.Printf("Message %d: %s", count, ce.FormatValue(mapping.MapPMFromProto(msg)))
	}
	return nil
}

func TestGetConversations(ctx context.Context) error {
	later := time.Now().AddDate(100, 0, 0)
	res, err := ChatService.GetPrivateConversations(ctx, &chat.GetPrivateConversationsRequest{
		UserId:     usrA.UserId,
		BeforeDate: timestamppb.New(later),
		Limit:      10,
	})
	if err != nil {
		return err
	}
	fmt.Printf("CONVERSATIONS USER A +++++++++++\n %s", ce.FormatValue(mapping.MapConversationsFromProto(res.Conversations)))
	resB, err := ChatService.GetPrivateConversations(ctx, &chat.GetPrivateConversationsRequest{
		UserId:     usrB.UserId,
		BeforeDate: timestamppb.New(later),
		Limit:      10,
	})
	if err != nil {
		return err
	}
	fmt.Printf("CONVERSATIONS USER B +++++++++++\n %s", ce.FormatValue(mapping.MapConversationsFromProto(resB.Conversations)))
	return nil
}

func TestGetPMs(ctx context.Context) error {
	resUserA, err := ChatService.GetPreviousPrivateMessages(ctx, &chat.GetPrivateMessagesRequest{
		UserId:            usrA.UserId,
		InterlocutorId:    usrB.UserId,
		BoundaryMessageId: 0,
		Limit:             10,
		RetrieveUsers:     true,
	})
	if err != nil {
		return err
	}
	modelUsrA := mapping.MapGetPMsRespFromProto(resUserA)

	resUserB, err := ChatService.GetPreviousPrivateMessages(ctx, &chat.GetPrivateMessagesRequest{
		UserId:            usrB.UserId,
		InterlocutorId:    usrA.UserId,
		BoundaryMessageId: 0,
		Limit:             10,
		RetrieveUsers:     true,
	})
	if err != nil {
		return err
	}
	modelUsrB := mapping.MapGetPMsRespFromProto(resUserB)
	if !slices.Equal(modelUsrA.Messages, modelUsrB.Messages) {
		return errors.New("user a and user b messages are not equal")
	}
	fmt.Printf("MESSAGES ++++++++++++\n %s", ce.FormatValue(modelUsrA))
	return nil
}

func TestGetNextPms(ctx context.Context) error {
	res, err := ChatService.GetNextPrivateMessages(ctx, &chat.GetPrivateMessagesRequest{
		UserId:            usrA.UserId,
		InterlocutorId:    usrB.UserId,
		BoundaryMessageId: 1,
		Limit:             10,
		RetrieveUsers:     true,
	})
	if err != nil {
		return err
	}
	fmt.Printf("MESSAGES ++++++++++++\n %s", ce.FormatValue(mapping.MapGetPMsRespFromProto(res)))
	return nil
}

func TestUnreadCount(ctx context.Context) error {
	msg, err := ChatService.CreatePrivateMessage(ctx, &chat.CreatePrivateMessageRequest{
		SenderId:       usrA.UserId,
		InterlocutorId: usrB.UserId,
		MessageText:    "test test test",
	})
	if err != nil {
		return err
	}
	fmt.Printf("Created Message: %s\n", ce.FormatValue(mapping.MapPMFromProto(msg)))

	later := time.Now().AddDate(100, 0, 0)
	resA, err := ChatService.GetPrivateConversations(ctx, &chat.GetPrivateConversationsRequest{
		UserId:     usrA.UserId,
		Limit:      1,
		BeforeDate: timestamppb.New(later),
	})
	if err != nil {
		return err
	}

	if resA.Conversations[0].UnreadCount != 0 {
		return fmt.Errorf("expected 0 unread got: %d", resA.Conversations[0].UnreadCount)
	}

	// fmt.Println("GetConvs A Res", ce.FormatValue(mapping.MapConversationsFromProto(resA.Conversations)))

	resB, err := ChatService.GetPrivateConversations(ctx, &chat.GetPrivateConversationsRequest{
		UserId:     usrB.UserId,
		Limit:      1,
		BeforeDate: timestamppb.New(later),
	})
	if err != nil {
		return err
	}
	fmt.Println("GetConvs B Res", ce.FormatValue(mapping.MapConversationsFromProto(resB.Conversations)))
	if resB.Conversations[0].UnreadCount == 0 {
		return fmt.Errorf("expected at least 1 unread got: 0")
	}
	return nil
}

func TestGetConversationsCountWithUnreadMsgs(ctx context.Context) error {
	msg, err := ChatService.CreatePrivateMessage(ctx, &chat.CreatePrivateMessageRequest{
		SenderId:       usrA.UserId,
		InterlocutorId: usrB.UserId,
		MessageText:    "test test test",
	})
	if err != nil {
		return err
	}
	fmt.Printf("Created Message: %s\n", ce.FormatValue(mapping.MapPMFromProto(msg)))
	count, err := ChatService.GetConvsWithUnreadsCount(ctx, &chat.GetConvsWithUnreadsCountRequest{UserId: usrB.UserId})
	if err != nil {
		return err
	}
	fmt.Printf("Convs with Unread Count: %d\n", count.Count)
	return nil
}

func TestGetConversationsBefore(ctx context.Context) error {
	msg, err := ChatService.CreatePrivateMessage(ctx, &chat.CreatePrivateMessageRequest{
		SenderId:       usrC.UserId,
		InterlocutorId: usrA.UserId,
		MessageText:    "C: test test",
	})
	if err != nil {
		return err
	}

	conv, err := ChatService.GetPrivateConversations(ctx, &chat.GetPrivateConversationsRequest{
		UserId:     usrA.UserId,
		BeforeDate: msg.CreatedAt,
		Limit:      10,
	})
	if err != nil {
		return err
	}
	fmt.Println(mapping.MapConversationsFromProto(conv.Conversations))

	// Fetch a conversation
	now := msg.CreatedAt.AsTime().Add(1 * time.Second)
	conv, err = ChatService.GetPrivateConversations(ctx, &chat.GetPrivateConversationsRequest{
		UserId:     usrA.UserId,
		BeforeDate: timestamppb.New(now),
		Limit:      10,
	})
	if err != nil {
		return err
	}
	fmt.Println("Fetching conversation")
	fmt.Println(ce.FormatValue(mapping.MapConversationsFromProto(conv.Conversations)))

	// Fetch conversation BeforeDate of fetched.UpdatedAt
	fetched := mapping.MapConversationFromProto(conv.Conversations[0])
	conv, err = ChatService.GetPrivateConversations(ctx, &chat.GetPrivateConversationsRequest{
		UserId:     usrA.UserId,
		BeforeDate: fetched.UpdatedAt.ToProto(),
		Limit:      10,
	})
	if err != nil {
		return err
	}
	fmt.Println("Fetch conversation BeforeDate of fetched.UpdatedAt")
	fmt.Println(ce.FormatValue(mapping.MapConversationsFromProto(conv.Conversations)))
	return nil
}

func TestGroupConversation(ctx context.Context) error {
	// Create the group
	groupId, err := UsersService.CreateGroup(ctx,
		&users.CreateGroupRequest{
			OwnerId:          usrA.UserId,
			GroupTitle:       "My Group",
			GroupDescription: "testing testing testing",
		})
	if err != nil {
		return err
	}

	_, err = UsersService.InviteToGroup(ctx, &users.InviteToGroupRequest{
		InviterId:  usrA.UserId,
		InvitedIds: &common.UserIds{Values: []int64{usrB.UserId}},
		GroupId:    groupId.Value,
	})
	if err != nil {
		fmt.Println("invite failed")
		return err
	}

	_, err = UsersService.RespondToGroupInvite(ctx, &users.HandleGroupInviteRequest{
		GroupId:   groupId.Value,
		InvitedId: usrB.UserId,
		Accepted:  true,
	})
	if err != nil {
		fmt.Println("Acceptance failed")
		return err
	}

	msg, err := ChatService.CreateGroupMessage(ctx,
		&chat.CreateGroupMessageRequest{
			GroupId:     groupId.Value,
			SenderId:    usrA.UserId,
			MessageText: "A: test test test",
		})
	if err != nil {
		return err
	}
	fmt.Printf("created group msg: %v", ce.FormatValue(mapping.MapGroupMessageFromProto(msg)))

	msg, err = ChatService.CreateGroupMessage(ctx,
		&chat.CreateGroupMessageRequest{
			GroupId:     groupId.Value,
			SenderId:    usrB.UserId,
			MessageText: "B: test test test",
		})
	if err != nil {
		return err
	}
	fmt.Printf("created group msg: %v", ce.FormatValue(mapping.MapGroupMessageFromProto(msg)))

	return nil
}

func TestGetGroupMessages(ctx context.Context) error {
	msgs, err := ChatService.GetPreviousGroupMessages(ctx,
		&chat.GetGroupMessagesRequest{
			GroupId:           1,
			MemberId:          usrA.UserId,
			BoundaryMessageId: 0,
			Limit:             10,
		})
	if err != nil {
		return err
	}
	fmt.Printf("fetched group msg: %v", ce.FormatValue(mapping.MapGroupMessagesFromProto(msgs.Messages)))

	msgs, err = ChatService.GetPreviousGroupMessages(ctx,
		&chat.GetGroupMessagesRequest{
			GroupId:           1,
			MemberId:          usrB.UserId,
			BoundaryMessageId: 0,
			Limit:             10,
			RetrieveUsers:     true,
		})
	if err != nil {
		return err
	}
	fmt.Printf("fetched group msg: %v", ce.FormatValue(mapping.MapGroupMessagesFromProto(msgs.Messages)))
	return nil
}

func AppendStructToUsersJSON[T any](item T) error {
	const filename = "users.json"

	var items []T

	// Try reading existing file
	data, err := os.ReadFile(filename)
	if err == nil && len(data) > 0 {
		// File exists → unmarshal existing array
		if err := json.Unmarshal(data, &items); err != nil {
			return err
		}
	} else if err != nil && os.IsNotExist(err) {
		// File does not exist → create it with empty JSON array
		empty := []byte("[]")
		if err := os.WriteFile(filename, empty, 0644); err != nil {
			return err
		}
		items = []T{} // initialize slice
	} else {
		// Unexpected read error
		return err
	}

	// Append new item
	items = append(items, item)

	// Marshal back to JSON
	data, err = json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}

	// Create or overwrite file with updated array
	return os.WriteFile(filename, data, 0644)
}

func ReadUserFromJSON() ([]users.RegisterUserResponse, error) {
	data, err := os.ReadFile("users.json")
	if err != nil {
		return nil, err
	}

	var users []users.RegisterUserResponse
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, err
	}

	return users, nil
}
