package handlers

import (
	"net/http"
	"social-network/shared/gen-go/chat"
	"social-network/shared/gen-go/media"
	"social-network/shared/gen-go/notifications"
	"social-network/shared/gen-go/posts"
	"social-network/shared/gen-go/users"
	middleware "social-network/shared/go/http-middleware"
	"social-network/shared/go/ratelimit"
	"social-network/shared/go/retrieveusers"
)

type Handlers struct {
	CacheService CacheService
	UsersService users.UserServiceClient
	PostsService posts.PostsServiceClient
	ChatService  chat.ChatServiceClient
	MediaService media.MediaServiceClient
	NotifService notifications.NotificationServiceClient
	RetriveUsers *retrieveusers.UserRetriever
}

func NewHandlers(
	serviceName string,
	CacheService CacheService,
	UsersService users.UserServiceClient,
	PostsService posts.PostsServiceClient,
	ChatService chat.ChatServiceClient,
	MediaService media.MediaServiceClient,
	NotifService notifications.NotificationServiceClient,
	RetrieveUsers *retrieveusers.UserRetriever,
) *http.ServeMux {
	handlers := Handlers{
		CacheService: CacheService,
		UsersService: UsersService,
		PostsService: PostsService,
		ChatService:  ChatService,
		MediaService: MediaService,
		NotifService: NotifService,
		RetriveUsers: RetrieveUsers,
	}
	return handlers.BuildMux(serviceName)
}

//TODO remove endpoint from chain, and find another way

