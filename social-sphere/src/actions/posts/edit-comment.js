"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function editComment(commentData) {
    try {
        const url = `/comments/${commentData.comment_id}`;
        const response = await serverApiRequest(url, {
            method: "POST",
            body: JSON.stringify(commentData),
            headers: {
                "Content-Type": "application/json"
            }
        });

        if (!response.ok) {
            return {success: false, status: response.status, error: response.message};
        }

        return { success: true, data: response.data };

    } catch (error) {
        console.error("Edit Comment Action Error:", error);
        return { success: false, error: error.message };
    }
}
