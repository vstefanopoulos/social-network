package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	ce "social-network/shared/go/commonerrors"
	ct "social-network/shared/go/ct"
	"social-network/shared/go/models"
)

// Returns five random ids that fit one of the following criteria:
// Users who liked one or more of *your public posts*
// Users who commented on your public posts
// Users who liked the same posts as you
// Users who commented on the same posts as you
// Actual Basic User Info will be retrieved by HANDLER from users
func (s *Application) SuggestUsersByPostActivity(ctx context.Context, req models.SimpleIdReq) ([]models.User, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return nil, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}
	ids, err := s.db.SuggestUsersByPostActivity(ctx, req.Id.Int64())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []models.User{}, nil
		}
		return nil, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}

	userMap, err := s.userRetriever.GetUsers(ctx, ct.FromInt64s(ids))
	if err != nil {
		return nil, ce.Wrap(nil, err, input).WithPublic("error retrieving user's info")
	}

	users := make([]models.User, 0, len(ids))
	for _, id := range ct.FromInt64s(ids) {
		if u, ok := userMap[id]; ok {
			users = append(users, u)
		}
	}

	return users, nil
}
