"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function unfollowUser(userId) {
    try {
        const url = `/users/${userId}/unfollow`;
        const response = await serverApiRequest(url, {
            method: "POST",
            forwardCookies: true
        });

        if (!response.ok) {
            return {success: false, status: response.status, error: response.message};
        }

        return { success: true, data: response.data };
    } catch (error) {
        console.error("Error unfollowing user:", error);
        return { success: false, error: error.message };
    }
}
