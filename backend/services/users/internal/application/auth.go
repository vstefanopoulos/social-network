package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	ds "social-network/services/users/internal/db/dbservice"
	"social-network/shared/gen-go/media"
	ce "social-network/shared/go/commonerrors"
	ct "social-network/shared/go/ct"
	"social-network/shared/go/models"
	tele "social-network/shared/go/telemetry"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

const genericPublic = "users service error"

func (s *Application) RegisterUser(ctx context.Context, req models.RegisterUserRequest) (models.RegisterUserResponse, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return models.RegisterUserResponse{}, ce.Wrap(ce.ErrInvalidArgument, err, input).WithPublic("invalid data received")
	}

	username := req.Username
	//if no username assign full name
	if username == "" {
		username = ct.Username(string(req.FirstName) + "_" + string(req.LastName))
	}

	// convert date
	dob := pgtype.Date{
		Time:  req.DateOfBirth.Time(),
		Valid: true,
	}

	var newId ct.Id

	err := s.txRunner.RunTx(ctx, func(q *ds.Queries) error {

		// Insert user
		userId, err := q.InsertNewUser(ctx, ds.InsertNewUserParams{
			Username:      username.String(),
			FirstName:     req.FirstName.String(),
			LastName:      req.LastName.String(),
			DateOfBirth:   dob,
			AvatarID:      req.AvatarId.Int64(),
			AboutMe:       req.About.String(),
			ProfilePublic: req.Public,
		})
		if err != nil {
			return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
		}
		newId = ct.Id(userId)

		// Insert auth
		return q.InsertNewUserAuth(ctx, ds.InsertNewUserAuthParams{
			UserID:       newId.Int64(),
			Email:        req.Email.String(),
			PasswordHash: req.Password.String(),
		})
	})

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return models.RegisterUserResponse{}, ce.New(ce.ErrAlreadyExists, err, input).WithPublic("email already exists")
			}
		}
		return models.RegisterUserResponse{}, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}

	return models.RegisterUserResponse{
		UserId:   newId.Int64(),
		Username: username,
	}, nil

}

func (s *Application) LoginUser(ctx context.Context, req models.LoginRequest) (models.User, error) {
	input := fmt.Sprintf("%#v", req)

	var u models.User

	if err := ct.ValidateStruct(req); err != nil {
		return models.User{}, ce.Wrap(ce.ErrInvalidArgument, err, input).WithPublic("invalid data received")
	}

	err := s.txRunner.RunTx(ctx, func(q *ds.Queries) error {
		row, err := q.GetUserForLogin(ctx, ds.GetUserForLoginParams{
			Username:     req.Identifier.String(),
			PasswordHash: req.Password.String(),
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ce.New(ce.ErrInvalidArgument, err, input).WithPublic("wrong credentials")
			}
			return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
		}

		u = models.User{
			UserId:   ct.Id(row.ID),
			Username: ct.Username(row.Username),
			AvatarId: ct.Id(row.AvatarID),
		}

		if u.AvatarId > 0 {
			imageUrl, err := s.mediaRetriever.GetImage(ctx, u.AvatarId.Int64(), media.FileVariant_THUMBNAIL)
			if err != nil {
				tele.Error(ctx, "media retriever failed for @1", "request", u.AvatarId, "error", err.Error()) //log error instead of returning it
				s.removeFailedImage(ctx, err, u.AvatarId.Int64())
			} else {
				u.AvatarURL = imageUrl
			}
		}

		return nil
	})

	if err != nil {
		return models.User{}, ce.Wrap(nil, err)
	}

	return u, nil
}

func (s *Application) UpdateUserPassword(ctx context.Context, req models.UpdatePasswordRequest) error {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return ce.Wrap(ce.ErrInvalidArgument, err, input).WithPublic("invalid data received")
	}

	err := s.txRunner.RunTx(ctx, func(q *ds.Queries) error {
		row, err := q.GetUserPassword(ctx, req.UserId.Int64())
		if err != nil {
			return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
		}

		if !checkPassword(row, req.OldPassword.String()) {
			return ce.New(ce.ErrPermissionDenied, fmt.Errorf("wrong previous password"), input).WithPublic("wrong previous password")
		}

		err = q.UpdateUserPassword(ctx, ds.UpdateUserPasswordParams{
			UserID:       req.UserId.Int64(),
			PasswordHash: req.NewPassword.String(),
		})
		if err != nil {
			return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
		}

		return nil
	})
	if err != nil {
		return ce.Wrap(nil, err)
	}
	return nil
}

func (s *Application) UpdateUserEmail(ctx context.Context, req models.UpdateEmailRequest) error {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return ce.Wrap(ce.ErrInvalidArgument, err, input).WithPublic("invalid data received")
	}

	err := s.db.UpdateUserEmail(ctx, ds.UpdateUserEmailParams{
		UserID: req.UserId.Int64(),
		Email:  req.Email.String(),
	})
	if err != nil {
		return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	return nil
}
