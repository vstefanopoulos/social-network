"use client";

import { useState } from "react";
import { Bell, CheckCheck } from "lucide-react";
import NotificationCard from "./NotificationCard";
import { markAllNotificationsAsRead } from "@/actions/notifs/mark-all-as-read";
import { getNotifs } from "@/actions/notifs/get-user-notifs";
import { useStore } from '@/store/store';

export default function NotificationsContent({ initialNotifications }) {
    const [notifications, setNotifications] = useState(initialNotifications || []);
    const [isMarkingAll, setIsMarkingAll] = useState(false);
    const [isLoadingMore, setIsLoadingMore] = useState(false);
    const [hasMore, setHasMore] = useState(initialNotifications?.length >= 20);
    const setUnreadNotifs = useStore((state) => state.setUnreadNotifs);

    const handleDelete = (notificationId) => {
        setNotifications((prev) => prev.filter((n) => n.id !== notificationId));
    };

    const handleUpdate = (notificationId, updates) => {
        setNotifications((prev) =>
            prev.map((n) => (n.id === notificationId ? { ...n, ...updates } : n))
        );
    };

    const handleMarkAllAsRead = async () => {
        setIsMarkingAll(true);
        try {
            const result = await markAllNotificationsAsRead();
            console.log("resowdjncjkqenvfjn", result)
            if (result.success) {
                // Optionally update UI to reflect all read
                setNotifications((prev) =>
                    prev.map((notif) => ({ ...notif, seen: true }))
                )
                // update unreadNotif state
                setUnreadNotifs(0)
            }
        } catch (error) {
            console.error("Error marking all as read:", error);
        } finally {
            setIsMarkingAll(false);
        }
    };

    const handleLoadMore = async () => {
        setIsLoadingMore(true);
        try {
            const moreNotifications = await getNotifs({
                limit: 20,
                offset: notifications.length
            });
            if (moreNotifications.success && moreNotifications.data?.length > 0) {
                setNotifications((prev) => [...prev, ...moreNotifications.data]);
                setHasMore(moreNotifications.data.length >= 20);
            } else {
                setHasMore(false);
            }
        } catch (error) {
            console.error("Error loading more notifications:", error);
        } finally {
            setIsLoadingMore(false);
        }
    };

    return (
        <div className="max-w-2xl mx-auto px-4 py-8">
            {/* Header */}
            <div className="flex items-center justify-between mb-6">
                <h1 className="text-2xl font-bold text-foreground">Notifications</h1>
                {notifications.length > 0 && (
                    <button
                        onClick={handleMarkAllAsRead}
                        disabled={isMarkingAll}
                        className="flex items-center gap-2 px-4 py-2 text-sm font-medium text-(--muted) hover:text-foreground hover:bg-(--muted)/10 rounded-full transition-colors disabled:opacity-50 cursor-pointer"
                    >
                        <CheckCheck className="w-4 h-4" />
                        Mark all as read
                    </button>
                )}
            </div>

            {/* Notifications List */}
            {notifications.length > 0 ? (
                <div className="space-y-3">
                    {notifications.map((notification) => (
                        <NotificationCard
                            key={notification.id}
                            notification={notification}
                            onDelete={handleDelete}
                            onUpdate={handleUpdate}
                        />
                    ))}

                    {/* Load More */}
                    {hasMore && (
                        <div className="pt-4 text-center">
                            <button
                                onClick={handleLoadMore}
                                disabled={isLoadingMore}
                                className="px-6 py-2 text-sm font-medium text-(--accent) hover:bg-(--accent)/10 rounded-full transition-colors disabled:opacity-50 cursor-pointer"
                            >
                                {isLoadingMore ? "Loading..." : "Load more"}
                            </button>
                        </div>
                    )}
                </div>
            ) : (
                /* Empty State */
                <div className="flex flex-col items-center justify-center py-16 text-center">
                    <div className="w-16 h-16 bg-(--muted)/10 rounded-full flex items-center justify-center mb-4">
                        <Bell className="w-8 h-8 text-(--muted)" />
                    </div>
                    <h2 className="text-lg font-semibold text-foreground mb-2">
                        No notifications yet
                    </h2>
                    <p className="text-sm text-(--muted) max-w-sm">
                        When you get notifications, they will show up here.
                    </p>
                </div>
            )}
        </div>
    );
}
