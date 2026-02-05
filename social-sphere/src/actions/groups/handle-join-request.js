"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function handleJoinRequest({ groupId, requesterId, accepted }) {
    try {
        const url = `/groups/${groupId}/join-response`;
        const response = await serverApiRequest(url, {
            method: "POST",
            body: JSON.stringify({
                requester_id: requesterId,
                accepted: accepted
            }),
            headers: {
                "Content-Type": "application/json"
            }
        });

        if (!response.ok) {
            return {success: false, status: response.status, error: response.message};
        }

        return { success: true, data: response.data };
    } catch (error) {
        console.error("Error handling join request:", error);
        return { success: false, error: error.message };
    }
}