// BuildMux builds and returns the HTTP request multiplexer with all routes and middleware applied
func (h *Handlers) BuildMux(serviceName string) *http.ServeMux {
	mux := http.NewServeMux()
	ratelimiter := ratelimit.NewRateLimiter(serviceName+":", h.CacheService)
	middlewareObj := middleware.NewMiddleware(ratelimiter, serviceName, mux)
	SetEndpoint := middlewareObj.SetEndpoint

	IP := middleware.IPLimit
	USERID := middleware.UserLimit

	// mux.
	// 	andleFunc("/test/{yo}/hello").
	// 		AllowedMethod("GET").
	// 		RateLimit(IP, 20, 5).
	// 		EnrichContext().
	// 		Finalize(h.testHandler())

	SetEndpoint("/login").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		EnrichContext().
		Finalize(h.loginHandler())

	SetEndpoint("/register").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		EnrichContext().
		Finalize(h.registerHandler())

		//fileid url, --DONE

	SetEndpoint("/files/{file_id}/validate").
		AllowedMethod("POST").
		RateLimit(IP, 5, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 5, 5).
		Finalize(h.validateFileUpload())

		//image id variant from url --DONE

	SetEndpoint("/files/images/{image_id}/{variant}").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 5, 5).
		Finalize(h.getImageUrl())

	SetEndpoint("/logout").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.logoutHandler())

		//START GROUPS ======================================
		//START GROUPS ======================================
		//START GROUPS ======================================

	SetEndpoint("/groups").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.createGroup())

		//TODO make params --DONE

	SetEndpoint("/groups").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.getAllGroupsPaginated())

		//TODO cut url, and json --DONE

	SetEndpoint("/groups/{group_id}").
		AllowedMethod("POST").
		RateLimit(IP, 5, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 5, 5).
		Finalize(h.updateGroup())

		//TODO slice url --DONE

	SetEndpoint("/groups/{group_id}").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.getGroupInfo())

		//TODO group id from url --DONE

	SetEndpoint("/groups/{group_id}/popular-post").
		AllowedMethod("GET").
		RateLimit(IP, 5, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 5, 5).
		Finalize(h.getMostPopularPostInGroup())

		//TODO get params, and groupid from url --DONE

	SetEndpoint("/groups/{group_id}/members").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.getGroupMembers())

		//TODO group id from url --DONE

	SetEndpoint("/groups/{group_id}/join-response").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.handleGroupJoinRequest())

		//TODO group id from url --DONE

	SetEndpoint("/groups/{group_id}/invite").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.inviteToGroup())

		//TODO get group id from url --DONE

	SetEndpoint("/groups/{group_id}/leave").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.leaveGroup())

		//TODO get group id from url --DONE

	SetEndpoint("/groups/{group_id}/remove-member").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.removeFromGroup())

		//TODO group id url --DONE

	SetEndpoint("/groups/{group_id}/join-request").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.requestJoinGroup())

		//TODO group id url --DONE

	SetEndpoint("/groups/{group_id}/cancel-join-request").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.cancelGroupJoinRequest())

		//groupid from url --DONE

	SetEndpoint("/groups/{group_id}/invite-response").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.respondToGroupInvite())

		//todo params and groupid from url --DONE

	SetEndpoint("/groups/{group_id}/pending-requests").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.getPendingGroupJoinRequests())

		//TODO group id, params --DONE

	SetEndpoint("/groups/{group_id}/pending-count").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.getPendingGroupJoinRequestsCount())

		//group id from url, params --DONE

	SetEndpoint("/groups/{group_id}/invitable-followers").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.GetFollowersNotInvitedToGroup())

		//params, groupid url --DONE

	SetEndpoint("/groups/search").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.searchGroups())

		//params, groupid url --DONE

	SetEndpoint("/groups/{group_id}/posts").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.getGroupPostsPaginated())

		//TODO get params --DONE

	SetEndpoint("/my/groups").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.getUserGroupsPaginated())

		//params, groupid url --DONE

	SetEndpoint("/groups/{group_id}/events").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.getEventsByGroupId())

		//groupid param --DONE

	SetEndpoint("/groups/{group_id}/events").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.createEvent())

		// USERS ======================================
		// USERS ======================================
		// USERS ======================================

		//  url --DONE

	SetEndpoint("/users/{user_id}/follow").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.followUser())

		//params, userid url --DONE

	SetEndpoint("/users/{user_id}/followers").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.getFollowersPaginated())

		//TODO currently unused in front

	SetEndpoint("/my/follow-suggestions").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.getFollowSuggestions())

		//TODO not used by front yet
		//params, userid url --DONE

	SetEndpoint("/users/{user_id}/following").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.getFollowingPaginated())

		//get user id url --DONE

	SetEndpoint("/users/{user_id}/follow-response").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.handleFollowRequest())

		//params --DONE

	SetEndpoint("/users/search").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.searchUsers())

		//userid url --DONE

	SetEndpoint("/users/{user_id}/unfollow").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.unFollowUser())

		//user id --DONE

	SetEndpoint("/users/{user_id}/profile").
		AllowedMethod("GET").
		RateLimit(IP, 40, 20).
		Auth().
		EnrichContext().
		RateLimit(USERID, 40, 20).
		Finalize(h.getUserProfile())

		//TODO params, user_id url --DONE

	SetEndpoint("/users/{user_id}/retrieve").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.retrieveUser())

	SetEndpoint("/users/retrieve").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.retrieveUsers())

	SetEndpoint("/users/{user_id}/posts").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.getUserPostsPaginated())

		// MY =====================
		// MY =====================
		// MY =====================
		// MY =====================

		//done

	SetEndpoint("/my/profile/privacy").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.updateProfilePrivacy())

	SetEndpoint("/my/profile/email").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.updateUserEmail())

	SetEndpoint("/my/profile/password").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.updateUserPassword())

	SetEndpoint("/my/profile").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.updateUserProfile())

		// POST =====================
		// POST =====================
		// POST =====================
		// POST =====================

		//TODO params --DONE

	SetEndpoint("/posts/public").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.getPublicFeed())

		//TODO params --DONE

	SetEndpoint("/posts/friends").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.getPersonalizedFeed())

		//params, postsid url --DONE

	SetEndpoint("/posts/{post_id}").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.getPostById())

		//done

	SetEndpoint("/posts").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.createPost())

		//post id url --DONE

	SetEndpoint("/posts/{post_id}").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.editPost())

		//post id url

	SetEndpoint("/posts/{post_id}").
		AllowedMethod("DELETE").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.deletePost())

		// COMMENTS ===================
		// COMMENTS ===================
		// COMMENTS ===================

	SetEndpoint("/comments").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.createComment())

		//todo commnetid url --DONE

	SetEndpoint("/comments/{comment_id}").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.editComment())

		//commentid ulr --DONE

	SetEndpoint("/comments/{comment_id}").
		AllowedMethod("DELETE").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.deleteComment())

		//params --DONE

	SetEndpoint("/comments").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.getCommentsByParentId())

		//EVENTS ===========================
		//EVENTS ===========================
		//EVENTS ===========================
		//EVENTS ===========================

		//  eventid url --DONE

	SetEndpoint("/events/{event_id}").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.editEvent())

		// eventid url --DONE

	SetEndpoint("/events/{event_id}").
		AllowedMethod("DELETE").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.deleteEvent())

		// --DONE

	SetEndpoint("/events/{event_id}/response").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.respondToEvent())

		// --DONE

	SetEndpoint("/events/{event_id}/response").
		AllowedMethod("DELETE").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.RemoveEventResponse())

		// --DONE

	SetEndpoint("/reactions").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.toggleOrInsertReaction())

	SetEndpoint("/reactions/{entity_id}").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.getWhoLikedEntityId())

		// NOTIFICATIONS =====================
		// NOTIFICATIONS =====================
		// NOTIFICATIONS =====================
		// NOTIFICATIONS =====================

		//TODO remove notification type

		//params //add unread parameter// and only read// maybe read type? --DONE

	SetEndpoint("/notifications").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.GetUserNotifications())

	SetEndpoint("/notifications-count").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.GetUnreadNotificationsCount())

		//--DONE

	SetEndpoint("/notifications/mark-all").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.MarkAllAsRead())

	SetEndpoint("/notifications/mark").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.MarkNotificationAsRead())

		//params

	SetEndpoint("/notifications/{notification_id}").
		AllowedMethod("DELETE").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.DeleteNotification())

		// CHAT ============================================
		// CHAT ============================================
		// CHAT ============================================
		// CHAT ============================================
		//--DONE

	SetEndpoint("/my/chat/{interlocutor_id}").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.CreatePrivateMsg())

	//--DONE

	SetEndpoint("/my/chat/{interlocutor_id}").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.GetPrivateMessagesPag())

	//conv id url, params --DONE

	SetEndpoint("/my/chat/get-unread-conversation-count").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.GetConvsWithUnreadsCount())

	SetEndpoint("/my/chat/{conversation_id}/preview").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.GetPrivateConversationById())

		//params

	SetEndpoint("/my/chat/previews").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.GetPrivateConversations())

		//group id url

	SetEndpoint("/groups/{group_id}/chat").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.CreateGroupMsg())

	SetEndpoint("/groups/{group_id}/chat").
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.GetGroupMessagesPag())

	SetEndpoint("/my/chat/read").
		AllowedMethod("POST").
		RateLimit(IP, 20, 5).
		Auth().
		EnrichContext().
		RateLimit(USERID, 20, 5).
		Finalize(h.UpdateLastRead())

	return mux
}
