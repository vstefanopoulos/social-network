"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function leaveGroup({ groupId }) {
    try {
        const url = `/groups/${groupId}/leave`;
        const response = await serverApiRequest(url, {
            method: "POST",
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
        console.error("Error leaving group:", error);
        return { success: false, error: error.message };
    }
}
