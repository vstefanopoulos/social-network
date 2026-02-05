"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function searchGroups({ query, limit = 10, offset = 0 }) {
    try {
        const url = `/groups/search?query=${query}&limit=${limit}&offset=${offset}`;
        const response = await serverApiRequest(url, {
            method: "GET",
            forwardCookies: true
        });

        if (!response.ok) {
            return {success: false, status: response.status, error: response.message};
        }

        return { success: true, data: response.data };

    } catch (error) {
        console.error("Error searching groups:", error);
        return { success: false, error: error.message };
    }
}
