"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function markAsRead({convID , lastMsgID}) {
    try {
        const apiResp = await serverApiRequest("/my/chat/read", {
            method: "POST",
            body: JSON.stringify({
                conversation_id: convID,
                last_read_message_id: lastMsgID
            }),
            forwardCookies: true,
            headers: {
                "Content-Type": "application/json"
            }
        });

        if (!apiResp.ok) {
            return {success: false, status: apiResp.status, error: apiResp.message};
        }

        return { success: true, data: apiResp.data };

    } catch (error) {
        console.error("Mark as read error: ", error);
        return { success: false, error: error.message };
    }
}