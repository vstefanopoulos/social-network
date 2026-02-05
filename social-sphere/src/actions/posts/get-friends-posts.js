"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function getFriendsPosts({ limit = 10, offset = 0 } = {}) {
    try {
        const url = `/posts/friends?limit=${limit}&offset=${offset}`;
        const response = await serverApiRequest(url, {
            method: "GET",
            forwardCookies: true
        });

        if (!response.ok) {
            return {success: false, status: response.status, error: response.message};
        }

        return { success: true, data: response.data };

    } catch (error) {
        console.error("Error fetching friends posts:", error);
        return { success: false, error: error.message };
    }
}
