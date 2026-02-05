/*
Expose methods via gRpc
*/

package handler

import (
	"context"
	cm "social-network/shared/gen-go/common"
	pb "social-network/shared/gen-go/posts"
	ce "social-network/shared/go/commonerrors"
	ct "social-network/shared/go/ct"
	"social-network/shared/go/models"
	tele "social-network/shared/go/telemetry"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// POSTS

func (s *PostsHandler) GetPostById(ctx context.Context, req *pb.GenericReq) (*pb.Post, error) {
	tele.Info(ctx, "GetPostById gRPC method called. @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	post, err := s.Application.GetPostById(ctx, models.GenericReq{
		RequesterId: ct.Id(req.RequesterId),
		EntityId:    ct.Id(req.EntityId),
	})
	if err != nil {
		tele.Error(ctx, "Error in GetPostById. @1 @2", "request", req.String(), "error", err.Error())
		return nil, ce.EncodeProto(err)
	}

	selectedUsers := make([]*cm.User, 0, len(post.SelectedAudienceUsers))
	for _, u := range post.SelectedAudienceUsers {
		selectedUsers = append(selectedUsers, &cm.User{
			UserId:    u.UserId.Int64(),
			Username:  u.Username.String(),
			Avatar:    u.AvatarId.Int64(),
			AvatarUrl: u.AvatarURL,
		})
	}

	return &pb.Post{
		PostId:   int64(post.PostId),
		PostBody: string(post.Body),

		User: &cm.User{
			UserId:    post.User.UserId.Int64(),
			Username:  post.User.Username.String(),
			Avatar:    post.User.AvatarId.Int64(),
			AvatarUrl: post.User.AvatarURL,
		},
		GroupId:         int64(post.GroupId),
		Audience:        post.Audience.String(),
		CommentsCount:   int32(post.CommentsCount),
		ReactionsCount:  int32(post.ReactionsCount),
		LastCommentedAt: post.LastCommentedAt.ToProto(),
		CreatedAt:       post.CreatedAt.ToProto(),
		UpdatedAt:       post.UpdatedAt.ToProto(),
		LikedByUser:     post.LikedByUser,
		ImageId:         int64(post.ImageId),
		ImageUrl:        post.ImageUrl,
		SelectedAudienceUsers: &cm.ListUsers{
			Users: selectedUsers,
		},
	}, nil
}

func (s *PostsHandler) CreatePost(ctx context.Context, req *pb.CreatePostReq) (*pb.IdResp, error) {
	tele.Info(ctx, "CreatePost gRPC method called. @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	postId, err := s.Application.CreatePost(ctx, models.CreatePostReq{
		CreatorId:   ct.Id(req.CreatorId),
		Body:        ct.PostBody(req.Body),
		GroupId:     ct.Id(req.GroupId),
		Audience:    ct.Audience(req.Audience),
		AudienceIds: ct.FromInt64s(req.AudienceIds.Values),
		ImageId:     ct.Id(req.ImageId),
	})
	if err != nil {
		tele.Error(ctx, "Error in CreatePost. @1 @2", "request", req.String(), "error", err.Error())
		return nil, ce.EncodeProto(err)
	}
	return &pb.IdResp{Id: postId}, nil
}

