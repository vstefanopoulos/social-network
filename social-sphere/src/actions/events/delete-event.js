"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function deleteEvent(eventId) {
    try {
        const url = `/events/${eventId}`;
        const apiResp = await serverApiRequest(url, {
            method: "DELETE",
            headers: {
                "Content-Type": "application/json"
            }
        });

        if (!apiResp.ok) {
            return {success: false, status: apiResp.status, error: apiResp.message};
        }

        return { success: true, data: apiResp.data };

    } catch (error) {
        console.error("Delete Event Action Error:", error);
        return { success: false, error: error.message };
    }
}
