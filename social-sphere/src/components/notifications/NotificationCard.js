"use client";

import { useState } from "react";
import Link from "next/link";
import { Bell, Trash2, Check, X, User, Users, Calendar, Heart, MessageCircle } from "lucide-react";
import { markNotificationAsRead } from "@/actions/notifs/mark-as-read";
import { deleteNotification } from "@/actions/notifs/delete-notification";
import { getRelativeTime } from "@/lib/time";
import { constructNotif } from "@/lib/notifications";
import { useStore } from '@/store/store';
import { useRouter } from "next/navigation";

export default function NotificationCard({ notification, onDelete, onUpdate }) {
    const [isActing, setIsActing] = useState(false);
    const [isDeleting, setIsDeleting] = useState(false);
    const [acted, setActed] = useState(notification.acted);
    const decrementNotifs = useStore((state) => state.decrementNotifs);
    const router = useRouter();

    const needsAction = notification.needs_action && !acted;

    const getNotificationIcon = () => {
        switch (notification.type) {
            case "follow_request":
            case "new_follower":
            case "follow_request_accepted":
                return <User className="w-4 h-4" />;
            case "group_invite":
            case "group_join_request":
            case "group_join_request_accepted":
            case "group_invite_accepted":
                return <Users className="w-4 h-4" />;
            case "new_event":
                return <Calendar className="w-4 h-4" />;
            case "post_reply":
                return <MessageCircle className="w-4 h-4" />;
            case "like":
                return <Heart className="w-4 h-4" />;
            default:
                return <Bell className="w-4 h-4" />;
        }
    };

    const handleAction = async (content, accept) => {
        setIsActing(true);
        try {
            console.log(`calling ${content.callback} with value ${accept}`)
            const result = await content.callback(accept);
            console.log("result from acting", result)
            if (result?.success) {

                console.log("Updating optimistically")
                // optimistically update acted state
                setActed(true);

                // mark as read
                await markNotificationAsRead(notification.id);
                decrementNotifs();
                onUpdate?.(notification.id, { acted: true, seen: true });
            }
        } catch (error) {
            console.error("Error handling notification action:", error);
        } finally {
            setIsActing(false);
        }
    };

    const handleDelete = async () => {
        setIsDeleting(true);
        try {
            const result = await deleteNotification(notification.id);
            if (result.success) {
                decrementNotifs();
                onDelete?.(notification.id);
            }
        } catch (error) {
            console.error("Error deleting notification:", error);
        } finally {
            setIsDeleting(false);
        }
    };

    const content = constructNotif(notification, true);

    return (
        <div className={`group bg-background border border-(--border) rounded-xl p-4 transition-all hover:border-(--muted)/40 hover:shadow-sm ${notification.seen ? "opacity-40" : ""}`}>
            <div className="flex items-start gap-3">
                {/* Icon */}
                <div className="shrink-0 w-10 h-10 bg-(--accent)/10 rounded-full flex items-center justify-center text-(--accent)">
                    {getNotificationIcon()}
                </div>

                {/* Content */}
                <div className="flex-1 min-w-0">
                    <div className="text-sm text-foreground leading-snug">
                        {content?.who && (
                            <button
                                onClick={async (e) => {
                                    e.preventDefault();
                                    e.stopPropagation();
                                    if (!notification.seen) {
                                        await markNotificationAsRead(notification.id);
                                        decrementNotifs();
                                    }
                                    router.push(`/profile/${content.whoID}`);
                                }}
                                className="text-sm text-(--accent) hover:text-(--accent-hover) hover:underline cursor-pointer"
                            >
                                {content.who}
                            </button>
                        )}
                        <span>{content?.message}</span>
                        {content?.wherePost && (
                            <button
                                onClick={async (e) => {
                                    e.preventDefault();
                                    e.stopPropagation();
                                    if (!notification.seen) {
                                        await markNotificationAsRead(notification.id);
                                        decrementNotifs();
                                    }
                                    router.push(`/posts/${content.whereID}`);
                                }}
                                className="text-sm text-(--accent) hover:text-(--accent-hover) hover:underline cursor-pointer"
                            >
                                {content.wherePost}
                            </button>
                        )}
                        {content?.whereGroup && (
                            <button
                                onClick={async (e) => {
                                    e.preventDefault();
                                    e.stopPropagation();
                                    if (!notification.seen) {
                                        await markNotificationAsRead(notification.id);
                                        decrementNotifs();
                                    }
                                    router.push(`/groups/${content.whereID}`);
                                }}
                                className="text-sm text-(--accent) hover:text-(--accent-hover) hover:underline cursor-pointer"
                            >
                                {content.whereGroup}
                            </button>
                        )}
                        {content?.whereEvent && (
                            <button
                                onClick={async (e) => {
                                    e.preventDefault();
                                    e.stopPropagation();
                                    if (!notification.seen) {
                                        await markNotificationAsRead(notification.id);
                                        decrementNotifs();
                                    }
                                    router.push(`/groups/${content.whereID}?t=events`);
                                }}
                                className="text-sm text-(--accent) hover:text-(--accent-hover) hover:underline cursor-pointer"
                            >
                                {content.whereEvent}
                            </button>
                        )}
                        {content?.whereUser && (
                            <Link
                                href={`/profile/${content.whereID}`}
                                className="font-semibold text-(--accent) hover:underline"
                            >
                                {content.whereUser}
                            </Link>
                        )}
                    </div>

                    <p className="text-xs text-(--muted) mt-1">
                        {getRelativeTime(notification.created_at)}
                    </p>

                    {/* Action Buttons for actionable notifications */}
                    {needsAction && (
                        <div className="flex items-center gap-2 mt-3">
                            <button
                                onClick={() => handleAction(content, true)}
                                disabled={isActing}
                                className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium bg-(--accent) text-white hover:bg-(--accent-hover) rounded-full transition-colors disabled:opacity-50 cursor-pointer"
                            >
                                <Check className="w-3 h-3" />
                                Accept
                            </button>
                            <button
                                onClick={() => handleAction(content, false)}
                                disabled={isActing}
                                className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium border border-(--border) text-foreground hover:bg-(--muted)/10 rounded-full transition-colors disabled:opacity-50 cursor-pointer"
                            >
                                <X className="w-3 h-3" />
                                Decline
                            </button>
                        </div>
                    )}
                </div>

                {/* Delete Button */}
                <button
                    onClick={handleDelete}
                    disabled={isDeleting}
                    className="shrink-0 p-2 text-(--muted) hover:text-red-500 hover:bg-red-500/5 rounded-full transition-colors opacity-0 group-hover:opacity-100 disabled:opacity-50 cursor-pointer"
                >
                    <Trash2 className="w-4 h-4" />
                </button>
            </div>
        </div>
    );
}
