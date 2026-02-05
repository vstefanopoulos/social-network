"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function getMessages({ interlocutorId, boundary = null, limit = 50, retrieveUsers = false, getPrevious = true }) {
    try {
        let url = `/my/chat/${interlocutorId}?limit=${limit}`;

        if (boundary) {
            url += `&boundary=${boundary}`;
        }
        if (retrieveUsers) {
            url += `&retrieve-users=${retrieveUsers}`;
        }
        if (!getPrevious) {
            url += `&get-previous=${getPrevious}`;
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
        console.error("Error fetching messages:", error);
        return { success: false, error: error.message };
    }
}
