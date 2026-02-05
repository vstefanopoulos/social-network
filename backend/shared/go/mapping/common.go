package mapping

import (
	commonpb "social-network/shared/gen-go/common"
	ct "social-network/shared/go/ct"
	md "social-network/shared/go/models"
)

func MapUserToProto(u md.User) *commonpb.User {
	return &commonpb.User{
		UserId:    u.UserId.Int64(),
		Username:  u.Username.String(),
		Avatar:    u.AvatarId.Int64(),
		AvatarUrl: u.AvatarURL,
	}
}

func MapUserFromProto(u *commonpb.User) md.User {
	if u == nil {
		return md.User{}
	}

	return md.User{
		UserId:    ct.Id(u.UserId),
		Username:  ct.Username(u.Username),
		AvatarId:  ct.Id(u.Avatar),
		AvatarURL: u.AvatarUrl,
	}
}
