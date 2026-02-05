package application

import (
	"context"
	"errors"
	"fmt"
	ds "social-network/services/users/internal/db/dbservice"
	ce "social-network/shared/go/commonerrors"
	ct "social-network/shared/go/ct"
	"social-network/shared/go/models"
	tele "social-network/shared/go/telemetry"
	"time"

	"social-network/shared/gen-go/media"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// not responsible for image url fetching
func (s *Application) GetBasicUserInfo(ctx context.Context, userId ct.Id) (resp models.User, err error) {
	input := fmt.Sprintf("%#v", userId)

	if err := userId.Validate(); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ce.New(ce.ErrInvalidArgument, err, input).WithPublic("user does not exist or is deleted")
		}
		return models.User{}, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	row, err := s.db.GetUserBasic(ctx, userId.Int64())
	if err != nil {
		return models.User{}, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	u := models.User{
		UserId:   ct.Id(userId),
		Username: ct.Username(row.Username),
		AvatarId: ct.Id(row.AvatarID),
	}
	return u, nil

}

// image url fetching happens in retrieve users
func (s *Application) GetBatchBasicUserInfo(ctx context.Context, userIds ct.Ids) ([]models.User, error) {
	input := fmt.Sprintf("%#v", userIds)

	if err := userIds.Validate(); err != nil {
		return nil, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	rows, err := s.db.GetBatchUsersBasic(ctx, userIds.Int64())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ce.New(ce.ErrInvalidArgument, err, input).WithPublic("no users found")
		}
		return nil, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}

	users := make([]models.User, 0, len(rows))
	for _, r := range rows {
		users = append(users, models.User{
			UserId:   ct.Id(r.ID),
			Username: ct.Username(r.Username),
			AvatarId: ct.Id(r.AvatarID),
		})
	}
	return users, nil
}

