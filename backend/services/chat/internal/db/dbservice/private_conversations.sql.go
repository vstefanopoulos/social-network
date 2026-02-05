package dbservice

import (
	"context"
	"errors"
	"fmt"
	"math"
	ce "social-network/shared/go/commonerrors"
	ct "social-network/shared/go/ct"
	md "social-network/shared/go/models"

	"github.com/jackc/pgx/v5"
)

type NewPrivateConversation struct {
	Id                 ct.Id
	UserA              ct.Id
	UserB              ct.Id
	LastReadMessageIdA ct.Id `validation:"nullable"`
	LastReadMessageIdB ct.Id `validation:"nullable"`
	CreatedAt          ct.GenDateTime
	UpdatedAt          ct.GenDateTime
	DeletedAt          ct.GenDateTime `validation:"nullable"`
}

func (q *Queries) GetPrivateConvById(ctx context.Context,
	arg md.GetPrivateConvByIdReq,
) (res md.PrivateConvsPreview, err error) {
	input := fmt.Sprintf("arg: %#v", arg)

	if err := ct.ValidateStruct(arg); err != nil {
		return res, ce.New(ce.ErrInvalidArgument, err, input)
	}

	row := q.db.QueryRow(
		ctx,
		getPrivateConvById,
		arg.UserId,
		arg.ConversationId,
	)

	err = row.Scan(
		&res.ConversationId,
		&res.UpdatedAt,
		&res.Interlocutor.UserId,
		&res.LastMessage.Id,
		&res.LastMessage.Sender.UserId,
		&res.LastMessage.MessageText,
		&res.LastMessage.CreatedAt,
		&res.UnreadCount,
	)

	if err == pgx.ErrNoRows {
		return res, ce.New(ce.ErrNotFound, err, input).WithPublic("conversation not found")
	}

	if err != nil {
		return res, ce.New(ce.ErrInternal, err, input)
	}
	res.LastMessage.ConversationId = res.ConversationId

	return res, nil
}

func (q *Queries) GetPrivateConvs(ctx context.Context,
	arg md.GetPrivateConvsReq) (res []md.PrivateConvsPreview, err error) {
	input := fmt.Sprintf("arg: %#v", arg)

	if err := ct.ValidateStruct(arg); err != nil {
		return res, ce.New(ce.ErrInvalidArgument, err, input)
	}

	rows, err := q.db.Query(ctx,
		getPrivateConvs,
		arg.UserId.Int64(),
		arg.BeforeDateUpdated.Time(),
		arg.Limit,
	)
	if err != nil {
		return res, ce.New(ce.ErrInternal, err, input)
	}
	defer rows.Close()

	for rows.Next() {
		var pc md.PrivateConvsPreview
		err := rows.Scan(
			&pc.ConversationId,
			&pc.UpdatedAt,
			&pc.Interlocutor.UserId,
			&pc.LastMessage.Id,
			&pc.LastMessage.Sender.UserId,
			&pc.LastMessage.MessageText,
			&pc.LastMessage.CreatedAt,
			&pc.UnreadCount,
		)
		if err != nil {
			return res, ce.New(ce.ErrInternal, err, input)
		}
		res = append(res, pc)
	}

	return res, nil
}

func (q *Queries) GetConvsWithUnreadsCount(ctx context.Context, userId ct.Id) (count int, err error) {
	input := fmt.Sprintf("userId: %v", userId)
	row := q.db.QueryRow(ctx,
		getConvsWithUnreadsCount,
		userId,
	)
	err = row.Scan(&count)
	if err != nil {
		return count, ce.New(ce.ErrInternal, err, input).WithPublic("internal error")
	}
	return count, nil
}

// Creates or fetches existing conversation between Sender and Interlocutor and creates a message with reference to conversation id.
func (q *Queries) CreateNewPrivateMessage(ctx context.Context, arg md.CreatePrivateMsgReq) (msg md.PrivateMsg, err error) {
	input := fmt.Sprintf("arg: %#v", arg)

	userA := min(arg.SenderId.Int64(), arg.InterlocutorId.Int64())
	userB := max(arg.SenderId.Int64(), arg.InterlocutorId.Int64())
	row := q.db.QueryRow(ctx,
		createMsgAndConv,
		userA,
		userB,
		arg.SenderId,
		arg.MessageText,
	)

	err = row.Scan(
		&msg.Id,
		&msg.ConversationId,
		&msg.Sender.UserId,
		&msg.MessageText,
		&msg.CreatedAt,
		&msg.UpdatedAt,
		&msg.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return msg, ce.New(ce.ErrPermissionDenied, err, input).
				WithPublic("conversation is deleted")
		}
		return msg, ce.New(ce.ErrInternal, err, input)
	}
	return msg, err
}

