package retrieveusers

import (
	"context"
	"fmt"

	cm "social-network/shared/gen-go/common"
	"social-network/shared/gen-go/media"
	ce "social-network/shared/go/commonerrors"
	ct "social-network/shared/go/ct"
	"social-network/shared/go/models"
	tele "social-network/shared/go/telemetry"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

// GetUsers returns a map[userID]User, using cache + batch RPC.
func (h *UserRetriever) GetUsers(ctx context.Context, userIds ct.Ids) (map[ct.Id]models.User, error) {
	input := fmt.Sprintf("user retriever: get users: uses ids: %v", userIds)
	//========================== STEP 1 : get user info from users ===============================================
	if len(userIds) == 0 {
		tele.Warn(ctx, "get users called with empty ids slice")
		return nil, nil
	}

	if err := userIds.Validate(); err != nil {
		return nil, ce.New(ce.ErrInvalidArgument, err, input)
	}

	tele.Debug(ctx, "get users called with ids @1", "ids", userIds)

	ids := userIds.Unique()

	users := make(map[ct.Id]models.User, len(ids))

	var missing ct.Ids

	// Cache lookup
	for _, id := range ids {
		//  Check local
		u, ok := h.GetFromLocal(ctx, id)
		if ok {
			users[id] = u
			continue
		}

		// Check Redis
		u, err := h.GetFromRedis(ctx, id)
		if err != nil {
			missing = append(missing, id)
			continue
		}

		h.SetToLocal(ctx, u)
		users[id] = u
	}

	// Early return if all found in cache
	if len(missing) == 0 {
		return users, nil
	}

	// Batch RPC for missing users

	var imageIds ct.Ids
	var usersWithAvatarsToFetch ct.Ids

	resp, err := h.client.GetBatchBasicUserInfo(ctx, &cm.UserIds{Values: missing.Int64()})
	if err != nil {
		return nil, ce.DecodeProto(err, input)
	}

	for _, u := range resp.Users {
		user := models.User{
			UserId:   ct.Id(u.UserId),
			Username: ct.Username(u.Username),
			AvatarId: ct.Id(u.Avatar),
		}
		users[user.UserId] = user

		if user.AvatarId > 0 { //exclude 0 imageIds
			imageIds = append(imageIds, user.AvatarId)
			usersWithAvatarsToFetch = append(usersWithAvatarsToFetch, user.UserId)
		}
	}

	//========================== STEP 2 : get avatars from media ===============================================
	if len(imageIds) > 0 {
		// Use shared MediaRetriever for images (handles caching and fetching)
		imageMap, imagesToDelete, err := h.mediaRetriever.GetImages(ctx, imageIds, media.FileVariant_THUMBNAIL)
		if err != nil {
			tele.Error(ctx, "media retriever failed for @1", "request", imageIds, "error", err.Error()) //log error instead of returning
		} else {
			for _, id := range usersWithAvatarsToFetch {
				u := users[id]
				if url, ok := imageMap[u.AvatarId.Int64()]; ok {
					u.AvatarURL = url
					users[id] = u
				}
			}
		}

		if len(imagesToDelete) > 0 {
			go h.removeFailedImages(ctx, imagesToDelete)
		}
	}

	// Add missing to caches
	for _, id := range missing {
		user := users[id]

		// Redis
		h.AddToRedis(ctx, user)

		// Local
		h.SetToLocal(ctx, user)
	}

	return users, nil
}

func (h *UserRetriever) GetUser(ctx context.Context, userId ct.Id) (models.User, error) {
	input := fmt.Sprintf("user retriever: get user: id: %v", userId)

	if err := userId.Validate(); err != nil {
		return models.User{}, ce.New(ce.ErrInvalidArgument, err, input)
	}

	tele.Debug(ctx, "retrieve user called with user id @1", "userId", userId)

	// Local cache lookup
	u, ok := h.GetFromLocal(ctx, userId)
	if ok {
		return u, nil
	}

	if user, err := h.GetFromRedis(ctx, userId); err == nil {
		h.SetToLocal(ctx, user)
		return user, nil
	}

	//========================== STEP 1 : get user info from users ===============================================

	resp, err := h.client.GetBasicUserInfo(ctx, wrapperspb.Int64(userId.Int64()))
	if err != nil {
		return models.User{}, ce.DecodeProto(err, input)
	}

	user := models.User{
		UserId:   ct.Id(resp.UserId),
		Username: ct.Username(resp.Username),
		AvatarId: ct.Id(resp.Avatar),
	}

	//========================== STEP 2 : get avatar from media ===============================================
	// Get image url for users

	if user.AvatarId > 0 { //exclude 0 imageIds

		// Use shared MediaRetriever for images (handles caching and fetching)
		imageUrl, err := h.mediaRetriever.GetImage(ctx, user.AvatarId.Int64(), media.FileVariant_THUMBNAIL)
		if err != nil {
			if err.IsClass(ce.ErrNotFound) {
				go h.removeFailedImages(ctx, []int64{user.AvatarId.Int64()})
			}
		} else {
			user.AvatarURL = imageUrl
		}

	}

	h.AddToRedis(ctx, user)

	h.SetToLocal(ctx, user)

	return user, nil
}
