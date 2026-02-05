package application

import (
	"context"
	"fmt"
	ds "social-network/services/posts/internal/db/dbservice"
	ce "social-network/shared/go/commonerrors"
	tele "social-network/shared/go/telemetry"
)

// group and post audience=group: only members can see
// post audience=everyone: everyone can see (can we check this before all the fetches from users?)
// post audience=followers: requester can see if they follow creator
// post audience=selected: requester can see if they are in post audience table
func (s *Application) hasRightToView(ctx context.Context, req accessContext) (bool, error) {
	input := fmt.Sprintf("%#v", req)

	row, err := s.db.GetEntityCreatorAndGroup(ctx, req.entityId)
	if err != nil {
		return false, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	tele.Info(ctx, "entity creator: @1, group: @2", "entityCreator", row.CreatorID, "groupId", row.GroupID)

	var targetUserId int64
	if row.ParentCreatorID > 0 { //in case of comment, we need the parent creator for isFollowing
		targetUserId = row.ParentCreatorID
	} else {
		targetUserId = row.CreatorID
	}
	isFollowing, err := s.clients.IsFollowing(ctx, req.requesterId, targetUserId)
	if err != nil {
		return false, ce.DecodeProto(err, input)
	}

	var isMember bool
	if row.GroupID > 0 {
		isMember, err = s.clients.IsGroupMember(ctx, req.requesterId, row.GroupID)
		tele.Info(ctx, "isMember is @1", "isMember", isMember)
		if err != nil {
			return false, ce.DecodeProto(err, input)
		}
	}

	entityID := req.entityId //this is the event or post id - in case of a comment we take the parent post id
	if row.ParentID > 0 {
		entityID = row.ParentID
	}

	canSee, err := s.db.CanUserSeeEntity(ctx, ds.CanUserSeeEntityParams{
		UserID:      req.requesterId,
		EntityID:    entityID,
		IsFollowing: isFollowing,
		IsMember:    isMember,
	})
	if err != nil {
		return false, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	return canSee, nil
}