func (s *PostsHandler) DeletePost(ctx context.Context, req *pb.GenericReq) (*emptypb.Empty, error) {
	tele.Info(ctx, "DeletePost gRPC method called. @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	err := s.Application.DeletePost(ctx, models.GenericReq{
		RequesterId: ct.Id(req.RequesterId),
		EntityId:    ct.Id(req.EntityId),
	})
	if err != nil {
		tele.Error(ctx, "Error in DeletePost. @1 @2", "request", req.String(), "error", err.Error())
		return nil, ce.EncodeProto(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *PostsHandler) EditPost(ctx context.Context, req *pb.EditPostReq) (*emptypb.Empty, error) {
	tele.Info(ctx, "EditPost gRPC method called. @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	err := s.Application.EditPost(ctx, models.EditPostReq{
		RequesterId: ct.Id(req.RequesterId),
		PostId:      ct.Id(req.PostId),
		NewBody:     ct.PostBody(req.Body),
		ImageId:     ct.Id(req.ImageId),
		Audience:    ct.Audience(req.Audience),
		AudienceIds: ct.FromInt64s(req.AudienceIds.Values),
		DeleteImage: req.GetDeleteImage(),
	})
	if err != nil {
		tele.Error(ctx, "Error in EditPost. @1 @2", "request", req.String(), "error", err.Error())
		return nil, ce.EncodeProto(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *PostsHandler) GetMostPopularPostInGroup(ctx context.Context, req *pb.SimpleIdReq) (*pb.Post, error) {
	tele.Info(ctx, "GetMostPopularPostInGroup gRPC method called. @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	post, err := s.Application.GetMostPopularPostInGroup(ctx, models.SimpleIdReq{
		Id: ct.Id(req.Id),
	})
	if err != nil {
		tele.Error(ctx, "Error in GetMostPopularPostInGroup. @1 @2", "request", req.String(), "error", err.Error())
		return nil, ce.EncodeProto(err)
	}

	return &pb.Post{
		PostId:   int64(post.PostId),
		PostBody: string(post.Body),
		User: &cm.User{
			UserId:    post.User.UserId.Int64(),
			Username:  post.User.Username.String(),
			Avatar:    post.User.AvatarId.Int64(),
			AvatarUrl: post.User.AvatarURL,
		},
		GroupId:         int64(post.GroupId),
		Audience:        post.Audience.String(),
		CommentsCount:   int32(post.CommentsCount),
		ReactionsCount:  int32(post.ReactionsCount),
		LastCommentedAt: post.LastCommentedAt.ToProto(),
		CreatedAt:       post.CreatedAt.ToProto(),
		UpdatedAt:       post.UpdatedAt.ToProto(),
		LikedByUser:     post.LikedByUser,
		ImageId:         int64(post.ImageId),
		ImageUrl:        post.ImageUrl,
	}, nil
}

func (s *PostsHandler) GetPersonalizedFeed(ctx context.Context, req *pb.GetPersonalizedFeedReq) (*pb.ListPosts, error) {
	tele.Info(ctx, "GetPersonalizedFeed gRPC method called. @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	posts, err := s.Application.GetPersonalizedFeed(ctx, models.GetPersonalizedFeedReq{
		RequesterId: ct.Id(req.RequesterId),
		Limit:       ct.Limit(req.Limit),
		Offset:      ct.Offset(req.Offset),
	})
	if err != nil {
		tele.Error(ctx, "Error in GetPersonalizedFeed. @1 @2", "request", req.String(), "error", err.Error())
		return nil, ce.EncodeProto(err)
	}
	pbPosts := make([]*pb.Post, 0, len(posts))
	for _, p := range posts {
		pbPosts = append(pbPosts, &pb.Post{
			PostId:   int64(p.PostId),
			PostBody: string(p.Body),
			User: &cm.User{
				UserId:    p.User.UserId.Int64(),
				Username:  p.User.Username.String(),
				Avatar:    p.User.AvatarId.Int64(),
				AvatarUrl: p.User.AvatarURL,
			},
			GroupId:         int64(p.GroupId),
			Audience:        p.Audience.String(),
			CommentsCount:   int32(p.CommentsCount),
			ReactionsCount:  int32(p.ReactionsCount),
			LastCommentedAt: p.LastCommentedAt.ToProto(),
			CreatedAt:       p.CreatedAt.ToProto(),
			UpdatedAt:       p.UpdatedAt.ToProto(),
			LikedByUser:     p.LikedByUser,
			ImageId:         int64(p.ImageId),
			ImageUrl:        p.ImageUrl,
		})
	}
	return &pb.ListPosts{Posts: pbPosts}, nil
}

