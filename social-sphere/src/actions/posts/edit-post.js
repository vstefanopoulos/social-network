"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function editPost(postData) {
    try {
        const url = `/posts/${postData.post_id}`;
        const response = await serverApiRequest(url, {
            method: "POST",
            body: JSON.stringify(postData),
            headers: {
                "Content-Type": "application/json"
            }
        });

        if (!response.ok) {
            return {success: false, status: response.status, error: response.message};
        }

        return { success: true, data: response.data };

    } catch (error) {
        console.error("Edit Post Action Error:", error);
        return { success: false, error: error.message };
    }
}
