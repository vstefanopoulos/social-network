"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function getGroupEvents({ groupId, limit = 10, offset = 0 }) {
    try {
        const url = `/groups/${groupId}/events?limit=${limit}&offset=${offset}`;
        const apiResp = await serverApiRequest(url, {
            method: "GET"
        });

        if (!apiResp.ok) {
            return {success: false, status: apiResp.status, error: apiResp.message};
        }

        return { success: true, data: apiResp.data };

    } catch (error) {
        console.error("Get Group Events Action Error:", error);
        return { success: false, error: error.message };
    }
}
