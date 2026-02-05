"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function getPendingRequests({ groupId, limit = 20, offset = 0 }) {
    try {
        const url = `/groups/${groupId}/pending-requests?limit=${limit}&offset=${offset}`;
        const response = await serverApiRequest(url, {
            method: "GET"
        });

        if (!response.ok) {
            return {success: false, status: response.status, error: response.message};
        }

        return { success: true, data: response.data };
    } catch (error) {
        console.error("Error fetching pending requests:", error);
        return { success: false, error: error.message };
    }
}
