"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function getPendingRequestsCount({ groupId }) {
    try {
        const url = `/groups/${groupId}/pending-count`
        const response = await serverApiRequest(url, {
            method: "GET",
            forwardCookies: true
        });

        if (!response.ok) {
            return {success: false, status: response.status, error: response.message};
        }

        return { success: true, data: response.data };

    } catch (error) {
        console.error("Error fetching user groups: ", error);
        return { success: false, error: error.message };
    }
}