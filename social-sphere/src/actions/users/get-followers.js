"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function getFollowers({ userId, limit = 100, offset = 0 } = {}) {
    try {
        if (!userId) {
            console.error("User ID is required to fetch followers");
            return { success: false, error: "User ID is required" };
        }
        const url = `/users/${userId}/followers?limit=${limit}&offset=${offset}`;
        const response = await serverApiRequest(url, {
            method: "GET"
        });

        if (!response.ok) {
            return {success: false, status: response.status, error: response.message};
        }

        return { success: true, data: response.data };

    } catch (error) {
        console.error("Error fetching followers:", error);
        return { success: false, error: error.message };
    }
}
