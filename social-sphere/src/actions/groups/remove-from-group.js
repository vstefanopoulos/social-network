"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function removeFromGroup({ groupId, memberId }) {
    try {
        const url = `/groups/${groupId}/remove-member`;
        const response = await serverApiRequest(url, {
            method: "POST",
            body: JSON.stringify({
                member_id: memberId,
            }),
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
        console.error("Error removing member from group:", error);
        return { success: false, error: error.message };
    }
}
