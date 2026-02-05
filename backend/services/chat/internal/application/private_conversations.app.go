package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"social-network/services/chat/internal/db/dbservice"
	ce "social-network/shared/go/commonerrors"
	ct "social-network/shared/go/ct"
	md "social-network/shared/go/models"
	tele "social-network/shared/go/telemetry"
)

var (
	ErrNotConnected = errors.New("users are not connected")
)

func (c *ChatService) GetPrivateConversationById(ctx context.Context,
	arg md.GetPrivateConvByIdReq,
) (res md.PrivateConvsPreview, Err *ce.Error) {
	input := fmt.Sprintf("arg: %#v", arg)
	if err := ct.ValidateStruct(arg); err != nil {
		return md.PrivateConvsPreview{}, ce.Wrap(ce.ErrInvalidArgument, err, input)
	}

	areConnected, Err := c.Clients.AreConnected(ctx, arg.UserId, arg.InterlocutorId)
	if Err != nil {
		return res, Err
	}
	if !areConnected {
		return res, ce.New(ce.ErrPermissionDenied, ErrNotConnected, input).WithPublic("users are not connected")
	}

	conv, err := c.Queries.GetPrivateConvById(ctx, arg)
	if err != nil {
		return res, ce.Wrap(nil, Err, input)
	}

	conv.Interlocutor, err = c.RetriveUsers.GetUser(ctx, conv.Interlocutor.UserId)
	if err != nil {
		return conv, ce.Wrap(nil, err, input)
	}

	return conv, nil
}

// Returns a sorted paginated list of private conversations
// older that the given BeforeDate where user with UserId is a member.
// Respose per PC includes last message and unread count from users side.
func (c *ChatService) GetPrivateConversations(ctx context.Context,
	arg md.GetPrivateConvsReq,
) ([]md.PrivateConvsPreview, *ce.Error) {

	input := fmt.Sprintf("arg: %#v", arg)

	err := ct.ValidateStruct(arg)
	if err != nil {
		return nil, ce.Wrap(ce.ErrInvalidArgument, err, input)
	}

	conversations, err := c.Queries.GetPrivateConvs(ctx, arg)
	if err != nil {
		return conversations, ce.Wrap(nil, err, input)
	}

	allMemberIDs := make(ct.Ids, 0)
	for _, r := range conversations {
		allMemberIDs = append(allMemberIDs, r.Interlocutor.UserId)
	}

	usersMap, err := c.RetriveUsers.GetUsers(ctx, allMemberIDs)
	if err != nil {
		return nil, ce.Wrap(nil, err, input)
	}

	for i := range conversations {
		retrieved := usersMap[conversations[i].Interlocutor.UserId]
		conversations[i].Interlocutor.Username = retrieved.Username
		conversations[i].Interlocutor.AvatarId = retrieved.AvatarId
		conversations[i].Interlocutor.AvatarURL = retrieved.AvatarURL
	}

	return conversations, nil
}

func (c *ChatService) GetConvsWithUnreadsCount(ctx context.Context, userId ct.Id) (count int, err error) {
	if err := userId.Validate(); err != nil {
		return 0, ce.Wrap(ce.ErrInvalidArgument, err)
	}
	return c.Queries.GetConvsWithUnreadsCount(ctx, userId)
}

