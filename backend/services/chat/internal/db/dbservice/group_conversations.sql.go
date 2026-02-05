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

func (q *Queries) CreateNewGroupMessage(ctx context.Context,
	arg md.CreateGroupMsgReq) (msg md.GroupMsg, err error) {
	input := fmt.Sprintf("arg: %#v", arg)

	row := q.db.QueryRow(ctx,
		createGroupMessage,
		arg.GroupId,
		arg.SenderId,
		arg.MessageText,
	)

	err = row.Scan(
		&msg.Id,
		&msg.GroupId,
		&msg.Sender.UserId,
		&msg.MessageText,
		&msg.CreatedAt,
		&msg.UpdatedAt,
		&msg.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return msg, ce.New(ce.ErrInvalidArgument, err, input).
				WithPublic("conversation does not exist or is deleted")
		}
		return msg, ce.New(ce.ErrInternal, err, input)
	}
	return msg, err
}

func (q *Queries) GetPrevGroupMessages(ctx context.Context,
	req md.GetGroupMsgsReq) (res md.GetGroupMsgsResp, err error) {
	input := fmt.Sprintf("arg: %#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return res, ce.New(ce.ErrInvalidArgument, err, input)
	}

	if req.BoundaryMessageId == 0 {
		req.BoundaryMessageId = math.MaxInt64
	}

	rows, err := q.db.Query(ctx,
		getPrevGroupMsgs,
		req.GroupId,
		// req.UserId,
		req.BoundaryMessageId,
		req.Limit+1,
	)
	if err != nil {
		return res, ce.New(ce.ErrInternal, err, input)
	}
	defer rows.Close()

	for rows.Next() {
		var message md.GroupMsg
		if err := rows.Scan(
			&message.Id,
			&message.GroupId,
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

	if len(res.Messages) > int(req.Limit) {
		res.Messages = res.Messages[:req.Limit]
		res.HaveMore = true
	}
	return res, nil
}

func (q *Queries) GetNextGroupMessages(ctx context.Context,
	arg md.GetGroupMsgsReq) (res md.GetGroupMsgsResp, err error) {
	input := fmt.Sprintf("arg: %#v", arg)

	if err := ct.ValidateStruct(arg); err != nil {
		return res, ce.New(ce.ErrInvalidArgument, err, input)
	}

	if arg.BoundaryMessageId == 0 {
		arg.BoundaryMessageId = math.MinInt64
	}

	rows, err := q.db.Query(ctx,
		getNextGroupMsgs,
		arg.GroupId,
		// arg.UserId,
		arg.BoundaryMessageId,
		arg.Limit+1,
	)
	if err != nil {
		return res, ce.New(ce.ErrInternal, err, input)
	}
	defer rows.Close()

	for rows.Next() {
		var message md.GroupMsg
		if err := rows.Scan(
			&message.Id,
			&message.GroupId,
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
