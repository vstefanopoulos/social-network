"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function deleteComment(commentId) {
    try {
        const url = `/comments/${commentId}`;
        const response = await serverApiRequest(url, {
            method: "DELETE"
        });

        if (!response.ok) {
            return {success: false, status: response.status, error: response.message};
        }

        return { success: true, data: response.data };

    } catch (error) {
        console.error("Delete Comment Action Error:", error);
        return { success: false, error: error.message };
    }
}
