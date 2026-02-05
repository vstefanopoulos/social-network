"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function sendGroupMsg({ groupId, msg }) {
    try {
        const url = `/groups/${groupId}/chat`;
        const apiResp = await serverApiRequest(url, {
            method: "POST",
            body: JSON.stringify({
                group_id: groupId,
                message_body: msg
            }),
            headers: {
                "Content-Type": "application/json"
            }
        });

        if (!apiResp.ok) {
            return {success: false, status: apiResp.status, error: apiResp.message};
        }

        return { success: true, data: apiResp.data };

    } catch (error) {
        console.error("Send Group Message Error:", error);
        return { success: false, error: error.message };
    }
}
