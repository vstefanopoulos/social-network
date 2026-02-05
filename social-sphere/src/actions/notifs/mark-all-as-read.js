"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function markAllNotificationsAsRead() {
    try {
        const url = `/notifications/mark-all`;
        await serverApiRequest(url, {
            method: "POST",
            forwardCookies: true
        });
        return { success: true };
    } catch (error) {
        console.error("Error marking all notifications as read:", error);
        return { success: false, error: error.message };
    }
}
