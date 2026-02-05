package dbservice

import (
	"context"
	ct "social-network/shared/go/ct"
	md "social-network/shared/go/models"
)

type Querier interface {
	// Creates a message with message body sender id and conversation Id.
	// Returns message and classified error using commonerrors classification.
	CreateNewGroupMessage(ctx context.Context,
		arg md.CreateGroupMsgReq) (msg md.GroupMsg, err error)

	// Creates a new PM if the conversation exist and sender is a member.
	CreateNewPrivateMessage(ctx context.Context,
		arg md.CreatePrivateMsgReq) (msg md.PrivateMsg, err error)

	// Returns a descending-ordered page of messages that appear chronologically
	// BEFORE a given message in a conversation. This query is used for backwards
	// pagination in chat history.
	//
	// Behavior:
	//
	//   - If the supplied BoundaryMessageId is 0, the query automatically
	//     substitutes the conversation's last message as the boundary (inclusive).
	//
	//   - The caller must be a member of the conversation. Membership is enforced
	//     through the private_conversation table.
	//
	//   - Results are ordered by m.id DESC so that the most recent messages in the
	//     requested page appear last. LIMIT/OFFSET is applied after ordering.
	//
	// Returned fields:
	//   - All message fields (id, conversation_id, sender_id, message_text, timestamps)
	//   - HaveMoreBefore bool. If true caller can use the older message as the new boundary.
	//
	// Use case:
	//
	// Scroll-up pagination in chat history.
	//  GetPreviousMessages(ctx context.Context,
	// 	 args md.GetPrevMessagesParams,
	//  ) (resp md.GetPrevMessagesResp, err error)
	GetPrevPrivateMsgs(ctx context.Context,
		arg md.GetPrivateMsgsReq) (res md.GetPrivateMsgsResp, err error)

	// Returns an ascending-ordered page of messages that appear chronologically
	// AFTER a given message in a conversation. This query is used for forward
	// pagination when loading newer messages.
	//
	// Behavior:
	//
	//   - If the supplied BoundaryMessageId ($1) is 0, the query automatically
	//     substitutes the conversation's first_message_id as the boundary.
	//
	//   - Only messages with id > boundary_id are returned.
	//
	//   - Only non-deleted messages are returned (deleted_at IS NULL).
	//
	//   - The caller must be a member of the conversation.
	//
	//   - Results are ordered by m.id ASC so that the oldest messages in the
	//     requested page appear first.
	//
	// Returned fields:
	//   - All message fields (id, conversation_id, sender_id, message_text, timestamps)
	//   - HaveMore bool. True if caller can use last message as new boundary.
	//
	// Use case:
	//
	// Scroll-down pagination or loading new messages after a known point or beggining.
	//  GetNextMessages(ctx context.Context, args md.GetPMsParams,
	//  ) (resp md.GetPMsResp, err error)
	GetNextPrivateMsgs(ctx context.Context,
		arg md.GetPrivateMsgsReq) (res md.GetPrivateMsgsResp, err error)

	GetPrivateConvById(ctx context.Context,
		arg md.GetPrivateConvByIdReq,
	) (md.PrivateConvsPreview, error)

	// Fetches paginated conversation details, conversation members,
	// Ids and unread messages count for a user and a group.
	GetPrivateConvs(ctx context.Context,
		arg md.GetPrivateConvsReq) (res []md.PrivateConvsPreview, err error)

	GetConvsWithUnreadsCount(ctx context.Context, userId ct.Id) (count int, err error)
	// Updates the given user's last read message in given private conversation to given message id.
	UpdateLastReadPrivateMsg(ctx context.Context, arg md.UpdateLastReadMsgParams) error

	// Gets paginated group messages that are updated before a given date time.
	GetPrevGroupMessages(ctx context.Context,
		req md.GetGroupMsgsReq) (msgs md.GetGroupMsgsResp, err error)

	// Gets paginated group messages that are updated after a given date time.
	GetNextGroupMessages(ctx context.Context,
		req md.GetGroupMsgsReq) (msgs md.GetGroupMsgsResp, err error)
}

var _ Querier = (*Queries)(nil)
