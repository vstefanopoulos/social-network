package mapping

import (
	pb "social-network/shared/gen-go/chat"
	ct "social-network/shared/go/ct"
	md "social-network/shared/go/models"
)

func MapConversationToProto(conv md.PrivateConvsPreview) *pb.PrivateConversationPreview {
	return &pb.PrivateConversationPreview{
		ConversationId: conv.ConversationId.Int64(),
		UpdatedAt:      conv.UpdatedAt.ToProto(),
		Interlocutor:   MapUserToProto(conv.Interlocutor),
		LastMessage:    MapPMToProto(conv.LastMessage),
		UnreadCount:    int32(conv.UnreadCount),
	}
}

func MapConversationFromProto(conv *pb.PrivateConversationPreview) md.PrivateConvsPreview {
	return md.PrivateConvsPreview{
		ConversationId: ct.Id(conv.ConversationId),
		UpdatedAt:      ct.GenDateTime(conv.UpdatedAt.AsTime()),
		Interlocutor:   MapUserFromProto(conv.Interlocutor),
		LastMessage:    MapPMFromProto(conv.LastMessage),
		UnreadCount:    int(conv.UnreadCount),
	}

}

func MapConversationsToProto(cs []md.PrivateConvsPreview) []*pb.PrivateConversationPreview {
	convs := make([]*pb.PrivateConversationPreview, 0, len(cs))
	for _, conv := range cs {
		convs = append(convs, MapConversationToProto(conv))
	}
	return convs
}

func MapConversationsFromProto(cs []*pb.PrivateConversationPreview) []md.PrivateConvsPreview {
	convs := make([]md.PrivateConvsPreview, 0, len(cs))
	for _, conv := range cs {
		convs = append(convs, MapConversationFromProto(conv))
	}
	return convs
}

func MapPMToProto(m md.PrivateMsg) *pb.PrivateMessage {
	return &pb.PrivateMessage{
		Id:             m.Id.Int64(),
		ConversationId: m.ConversationId.Int64(),
		Sender:         MapUserToProto(m.Sender),
		ReceiverId:     m.ReceiverId.Int64(),
		MessageText:    m.MessageText.String(),
		CreatedAt:      m.CreatedAt.ToProto(),
		UpdatedAt:      m.UpdatedAt.ToProto(),
		DeletedAt:      m.DeletedAt.ToProto(),
	}
}

func MapPMFromProto(p *pb.PrivateMessage) md.PrivateMsg {
	if p == nil {
		return md.PrivateMsg{}
	}

	return md.PrivateMsg{
		Id:             ct.Id(p.Id),
		ConversationId: ct.Id(p.ConversationId),
		Sender:         MapUserFromProto(p.Sender),
		ReceiverId:     ct.Id(p.ReceiverId),
		MessageText:    ct.MsgBody(p.MessageText),
		CreatedAt:      ct.GenDateTime(p.CreatedAt.AsTime()),
		UpdatedAt:      ct.GenDateTime(p.UpdatedAt.AsTime()),
		DeletedAt:      ct.GenDateTime(p.DeletedAt.AsTime()),
	}
}

func MapGetPMsResp(res md.GetPrivateMsgsResp) *pb.GetPrivateMessagesResponse {
	msgs := make([]*pb.PrivateMessage, 0, len(res.Messages))
	for _, m := range res.Messages {
		msgs = append(msgs, MapPMToProto(m))
	}

	return &pb.GetPrivateMessagesResponse{
		HaveMore: res.HaveMore,
		Messages: msgs,
	}
}

func MapGetPMsRespFromProto(p *pb.GetPrivateMessagesResponse) md.GetPrivateMsgsResp {
	if p == nil {
		return md.GetPrivateMsgsResp{}
	}

	msgs := make([]md.PrivateMsg, 0, len(p.Messages))
	for _, m := range p.Messages {
		msgs = append(msgs, MapPMFromProto(m))
	}

	return md.GetPrivateMsgsResp{
		HaveMore: p.HaveMore,
		Messages: msgs,
	}
}

// MapGroupMessagesToProto maps group message models to proto messages.
func MapGroupMessagesToProto(msgs []md.GroupMsg) []*pb.GroupMessage {
	if len(msgs) == 0 {
		return []*pb.GroupMessage{}
	}

	out := make([]*pb.GroupMessage, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, MapGroupMessageToProto(m))
	}

	return out
}

func MapGroupMessagesFromProto(msgs []*pb.GroupMessage) []md.GroupMsg {
	if len(msgs) == 0 {
		return []md.GroupMsg{}
	}

	out := make([]md.GroupMsg, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, MapGroupMessageFromProto(m))
	}

	return out
}

func MapGroupMessageToProto(m md.GroupMsg) *pb.GroupMessage {
	return &pb.GroupMessage{
		Id:          m.Id.Int64(),
		GroupId:     m.GroupId.Int64(),
		Sender:      MapUserToProto(m.Sender),
		MessageText: m.MessageText.String(),
		CreatedAt:   m.CreatedAt.ToProto(),
		UpdatedAt:   m.UpdatedAt.ToProto(),
		DeletedAt:   m.DeletedAt.ToProto(),
	}
}

func MapGroupMessageFromProto(m *pb.GroupMessage) md.GroupMsg {
	return md.GroupMsg{
		Id:          ct.Id(m.Id),
		GroupId:     ct.Id(m.GroupId),
		Sender:      MapUserFromProto(m.Sender),
		MessageText: ct.MsgBody(m.MessageText),
		CreatedAt:   ct.GenDateTime(m.CreatedAt.AsTime()),
		UpdatedAt:   ct.GenDateTime(m.UpdatedAt.AsTime()),
		DeletedAt:   ct.GenDateTime(m.DeletedAt.AsTime()),
	}
}
