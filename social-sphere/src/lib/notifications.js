import { handleFollowRequest } from '@/actions/requests/handle-request';
import { respondToGroupInvite } from '@/actions/groups/respond-to-invite';
import { handleJoinRequest } from '@/actions/groups/handle-join-request';

export function constructNotif(notif, aggregate = false) {

    // FOLLOWERS 
    if (notif.type === "new_follower") {
        return {
            who: notif.payload.follower_name,
            whoID: notif.payload.follower_id,
            message: " is now following you"
        };
    }

    if (notif.type === "follow_request") {
        if (aggregate) {
            if (notif.acted) {
                return {
                    message: "You are now friends with ",
                    whereUser: notif.payload.requester_name,
                    whereID: notif.payload.requester_id
                };
            }
        }
        return {
            who: notif.payload.requester_name,
            whoID: notif.payload.requester_id,
            message: " wants to follow you",
            action: notif.needs_action,
            callback: async (response) => {
                return await handleFollowRequest({ requesterId: notif.payload.requester_id, accept: response })
            }
        };
    }

    if (notif.type === "follow_request_accepted") {
        return {
            who: notif.payload.target_name,
            whoID: notif.payload.target_id,
            message: " accepted your follow request"
        };
    }

    // POSTS 
    if (notif.type === "post_reply") {
        if (aggregate) {
            if (notif.count > 1) {
                return {
                    message: `${notif.count} people have commented on your post: `,
                    wherePost: notif.payload.post_content,
                    whereID: notif.payload.post_id,
                };
            }
        }

        return {
            who: notif.payload.commenter_name,
            whoID: notif.payload.commenter_id,
            message: " commented on your post: ",
            wherePost: notif.payload.post_content,
            whereID: notif.payload.post_id,
        };
    }

    if (notif.type === "like") {
        if (aggregate) {
            if (notif.count > 1) {
                return {
                    message: `${notif.count} people have liked your `,
                    wherePost: "post",
                    whereID: notif.payload.post_id,
                };
            }
        }

        return {
            who: notif.payload.liker_name,
            whoID: notif.payload.liker_id,
            message: " liked your ",
            wherePost: "post",
            whereID: notif.payload.post_id
        };
    }

    // GROUPS
    if (notif.type === "group_invite") {

        if (aggregate) {
            if (notif.acted) {
                return {
                    message: "You accepted to join group: ",
                    where: notif.payload.requester_name,
                    whereID: notif.payload.requester_id,
                    whereGroup: notif.payload.group_name,
                    whereID: notif.payload.group_id,
                };
            }
        }
        return {
            who: notif.payload.inviter_name,
            whoID: notif.payload.inviter_id,
            message: " invited you to join group: ",
            whereGroup: notif.payload.group_name,
            whereID: notif.payload.group_id,
            action: notif.needs_action,
            callback: async (response) => {
                return await respondToGroupInvite({ groupId: notif.payload.group_id, accept: response })
            }
        };
    }

    if (notif.type === "group_join_request") {
        if (aggregate) {
            if (notif.acted) {
                return {
                    who: notif.payload.requester_name,
                    whoID: notif.payload.requester_id,
                    message: " is now a member of your group: ",
                    whereGroup: notif.payload.group_name,
                    whereID: notif.payload.group_id,
                };
            }
        }

        return {
            who: notif.payload.requester_name,
            whoID: notif.payload.requester_id,
            message: " wants to join your group: ",
            whereGroup: notif.payload.group_name,
            whereID: notif.payload.group_id,
            action: notif.needs_action,
            callback: async (response) => {
                return await handleJoinRequest({ groupId: notif.payload.group_id, requesterId: notif.payload.requester_id, accepted: response })
            }
        };
    }

    if (notif.type === "group_join_request_accepted") {
        return {
            message: "You were accepted to join group: ",
            whereGroup: notif.payload.group_name,
            whereID: notif.payload.group_id
        };
    }

    if (notif.type === "group_invite_accepted") {
        return {
            who: notif.payload.invited_name,
            whoID: notif.payload.invited_id,
            message: " accepted your invitation to join group: ",
            whereGroup: notif.payload.group_name,
            whereID: notif.payload.group_id
        };
    }

    if (notif.type === "new_event") {
        return {
            whoEvent: notif.payload.event_title,
            whoID: notif.payload.group_id,
            message: " event was created in group: ",
            whereEvent: notif.payload.group_name,
            whereID: notif.payload.group_id
        }
    }
}
