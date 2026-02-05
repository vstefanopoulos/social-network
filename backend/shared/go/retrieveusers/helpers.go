package retrieveusers

import (
	"context"
	userpb "social-network/shared/gen-go/users"
	ct "social-network/shared/go/ct"
	"social-network/shared/go/models"
	tele "social-network/shared/go/telemetry"
	"time"
)

// False means either cache is nil or user is not logged.
// True but empty user means user does not exist.
func (h *UserRetriever) GetFromLocal(ctx context.Context, userId ct.Id) (models.User, bool) {
	if h.localCache != nil {
		u, ok := h.localCache.Get(userId)
		if u != nil {
			tele.Debug(ctx, "found user on local cache", "username", u.Username)
			return *u, true
		} else {
			return models.User{}, ok
		}
	}
	return models.User{}, false
}

func (h *UserRetriever) SetToLocal(ctx context.Context, user models.User) {
	if h.localCache != nil {
		h.localCache.SetWithTTL(user.UserId, &user, 1, time.Duration(10*time.Second))
	}
}

func (h *UserRetriever) GetFromRedis(ctx context.Context, id ct.Id) (models.User, error) {
	var u models.User
	// Check redis
	key, err := ct.BasicUserInfoKey{Id: id}.GenKey()
	if err != nil {
		tele.Warn(ctx, "failed to construct redis key for id @1: @2", "userId", id, "error", err.Error())
		return u, err
	}

	if err := h.cache.GetObj(ctx, key, &u); err != nil {
		return u, err
	}

	tele.Info(ctx, "found user on redis: @1", "user", u)
	return u, nil
}

func (h *UserRetriever) AddToRedis(ctx context.Context, user models.User) error {
	key, err := ct.BasicUserInfoKey{Id: user.UserId}.GenKey()
	if err == nil {
		_ = h.cache.SetObj(ctx,
			key,
			user,
			h.ttl,
		)
		tele.Debug(ctx, "user set on redis: @1 with key @2", "user", user, "key", key)
	} else {
		tele.Warn(ctx, "failed to construct redis key for user @1: @2", "userId", user.UserId, "error", err.Error())
		return err
	}
	return nil
}

func (h *UserRetriever) removeFailedImages(ctx context.Context, imgIds []int64) {
	req := &userpb.FailedImageIds{
		ImgIds: imgIds,
	}

	tele.Info(ctx, "removing avatar ids @1 from users", "failedImageIds", imgIds)
	_, err := h.client.RemoveImages(context.WithoutCancel(ctx), req)
	if err != nil {
		tele.Warn(ctx, "failed  to delete failed images @1 from users: @2", "failedImageIds", imgIds, "error", err.Error())
	}
}
