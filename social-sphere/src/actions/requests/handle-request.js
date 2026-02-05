"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function handleFollowRequest({ requesterId, accept }) {
    try {
        const url = `/users/${requesterId}/follow-response`
        const response = await serverApiRequest(url, {
            method: "POST",
            body: JSON.stringify({
                accept: accept,
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
        console.error("Error handling follow request:", error);
        return { success: false, error: error.message };
    }
}
