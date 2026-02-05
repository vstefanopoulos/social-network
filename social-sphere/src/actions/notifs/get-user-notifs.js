"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function getNotifs({ limit = 10, offset = 0 } = {}) {
    try {
        const url = `/notifications?limit=${limit}&offset=${offset}&read_only=false`;
        const response = await serverApiRequest(url, {
            method: "GET",
            forwardCookies: true
        });

        if (!response.ok) {
            return {success: false, status: response.status, error: response.message};
        }

        return { success: true, data: response.data };

    } catch (error) {
        console.error("Error fetching notifications:", error);
        return { success: false, error: error.message };
    }
}