func (s *PostsHandler) GetPublicFeed(ctx context.Context, req *pb.GenericPaginatedReq) (*pb.ListPosts, error) {
	tele.Info(ctx, "GetPublicFeed gRPC method called. @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	posts, err := s.Application.GetPublicFeed(ctx, models.GenericPaginatedReq{
		RequesterId: ct.Id(req.RequesterId),
		Limit:       ct.Limit(req.Limit),
		Offset:      ct.Offset(req.Offset),
	})
	if err != nil {
		tele.Error(ctx, "Error in GetPublicFeed @1 @2", "request", req.String(), "error", err.Error())
		return nil, ce.EncodeProto(err)
	}
	pbPosts := make([]*pb.Post, 0, len(posts))
	for _, p := range posts {
		pbPosts = append(pbPosts, &pb.Post{
			PostId:   int64(p.PostId),
			PostBody: string(p.Body),
			User: &cm.User{
				UserId:    p.User.UserId.Int64(),
				Username:  p.User.Username.String(),
				Avatar:    p.User.AvatarId.Int64(),
				AvatarUrl: p.User.AvatarURL,
			},
			GroupId:         int64(p.GroupId),
			Audience:        p.Audience.String(),
			CommentsCount:   int32(p.CommentsCount),
			ReactionsCount:  int32(p.ReactionsCount),
			LastCommentedAt: p.LastCommentedAt.ToProto(),
			CreatedAt:       p.CreatedAt.ToProto(),
			UpdatedAt:       p.UpdatedAt.ToProto(),
			LikedByUser:     p.LikedByUser,
			ImageId:         int64(p.ImageId),
			ImageUrl:        p.ImageUrl,
		})
	}
	return &pb.ListPosts{Posts: pbPosts}, nil
}

func (s *PostsHandler) GetUserPostsPaginated(ctx context.Context, req *pb.GetUserPostsReq) (*pb.ListPosts, error) {
	tele.Info(ctx, "GetUserPostsPaginated gRPC method called. @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	posts, err := s.Application.GetUserPostsPaginated(ctx, models.GetUserPostsReq{
		CreatorId:   ct.Id(req.CreatorId),
		RequesterId: ct.Id(req.RequesterId),
		Limit:       ct.Limit(req.Limit),
		Offset:      ct.Offset(req.Offset),
	})
	if err != nil {
		tele.Error(ctx, "Error in GetUserPostsPaginated @1 @2", "request", req.String(), "error", err.Error())
		return nil, ce.EncodeProto(err)
	}
	pbPosts := make([]*pb.Post, 0, len(posts))
	for _, p := range posts {
		pbPosts = append(pbPosts, &pb.Post{
			PostId:   int64(p.PostId),
			PostBody: string(p.Body),
			User: &cm.User{
				UserId:    p.User.UserId.Int64(),
				Username:  p.User.Username.String(),
				Avatar:    p.User.AvatarId.Int64(),
				AvatarUrl: p.User.AvatarURL,
			},
			GroupId:         int64(p.GroupId),
			Audience:        p.Audience.String(),
			CommentsCount:   int32(p.CommentsCount),
			ReactionsCount:  int32(p.ReactionsCount),
			LastCommentedAt: p.LastCommentedAt.ToProto(),
			CreatedAt:       p.CreatedAt.ToProto(),
			UpdatedAt:       p.UpdatedAt.ToProto(),
			LikedByUser:     p.LikedByUser,
			ImageId:         int64(p.ImageId),
			ImageUrl:        p.ImageUrl,
		})
	}
	return &pb.ListPosts{Posts: pbPosts}, nil
}

func (s *PostsHandler) GetGroupPostsPaginated(ctx context.Context, req *pb.GetGroupPostsReq) (*pb.ListPosts, error) {
	tele.Info(ctx, "GetGroupPostsPaginated gRPC method called. @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	posts, err := s.Application.GetGroupPostsPaginated(ctx, models.GetGroupPostsReq{
		GroupId:     ct.Id(req.GroupId),
		RequesterId: ct.Id(req.RequesterId),
		Limit:       ct.Limit(req.Limit),
		Offset:      ct.Offset(req.Offset),
	})
	if err != nil {
		tele.Error(ctx, "Error in GetGroupPostsPaginated @1 @2", "request", req.String(), "error", err.Error())
		return nil, ce.EncodeProto(err)
	}
	pbPosts := make([]*pb.Post, 0, len(posts))
	for _, p := range posts {
		pbPosts = append(pbPosts, &pb.Post{
			PostId:   int64(p.PostId),
			PostBody: string(p.Body),
			User: &cm.User{
				UserId:    p.User.UserId.Int64(),
				Username:  p.User.Username.String(),
				Avatar:    p.User.AvatarId.Int64(),
				AvatarUrl: p.User.AvatarURL,
			},
			GroupId:         int64(p.GroupId),
			Audience:        p.Audience.String(),
			CommentsCount:   int32(p.CommentsCount),
			ReactionsCount:  int32(p.ReactionsCount),
			LastCommentedAt: p.LastCommentedAt.ToProto(),
			CreatedAt:       p.CreatedAt.ToProto(),
			UpdatedAt:       p.UpdatedAt.ToProto(),
			LikedByUser:     p.LikedByUser,
			ImageId:         int64(p.ImageId),
			ImageUrl:        p.ImageUrl,
		})
	}
	return &pb.ListPosts{Posts: pbPosts}, nil
}

