package application

import (
	"context"
	"encoding/json"
	"fmt"
	ce "social-network/shared/go/commonerrors"
	ct "social-network/shared/go/ct"
	md "social-network/shared/go/models"
	tele "social-network/shared/go/telemetry"
)

type CreateMessageInGroupReq struct {
	GroupId     ct.Id
	SenderId    ct.Id
	MessageBody ct.MsgBody
}

func (c *ChatService) CreateMessageInGroup(ctx context.Context,
	req CreateMessageInGroupReq) (res md.GroupMsg, Err *ce.Error) {
	input := fmt.Sprintf("params: %#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return res, ce.New(ce.ErrInvalidArgument, err, input)
	}

	if err := c.isMember(ctx, req.GroupId, req.SenderId, input); err != nil {
		return res, ce.Wrap(nil, err)
	}

	// Add message
	res, err := c.Queries.CreateNewGroupMessage(ctx, md.CreateGroupMsgReq{
		GroupId:     req.GroupId,
		SenderId:    req.SenderId,
		MessageText: req.MessageBody,
	})

	res.Sender, err = c.RetriveUsers.GetUser(ctx, res.Sender.UserId)
	if err != nil {
		return res, ce.Wrap(nil, err, input)
	}

	messageBytes, err := json.Marshal(res)
	if err != nil {
		err = ce.New(ce.ErrInternal, err, input)
		tele.Error(ctx, "failed to publish group message to nats: @1", "error", err.Error())
	}

	err = c.NatsConn.Publish(ct.GroupKey(res.GroupId.Int64()), messageBytes)
	if err != nil {
		err = ce.New(ce.ErrInternal, err, input)
		tele.Error(ctx, "failed to publish private message to nats: @1", "error", err.Error())
	}

	if err != nil {
		return res, ce.Wrap(nil, err, input)
	}

	return res, nil
}

func (c *ChatService) GetPrevGroupMessages(ctx context.Context,
	req md.GetGroupMsgsReq) (res md.GetGroupMsgsResp, Err *ce.Error) {

	input := fmt.Sprintf("req: %#v", req)
	if err := ct.ValidateStruct(req); err != nil {
		return res, ce.Wrap(ce.ErrInvalidArgument, err, input)
	}

	if err := c.isMember(ctx, req.GroupId, req.UserId, input); err != nil {
		return res, ce.Wrap(nil, err)
	}

	res, err := c.Queries.GetPrevGroupMessages(ctx, req)
	if err != nil {
		return res, ce.Wrap(nil, err, input)
	}

	if req.RetrieveUsers {
		err = c.retrieveMessageSenders(ctx, res.Messages, input)
		if err != nil {
			tele.Error(ctx, "failed to retrieve users for messages", "input", input, "error", err)
		}
	}

	return res, nil
}

func (c *ChatService) GetNextGroupMessages(ctx context.Context,
	req md.GetGroupMsgsReq) (res md.GetGroupMsgsResp, Err *ce.Error) {

	input := fmt.Sprintf("req: %#v", req)
	if err := ct.ValidateStruct(req); err != nil {
		return res, ce.Wrap(ce.ErrInvalidArgument, err, input)
	}

	if err := c.isMember(ctx, req.GroupId, req.UserId, input); err != nil {
		return res, ce.Wrap(nil, err)
	}

	res, err := c.Queries.GetNextGroupMessages(ctx, req)
	if err != nil {
		return res, ce.Wrap(nil, err, input)
	}

	if req.RetrieveUsers {
		err = c.retrieveMessageSenders(ctx, res.Messages, input)
		if err != nil {
			tele.Error(ctx, "failed to retrieve users for messages", "input", input, "error", err)
		}
	}

	return res, nil
}

func (c *ChatService) retrieveMessageSenders(ctx context.Context, msgs []md.GroupMsg, input string) error {
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

// Returns a commonerrors Error type with public message if user is not a group member.
// Input is the caller function contextual details.
func (c *ChatService) isMember(ctx context.Context, groupId, memberId ct.Id, input string) *ce.Error {
	isMember, err := c.Clients.IsGroupMember(ctx, groupId, memberId)
	if err != nil {
		return ce.DecodeProto(err, input)
	}

	if !isMember {
		return ce.New(ce.ErrPermissionDenied,
			fmt.Errorf("user id: %d not a member of group id: %d",
				memberId, groupId),
			input,
		).WithPublic("user not a group member")
	}
	return nil
}
