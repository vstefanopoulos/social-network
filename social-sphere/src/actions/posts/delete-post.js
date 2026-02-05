"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function deletePost(postId) {
    try {
        const url = `/posts/${postId}`;
        const response = await serverApiRequest(url, {
            method: "DELETE",
            headers: {
                "Content-Type": "application/json"
            }
        });

        if (!response.ok) {
            return {success: false, status: response.status, error: response.message};
        }

        return { success: true, data: response.data };

    } catch (error) {
        console.error("Delete Post Action Error:", error);
        return { success: false, error: error.message };
    }
}
