"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function toggleReaction(postId) {
    try {

        const response = await serverApiRequest("/reactions", {
            method: "POST",
            body: JSON.stringify({ entity_id: postId }),
            headers: {
                "Content-Type": "application/json"
            }
        });

        if (!response.ok) {
            return {success: false, status: response.status, error: response.message};
        }

        return { success: true, data: response.data };

    } catch (error) {
        console.error("Toggle Reaction Action Error:", error);
        return { success: false, error: error.message };
    }
}