// Creates a private message and returns an id
func (c *ChatService) CreatePrivateMessage(ctx context.Context,
	arg md.CreatePrivateMsgReq) (msg md.PrivateMsg, Err *ce.Error) {

	input := fmt.Sprintf("params: %#v", arg)
	err := ct.ValidateStruct(arg)
	if err != nil {
		return msg, ce.New(ce.ErrInvalidArgument, err, input)
	}

	areConnected, Err := c.Clients.AreConnected(ctx, arg.SenderId, arg.InterlocutorId)
	if Err != nil {
		return msg, Err
	}
	if !areConnected {
		return msg, ce.New(ce.ErrPermissionDenied, ErrNotConnected, input).WithPublic("users are not connected")
	}
	c.txRunner.RunTx(ctx, func(q *dbservice.Queries) error {
		msg, err = q.CreateNewPrivateMessage(ctx, arg)
		if err != nil {
			return err
		}

		q.UpdateLastReadPrivateMsg(ctx, md.UpdateLastReadMsgParams{
			UserId:            arg.SenderId,
			ConversationId:    msg.ConversationId,
			LastReadMessageId: msg.Id,
		})
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return msg, ce.Wrap(nil, err, input)
	}

	messageBytes, err := json.Marshal(msg)
	if err != nil {
		err = ce.New(ce.ErrInternal, err, input)
		tele.Error(ctx, "failed to publish private message to nats: @1", "error", err.Error())
	}

	err = c.NatsConn.Publish(ct.UserKey(arg.InterlocutorId), messageBytes)
	if err != nil {
		err = ce.New(ce.ErrInternal, err, input)
		tele.Error(ctx, "failed to publish private message to nats: @1", "error", err.Error())
	}
	msg.ReceiverId = arg.InterlocutorId
	return msg, nil
}

func (c *ChatService) GetPreviousPMs(ctx context.Context,
	arg md.GetPrivateMsgsReq) (res md.GetPrivateMsgsResp, Err *ce.Error) {
	input := fmt.Sprintf("arg: %#v", arg)

	if err := ct.ValidateStruct(arg); err != nil {
		return res, ce.New(ce.ErrInvalidArgument, err, input)
	}

	res, err := c.Queries.GetPrevPrivateMsgs(ctx, arg)
	if err != nil {
		return res, ce.Wrap(nil, err, input)
	}

	if arg.RetrieveUsers {
		if err := c.retrievePrivateMessageSenders(ctx, res.Messages, input); err != nil {
			tele.Error(ctx, "failed to retrieve users for messages", "input", input, "error", err)
		}
	}

	return res, nil
}

func (c *ChatService) GetNextPMs(ctx context.Context,
	arg md.GetPrivateMsgsReq) (res md.GetPrivateMsgsResp, Err *ce.Error) {
	input := fmt.Sprintf("arg: %#v", arg)

	if err := ct.ValidateStruct(arg); err != nil {
		return res, ce.New(ce.ErrInvalidArgument, err, input)
	}

	res, err := c.Queries.GetNextPrivateMsgs(ctx, arg)
	if err != nil {
		return res, ce.Wrap(nil, err, input)
	}

	if arg.RetrieveUsers {
		if err := c.retrievePrivateMessageSenders(ctx, res.Messages, input); err != nil {
			tele.Error(ctx, "failed to retrieve users for messages", "input", input, "error", err)
		}
	}
	return res, nil
}

func (c *ChatService) UpdateLastReadPrivateMsg(ctx context.Context, arg md.UpdateLastReadMsgParams) *ce.Error {
	input := fmt.Sprintf("arg: %#v", arg)
	tele.Debug(ctx, "update last read message called @1", "input:", input)

	if err := ct.ValidateStruct(arg); err != nil {
		return ce.New(ce.ErrInvalidArgument, err, input)
	}

	err := c.Queries.UpdateLastReadPrivateMsg(ctx, arg)
	if err != nil {
		tele.Error(ctx, "failed to publish private message to nats: @1", "error", err.Error())
		return ce.Wrap(nil, err, input)
	}
	return nil
}

func (c *ChatService) retrievePrivateMessageSenders(ctx context.Context, msgs []md.PrivateMsg, input string) error {
	allMemberIDs := make(ct.Ids, 0)
	for _, r := range msgs {
		allMemberIDs = append(allMemberIDs, r.Sender.UserId)
	}

	usersMap, err := c.RetriveUsers.GetUsers(ctx, allMemberIDs)
	if err != nil {
		return ce.Wrap(nil, err, input)
	}

	for i := range msgs {
		retrieved := usersMap[msgs[i].Sender.UserId]
		msgs[i].Sender.Username = retrieved.Username
		msgs[i].Sender.AvatarId = retrieved.AvatarId
		msgs[i].Sender.AvatarURL = retrieved.AvatarURL
	}
	return nil
}
