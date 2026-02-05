"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function markNotificationAsRead(notificationId) {
    try {
        const url = `/notifications/mark?notification_id=${notificationId}`;
        await serverApiRequest(url, {
            method: "POST",
            forwardCookies: true
        });
        return { success: true };
    } catch (error) {
        console.error("Error marking notification as read:", error);
        return { success: false, error: error.message };
    }
}
