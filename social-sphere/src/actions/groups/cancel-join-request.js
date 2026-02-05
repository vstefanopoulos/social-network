"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function cancelJoinRequest({ groupId }) {
    try {
        const url = `/groups/${groupId}/cancel-join-request`;
        const response = await serverApiRequest(url, {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            }
        });

        if (!response.ok) {
            return {success: false, status: response.status, error: response.message};
        }

        return { success: true, data: response };
    } catch (error) {
        console.error("Error canceling join request:", error);
        return { success: false, error: error.message };
    }
}
