package handler

import (
	"context"

	pb "social-network/shared/gen-go/chat"
	ce "social-network/shared/go/commonerrors"
	"social-network/shared/go/ct"
	mp "social-network/shared/go/mapping"
	md "social-network/shared/go/models"
	tele "social-network/shared/go/telemetry"

	"google.golang.org/protobuf/types/known/emptypb"

	_ "github.com/lib/pq"
)

// Retrieves a single conversation preview by conversation id.
// Includes conversation ID, last update time, interlocutor,
// last message, and unread count.
func (h *ChatHandler) GetPrivateConversationById(
	ctx context.Context,
	params *pb.GetPrivateConversationByIdRequest,
) (*pb.PrivateConversationPreview, error) {
	tele.Info(ctx, "get private conversation by id called: @1", "params", params.String())

	conv, err := h.Application.GetPrivateConversationById(ctx, md.GetPrivateConvByIdReq{
		UserId:         ct.Id(params.UserId),
		ConversationId: ct.Id(params.ConversationId),
		InterlocutorId: ct.Id(params.InterlocutorId),
	})
	if err != nil {
		tele.Error(ctx, "get private conversations by id @1 \n\n@2\n\n",
			"request", params,
			"error", err.Error(),
		)
		return nil, ce.EncodeProto(err)
	}

	res := mp.MapConversationToProto(conv)

	tele.Info(ctx, "get private conversations succes: @1 @2",
		"response", res,
	)
	return res, nil
}

// Retrieves a paginated list of private conversations for a user.
func (h *ChatHandler) GetPrivateConversations(
	ctx context.Context,
	params *pb.GetPrivateConversationsRequest,
) (*pb.GetPrivateConversationsResponse, error) {
	tele.Info(ctx, "get private conversations called: @1", "params", params)

	convs, err := h.Application.GetPrivateConversations(ctx, md.GetPrivateConvsReq{
		UserId:            ct.Id(params.UserId),
		BeforeDateUpdated: ct.GenDateTime(params.BeforeDate.AsTime()),
		Limit:             ct.Limit(params.Limit),
	})
	if err != nil {
		tele.Error(ctx, "get private conversations @1 \n\n@2\n\n",
			"request", params,
			"error", err.Error(),
		)
		return nil, ce.EncodeProto(err)
	}

	res := &pb.GetPrivateConversationsResponse{
		Conversations: mp.MapConversationsToProto(convs),
	}

	tele.Info(ctx, "get private conversations success: @1",
		"response", res,
	)
	return res, nil
}

func (h *ChatHandler) GetConvsWithUnreadsCount(
	ctx context.Context,
	params *pb.GetConvsWithUnreadsCountRequest,
) (*pb.GetConvsWithUnreadsCountResponse, error) {
	tele.Info(ctx, "get conversations with unread count called @1", "userId", params.UserId)
	count, err := h.Application.GetConvsWithUnreadsCount(ctx, ct.Id(params.UserId))
	if err != nil {
		tele.Error(ctx, "get conversations with unread count @1 \n\n@2\n\n",
			"request", params,
			"error", err.Error(),
		)
		return nil, ce.EncodeProto(err)
	}
	return &pb.GetConvsWithUnreadsCountResponse{Count: int64(count)}, nil
}

// Creates a new private message and returns the created message details.
func (h *ChatHandler) CreatePrivateMessage(
	ctx context.Context,
	params *pb.CreatePrivateMessageRequest,
) (*pb.PrivateMessage, error) {
	tele.Info(ctx, "creating private message: @1", "params", params)

	// Call application layer
	msg, Err := h.Application.CreatePrivateMessage(ctx, md.CreatePrivateMsgReq{
		SenderId:       ct.Id(params.SenderId),
		InterlocutorId: ct.Id(params.InterlocutorId),
		MessageText:    ct.MsgBody(params.MessageText),
	})
	if Err != nil {
		tele.Error(ctx, "create private message @1 \n\n@2\n\n",
			"request", params,
			"error", Err.Error(),
		)
		return nil, ce.EncodeProto(Err)
	}

	resp := mp.MapPMToProto(msg)

	tele.Info(ctx, "create private message success. @1",
		"response", resp,
	)

	return resp, nil
}

// Retrieves previous private messages (older than the boundary message) in a conversation.
func (h *ChatHandler) GetPreviousPrivateMessages(
	ctx context.Context,
	params *pb.GetPrivateMessagesRequest,
) (*pb.GetPrivateMessagesResponse, error) {

	tele.Info(ctx, "get previous private messages called @1", "request", params)

	// Call application layer
	res, err := h.Application.GetPreviousPMs(ctx, md.GetPrivateMsgsReq{
		InterlocutorId:    ct.Id(params.InterlocutorId),
		UserId:            ct.Id(params.UserId),
		BoundaryMessageId: ct.Id(params.BoundaryMessageId),
		Limit:             ct.Limit(params.Limit),
		RetrieveUsers:     params.RetrieveUsers,
	})
	if err != nil {
		tele.Error(ctx, "get previous private messages error",
			"request", params,
			"error", err.Error(),
		)
		return nil, ce.EncodeProto(err)
	}

	resp := mp.MapGetPMsResp(res)

	tele.Info(ctx, "get previous private messages success. @1 @2",
		"request", params,
		"response", resp,
	)

	return resp, nil
}

// Retrieves next private messages (newer than the boundary message) in a conversation.
func (h *ChatHandler) GetNextPrivateMessages(
	ctx context.Context,
	params *pb.GetPrivateMessagesRequest,
) (*pb.GetPrivateMessagesResponse, error) {

	tele.Info(ctx, "get next private messages called @1", "request", params)

	// Call application layer
	res, err := h.Application.GetNextPMs(ctx, md.GetPrivateMsgsReq{
		InterlocutorId:    ct.Id(params.InterlocutorId),
		UserId:            ct.Id(params.UserId),
		BoundaryMessageId: ct.Id(params.BoundaryMessageId),
		Limit:             ct.Limit(params.Limit),
		RetrieveUsers:     params.RetrieveUsers,
	})
	if err != nil {
		tele.Error(ctx, "get next private messages error",
			"request", params,
			"error", err.Error(),
		)
		return nil, ce.EncodeProto(err)
	}

	resp := mp.MapGetPMsResp(res)

	tele.Info(ctx, "get next private messages success. @1 @2",
		"request", params,
		"response", resp,
	)

	return resp, nil
}

// Updates the last read message pointer for a user in a private conversation.
func (h *ChatHandler) UpdateLastReadPrivateMessage(
	ctx context.Context,
	params *pb.UpdateLastReadPrivateMessageRequest,
) (*emptypb.Empty, error) {

	tele.Info(ctx, "update last read private message called @1", "request", params)

	// Call application layer
	err := h.Application.UpdateLastReadPrivateMsg(ctx, md.UpdateLastReadMsgParams{
		ConversationId:    ct.Id(params.ConversationId),
		UserId:            ct.Id(params.UserId),
		LastReadMessageId: ct.Id(params.LastReadMessageId),
	})
	if err != nil {
		tele.Error(ctx, "update last read private message error",
			"request", params,
			"error", err.Error(),
		)
		return nil, ce.EncodeProto(err)
	}

	resp := &emptypb.Empty{}

	tele.Info(ctx, "update last read private message success. @1 @2",
		"request", params,
		"response", resp,
	)

	return resp, nil
}