func (s *Application) GetUserProfile(ctx context.Context, req models.UserProfileRequest) (models.UserProfileResponse, error) {
	input := fmt.Sprintf("%#v", req)

	var profile models.UserProfileResponse
	if err := ct.ValidateStruct(req); err != nil {
		return profile, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	row, err := s.db.GetUserProfile(ctx, req.UserId.Int64())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.UserProfileResponse{}, ce.New(ce.ErrNotFound, err, input).
				WithPublic("profile not found")
		}
		return models.UserProfileResponse{}, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	dob := time.Time{}
	if row.DateOfBirth.Valid {
		dob = row.DateOfBirth.Time
	}

	profile = models.UserProfileResponse{
		UserId:      ct.Id(row.ID),
		Username:    ct.Username(row.Username),
		FirstName:   ct.Name(row.FirstName),
		LastName:    ct.Name(row.LastName),
		DateOfBirth: ct.DateOfBirth(dob),
		AvatarId:    ct.Id(row.AvatarID),
		About:       ct.About(row.AboutMe),
		Public:      row.ProfilePublic,
		CreatedAt:   ct.GenDateTime(row.CreatedAt.Time),
		Email:       ct.Email(row.Email),
	}

	followingParams := models.FollowUserReq{
		FollowerId:   req.RequesterId,
		TargetUserId: req.UserId,
	}

	profile.ViewerIsFollowing, err = s.IsFollowing(ctx, followingParams)
	if err != nil {
		return models.UserProfileResponse{}, ce.Wrap(nil, err)
	}

	profile.IsPending, err = s.isFollowRequestPending(ctx, followingParams)
	if err != nil {
		return models.UserProfileResponse{}, ce.Wrap(nil, err)
	}

	//now check if the profile owner has a pending follow request towards the viewer
	reverseFollowingParams := models.FollowUserReq{
		FollowerId:   req.UserId,
		TargetUserId: req.RequesterId,
	}

	profile.FollowRequestFromProfileOwner, err = s.isFollowRequestPending(ctx, reverseFollowingParams)
	if err != nil {
		return models.UserProfileResponse{}, ce.Wrap(nil, err)
	}

	if req.RequesterId == req.UserId {
		profile.OwnProfile = true
	}

	// if !profile.Public && !profile.ViewerIsFollowing && !profile.OwnProfile {
	// 	return models.UserProfileResponse{}, ErrProfilePrivate
	// }

	profile.FollowersCount, err = s.db.GetFollowerCount(ctx, req.UserId.Int64())
	if err != nil {
		return models.UserProfileResponse{}, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	profile.FollowingCount, err = s.db.GetFollowingCount(ctx, req.UserId.Int64())
	if err != nil {
		return models.UserProfileResponse{}, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}

	groupsRow, err := s.db.UserGroupCountsPerRole(ctx, req.UserId.Int64())
	if err != nil {
		return models.UserProfileResponse{}, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}

	profile.GroupsCount = groupsRow.TotalMemberships //owner and member, can change to member only
	profile.OwnedGroupsCount = groupsRow.OwnerCount

	if profile.AvatarId > 0 {
		imageUrl, err := s.mediaRetriever.GetImage(ctx, profile.AvatarId.Int64(), media.FileVariant(1))
		if err != nil {
			tele.Error(ctx, "media retriever failed for @1", "request", profile.AvatarId, "error", err.Error()) //log error instead of returning
			s.removeFailedImage(ctx, err, profile.AvatarId.Int64())
		} else {

			profile.AvatarURL = imageUrl
		}
	}

	return profile, nil

}

func (s *Application) SearchUsers(ctx context.Context, req models.UserSearchReq) ([]models.User, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return []models.User{}, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	rows, err := s.db.SearchUsers(ctx, ds.SearchUsersParams{
		Query: req.SearchTerm.String(),
		Limit: req.Limit.Int32(),
	})

	if err != nil {
		return nil, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}

	if len(rows) == 0 {
		return []models.User{}, nil
	}

	users := make([]models.User, 0, len(rows))
	var imageIds ct.Ids
	for _, r := range rows {
		users = append(users, models.User{
			UserId:   ct.Id(r.ID),
			Username: ct.Username(r.Username),
			AvatarId: ct.Id(r.AvatarID),
		})
		if r.AvatarID > 0 {
			imageIds = append(imageIds, ct.Id(r.AvatarID))
		}
	}

	//get avatar urls
	if len(imageIds) > 0 {
		avatarMap, failedImageIds, err := s.mediaRetriever.GetImages(ctx, imageIds, media.FileVariant_THUMBNAIL)
		if err != nil {
			tele.Error(ctx, "media retriever failed for @1", "request", imageIds, "error", err.Error()) //log error instead of returning

		} else {
			for i := range users {
				users[i].AvatarURL = avatarMap[users[i].AvatarId.Int64()]
			}
			s.removeFailedImagesAsync(ctx, failedImageIds)
		}
	}

	return users, nil
}

func (s *Application) UpdateUserProfile(ctx context.Context, req models.UpdateProfileRequest) (models.UserProfileResponse, error) {
	//NOTE front needs to send everything, not just changed fields
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return models.UserProfileResponse{}, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	dob := pgtype.Date{
		Time:  req.DateOfBirth.Time(),
		Valid: true,
	}

	avatarId := req.AvatarId.Int64()
	if req.DeleteImage {
		avatarId = 0
	}

	row, err := s.db.UpdateUserProfile(ctx, ds.UpdateUserProfileParams{
		ID:          req.UserId.Int64(),
		Username:    req.Username.String(),
		FirstName:   req.FirstName.String(),
		LastName:    req.LastName.String(),
		DateOfBirth: dob,
		AvatarID:    avatarId,
		AboutMe:     req.About.String(),
	})
	if err != nil {
		return models.UserProfileResponse{}, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}

	//delete redis basic user info if exists

	key, err := ct.BasicUserInfoKey{Id: req.UserId}.GenKey()
	_ = s.clients.Del(ctx, key)
	if err != nil {
		tele.Warn(ctx, "could not delete basic user info for user @1 using key @2", "userId", req.UserId, "key", key)
	}

	newDob := time.Time{}
	if row.DateOfBirth.Valid {
		newDob = row.DateOfBirth.Time
	}

	profile := models.UserProfileResponse{
		UserId:      req.UserId,
		Username:    ct.Username(row.Username),
		FirstName:   ct.Name(row.FirstName),
		LastName:    ct.Name(row.LastName),
		DateOfBirth: ct.DateOfBirth(newDob),
		AvatarId:    ct.Id(row.AvatarID),
		About:       ct.About(row.AboutMe),
		Public:      row.ProfilePublic,
	}

	return profile, nil

}

func (s *Application) UpdateProfilePrivacy(ctx context.Context, req models.UpdateProfilePrivacyRequest) error {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	err := s.db.UpdateProfilePrivacy(ctx, ds.UpdateProfilePrivacyParams{
		ID:            req.UserId.Int64(),
		ProfilePublic: req.Public,
	})
	if err != nil {
		return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}

	return nil
}

func (s *Application) RemoveImages(ctx context.Context, failedImages []int64) error {
	//input := fmt.Sprintf("%#v", failedImages)

	err := s.db.RemoveImages(ctx, failedImages)
	if err != nil {
		tele.Warn(ctx, "images @1 could not be deleted", "imageIds", failedImages)

	}

	return nil
}
