"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function deleteNotification(notificationId) {
    try {
        const url = `/notifications/${notificationId}`;
        const response = await serverApiRequest(url, {
            method: "DELETE",
            forwardCookies: true
        });

        if (!response.ok) {
            return {success: false, status: response.status, error: response.message};
        }

        return { success: true, data: response.data };
    } catch (error) {
        console.error("Error deleting notification:", error);
        return { success: false, error: error.message };
    }
}