func (s *PostsHandler) CreateComment(ctx context.Context, req *pb.CreateCommentReq) (*pb.IdResp, error) {
	tele.Info(ctx, "CreateComment gRPC method called. @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	commentId, err := s.Application.CreateComment(ctx, models.CreateCommentReq{
		CreatorId: ct.Id(req.CreatorId),
		ParentId:  ct.Id(req.ParentId),
		Body:      ct.CommentBody(req.Body),
		ImageId:   ct.Id(req.ImageId),
	})
	if err != nil {
		tele.Error(ctx, "Error in CreateComment @1 @2", "request", req.String(), "error", err.Error())
		return nil, ce.EncodeProto(err)
	}
	return &pb.IdResp{Id: commentId}, nil
}

func (s *PostsHandler) EditComment(ctx context.Context, req *pb.EditCommentReq) (*emptypb.Empty, error) {
	tele.Info(ctx, "EditComment gRPC method called. @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	err := s.Application.EditComment(ctx, models.EditCommentReq{
		CreatorId:   ct.Id(req.CreatorId),
		CommentId:   ct.Id(req.CommentId),
		Body:        ct.CommentBody(req.Body),
		ImageId:     ct.Id(req.ImageId),
		DeleteImage: req.GetDeleteImage(),
	})
	if err != nil {
		tele.Error(ctx, "Error in EditComment @1 @2", "request", req.String(), "error", err.Error())
		return nil, ce.EncodeProto(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *PostsHandler) DeleteComment(ctx context.Context, req *pb.GenericReq) (*emptypb.Empty, error) {
	tele.Info(ctx, "DeleteComment gRPC method called. @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	err := s.Application.DeleteComment(ctx, models.GenericReq{
		RequesterId: ct.Id(req.RequesterId),
		EntityId:    ct.Id(req.EntityId),
	})
	if err != nil {
		tele.Error(ctx, "Error in DeleteComment @1 @2", "request", req.String(), "error", err.Error())
		return nil, ce.EncodeProto(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *PostsHandler) GetCommentsByParentId(ctx context.Context, req *pb.EntityIdPaginatedReq) (*pb.ListComments, error) {
	tele.Info(ctx, "GetCommentsByParentId gRPC method called. @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	comments, err := s.Application.GetCommentsByParentId(ctx, models.EntityIdPaginatedReq{
		RequesterId: ct.Id(req.RequesterId),
		EntityId:    ct.Id(req.EntityId),
		Limit:       ct.Limit(req.Limit),
		Offset:      ct.Offset(req.Offset),
	})
	if err != nil {
		tele.Error(ctx, "Error in GetCommentsByParentId @1 @2", "request", req.String(), "error", err.Error())
		return nil, ce.EncodeProto(err)
	}
	pbComments := make([]*pb.Comment, 0, len(comments))
	for _, c := range comments {
		pbComments = append(pbComments, &pb.Comment{
			CommentId: int64(c.CommentId),
			ParentId:  int64(c.ParentId),
			Body:      string(c.Body),
			User: &cm.User{
				UserId:    c.User.UserId.Int64(),
				Username:  c.User.Username.String(),
				Avatar:    c.User.AvatarId.Int64(),
				AvatarUrl: c.User.AvatarURL,
			},
			ReactionsCount: int32(c.ReactionsCount),
			CreatedAt:      c.CreatedAt.ToProto(),
			UpdatedAt:      c.UpdatedAt.ToProto(),
			LikedByUser:    c.LikedByUser,
			ImageId:        int64(c.ImageId),
			ImageUrl:       c.ImageUrl,
		})
	}
	return &pb.ListComments{Comments: pbComments}, nil
}

func (s *PostsHandler) CreateEvent(ctx context.Context, req *pb.CreateEventReq) (*pb.IdResp, error) {
	tele.Info(ctx, "CreateEvent gRPC method called. @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	eventId, err := s.Application.CreateEvent(ctx, models.CreateEventReq{
		Title:     ct.Title(req.Title),
		Body:      ct.EventBody(req.Body),
		CreatorId: ct.Id(req.CreatorId),
		GroupId:   ct.Id(req.GroupId),
		ImageId:   ct.Id(req.ImageId),
		EventDate: ct.EventDateTime(req.EventDate.AsTime()),
	})
	if err != nil {
		tele.Error(ctx, "Error in CreateEvent @1 @2", "request", req.String(), "error", err.Error())
		return nil, ce.EncodeProto(err)
	}
	return &pb.IdResp{Id: eventId}, nil
}

