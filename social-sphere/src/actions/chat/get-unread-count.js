"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function getUnreadCount() {
    try {
        const url = `/my/chat/get-unread-conversation-count`

        const response = await serverApiRequest(url, {
            method: "GET",
            forwardCookies: true,
            headers: {
                "Content-Type": "application/json"
            }
        });

        if (!response.ok) {
            return {success: false, status: response.status, error: response.message};
        }

        return { success: true, data: response.data };

    } catch (error) {
        console.error("Error fetching unread count:", error);
        return { success: false, error: error.message };
    }
}
