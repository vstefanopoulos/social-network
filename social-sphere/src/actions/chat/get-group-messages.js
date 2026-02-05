"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function getGroupMessages({ groupId, boundary = null, limit = 50, getPrevious = true }) {
    try {
        let url = `/groups/${groupId}/chat?group_id=${groupId}&limit=${limit}`;

        if (boundary) {
            url += `&boundary=${boundary}`;
        }
        if (!getPrevious) {
            url += `&get_previous=${getPrevious}`;
        }

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
        console.error("Error fetching group messages:", error);
        return { success: false, error: error.message };
    }
}