func (s *PostsHandler) DeleteEvent(ctx context.Context, req *pb.GenericReq) (*emptypb.Empty, error) {
	tele.Info(ctx, "DeleteEvent gRPC method called. @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	err := s.Application.DeleteEvent(ctx, models.GenericReq{
		RequesterId: ct.Id(req.RequesterId),
		EntityId:    ct.Id(req.EntityId),
	})
	if err != nil {
		tele.Error(ctx, "Error in DeleteEvent @1 @2", "request", req.String(), "error", err.Error())
		return nil, ce.EncodeProto(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *PostsHandler) EditEvent(ctx context.Context, req *pb.EditEventReq) (*emptypb.Empty, error) {
	tele.Info(ctx, "EditEvent gRPC method called. @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	err := s.Application.EditEvent(ctx, models.EditEventReq{
		EventId:     ct.Id(req.EventId),
		RequesterId: ct.Id(req.RequesterId),
		Title:       ct.Title(req.Title),
		Body:        ct.EventBody(req.Body),
		Image:       ct.Id(req.ImageId),
		EventDate:   ct.EventDateTime(req.EventDate.AsTime()),
		DeleteImage: req.GetDeleteImage(),
	})
	if err != nil {
		tele.Error(ctx, "Error in EditEvent @1 @2", "request", req.String(), "error", err.Error())
		return nil, ce.EncodeProto(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *PostsHandler) GetEventsByGroupId(ctx context.Context, req *pb.EntityIdPaginatedReq) (*pb.ListEvents, error) {
	tele.Info(ctx, "GetEventsByGroupId gRPC method called. @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	events, err := s.Application.GetEventsByGroupId(ctx, models.EntityIdPaginatedReq{
		RequesterId: ct.Id(req.RequesterId),
		EntityId:    ct.Id(req.EntityId),
		Limit:       ct.Limit(req.Limit),
		Offset:      ct.Offset(req.Offset),
	})
	if err != nil {
		tele.Error(ctx, "Error in GetEventsByGroupId @1 @2", "request", req.String(), "error", err.Error())
		return nil, ce.EncodeProto(err)
	}
	pbEvents := make([]*pb.Event, 0, len(events))
	for _, e := range events {
		var ur *wrapperspb.BoolValue
		if e.UserResponse != nil {
			ur = wrapperspb.Bool(*e.UserResponse)
		}

		pbEvents = append(pbEvents, &pb.Event{
			EventId: int64(e.EventId),
			Title:   string(e.Title),
			Body:    string(e.Body),
			User: &cm.User{
				UserId:    e.User.UserId.Int64(),
				Username:  e.User.Username.String(),
				Avatar:    e.User.AvatarId.Int64(),
				AvatarUrl: e.User.AvatarURL,
			},
			GroupId:       int64(e.GroupId),
			EventDate:     e.EventDate.ToProto(),
			GoingCount:    int32(e.GoingCount),
			NotGoingCount: int32(e.NotGoingCount),
			ImageId:       int64(e.ImageId),
			ImageUrl:      e.ImageUrl,
			CreatedAt:     e.CreatedAt.ToProto(),
			UpdatedAt:     e.UpdatedAt.ToProto(),
			UserResponse:  ur,
		})
	}
	return &pb.ListEvents{Events: pbEvents}, nil
}

func (s *PostsHandler) RespondToEvent(ctx context.Context, req *pb.RespondToEventReq) (*emptypb.Empty, error) {
	tele.Info(ctx, "RespondToEvent gRPC method called. @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	err := s.Application.RespondToEvent(ctx, models.RespondToEventReq{
		EventId:     ct.Id(req.EventId),
		ResponderId: ct.Id(req.ResponderId),
		Going:       req.Going,
	})
	if err != nil {
		tele.Error(ctx, "Error in RespondToEvent @1 @2", "request", req.String(), "error", err.Error())
		return nil, ce.EncodeProto(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *PostsHandler) RemoveEventResponse(ctx context.Context, req *pb.GenericReq) (*emptypb.Empty, error) {
	tele.Info(ctx, "RemoveEventResponse gRPC method called. @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	err := s.Application.RemoveEventResponse(ctx, models.GenericReq{
		RequesterId: ct.Id(req.RequesterId),
		EntityId:    ct.Id(req.EntityId),
	})
	if err != nil {
		tele.Error(ctx, "Error in RemoveEventResponse @1 @2", "request", req.String(), "error", err.Error())
		return nil, ce.EncodeProto(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *PostsHandler) SuggestUsersByPostActivity(ctx context.Context, req *pb.SimpleIdReq) (*cm.ListUsers, error) {
	tele.Info(ctx, "SuggestUsersByPostActivity gRPC method called. @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	users, err := s.Application.SuggestUsersByPostActivity(ctx, models.SimpleIdReq{
		Id: ct.Id(req.Id),
	})
	if err != nil {
		tele.Error(ctx, "Error in SuggestUsersByPostActivity @1 @2", "request", req.String(), "error", err.Error())
		return nil, ce.EncodeProto(err)
	}
	pbUsers := make([]*cm.User, 0, len(users))
	for _, u := range users {
		pbUsers = append(pbUsers, &cm.User{
			UserId:    u.UserId.Int64(),
			Username:  u.Username.String(),
			Avatar:    u.AvatarId.Int64(),
			AvatarUrl: u.AvatarURL,
		})
	}
	return &cm.ListUsers{Users: pbUsers}, nil
}

func (s *PostsHandler) ToggleOrInsertReaction(ctx context.Context, req *pb.GenericReq) (*emptypb.Empty, error) {
	tele.Info(ctx, "ToggleOrInsertReaction gRPC method called. @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	err := s.Application.ToggleOrInsertReaction(ctx, models.GenericReq{
		RequesterId: ct.Id(req.RequesterId),
		EntityId:    ct.Id(req.EntityId),
	})
	if err != nil {
		tele.Error(ctx, "Error in ToggleOrInsertReaction @1 @2", "request", req.String(), "error", err.Error())
		return nil, ce.EncodeProto(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *PostsHandler) GetWhoLikedEntityId(ctx context.Context, req *pb.GenericReq) (*cm.ListUsers, error) {
	tele.Info(ctx, "GetWhoLikedEntityId gRPC method called. @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	users, err := s.Application.GetWhoLikedEntityId(ctx, models.GenericReq{
		RequesterId: ct.Id(req.RequesterId),
		EntityId:    ct.Id(req.EntityId),
	})
	if err != nil {
		tele.Error(ctx, "Error in GetWhoLikedEntityId @1 @2", "request", req.String(), "error", err.Error())
		return nil, ce.EncodeProto(err)
	}
	pbUsers := make([]*cm.User, 0, len(users))
	for _, u := range users {
		pbUsers = append(pbUsers, &cm.User{
			UserId:    u.UserId.Int64(),
			Username:  u.Username.String(),
			Avatar:    u.AvatarId.Int64(),
			AvatarUrl: u.AvatarURL,
		})
	}
	return &cm.ListUsers{Users: pbUsers}, nil
}

func (s *PostsHandler) GetPostAudienceForComment(ctx context.Context, req *pb.SimpleIdReq) (*pb.AudienceResp, error) {
	tele.Info(ctx, "GetPostAudienceForComment gRPC method called. @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	audience, err := s.Application.GetPostAudienceForComment(ctx, req.Id)
	if err != nil {
		tele.Error(ctx, "Error in GetPostAudienceForComment @1 @2", "request", req.String(), "error", err.Error())
		return nil, ce.EncodeProto(err)
	}
	return &pb.AudienceResp{
		Audience: audience,
	}, nil
}