func (q *Queries) GetPrevPrivateMsgs(ctx context.Context,
	arg md.GetPrivateMsgsReq) (res md.GetPrivateMsgsResp, err error) {
	input := fmt.Sprintf("arg: %#v", arg)

	if err := ct.ValidateStruct(arg); err != nil {
		return res, ce.New(ce.ErrInvalidArgument, err, input)
	}

	if arg.BoundaryMessageId == 0 {
		arg.BoundaryMessageId = math.MaxInt64
	}

	rows, err := q.db.Query(ctx,
		getPrevPrivateMsgs,
		arg.UserId,
		arg.InterlocutorId,
		arg.BoundaryMessageId,
		arg.Limit+1,
	)
	if err != nil {
		return res, ce.New(ce.ErrInternal, err, input)
	}
	defer rows.Close()

	for rows.Next() {
		var message md.PrivateMsg
		if err := rows.Scan(
			&message.Id,
			&message.ConversationId,
			&message.Sender.UserId,
			&message.MessageText,
			&message.CreatedAt,
			&message.UpdatedAt,
			&message.DeletedAt,
		); err != nil {
			return res, ce.New(ce.ErrInternal, err, input)
		}
		res.Messages = append(res.Messages, message)
	}

	if len(res.Messages) > int(arg.Limit) {
		res.Messages = res.Messages[:arg.Limit]
		res.HaveMore = true
	}

	return res, nil
}

func (q *Queries) GetNextPrivateMsgs(ctx context.Context,
	arg md.GetPrivateMsgsReq) (res md.GetPrivateMsgsResp, err error) {
	input := fmt.Sprintf("arg: %#v", arg)

	if err := ct.ValidateStruct(arg); err != nil {
		return res, ce.New(ce.ErrInvalidArgument, err, input)
	}

	if arg.BoundaryMessageId == 0 {
		arg.BoundaryMessageId = math.MinInt64
	}

	rows, err := q.db.Query(ctx,
		getNextPrivateMsgs,
		arg.UserId,
		arg.InterlocutorId,
		arg.BoundaryMessageId,
		arg.Limit+1,
	)
	if err != nil {
		return res, ce.New(ce.ErrInternal, err, input)
	}
	defer rows.Close()

	for rows.Next() {
		var message md.PrivateMsg
		if err := rows.Scan(
			&message.Id,
			&message.ConversationId,
			&message.Sender.UserId,
			&message.MessageText,
			&message.CreatedAt,
			&message.UpdatedAt,
			&message.DeletedAt,
		); err != nil {
			return res, ce.New(ce.ErrInternal, err, input)
		}
		res.Messages = append(res.Messages, message)
	}

	if len(res.Messages) > int(arg.Limit) {
		res.Messages = res.Messages[:arg.Limit]
		res.HaveMore = true
	}

	return res, nil
}

func (q *Queries) UpdateLastReadPrivateMsg(ctx context.Context, arg md.UpdateLastReadMsgParams) error {
	input := fmt.Sprintf("arg: %#v", arg)
	if err := ct.ValidateStruct(arg); err != nil {
		return ce.New(ce.ErrInvalidArgument, err, input)
	}
	res, err := q.db.Exec(ctx, updateLastReadMessage,
		arg.ConversationId,
		arg.UserId,
		arg.LastReadMessageId,
	)
	if err != nil {
		return ce.New(ce.ErrInternal, err, input)
	}

	rows := res.RowsAffected()

	if rows == 0 {
		// Either:
		// - conversation does not exist
		// - user is not part of conversation
		// - conversation is deleted
		return ce.New(ce.ErrNotFound, err, input).WithPublic("conversation not found or access denied")
	}
	return nil
}
