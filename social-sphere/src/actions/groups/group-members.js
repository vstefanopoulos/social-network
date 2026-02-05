"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function getGroupMembers({ group_id, limit, offset }) {
    try {
        const url = `/groups/${group_id}/members?limit=${limit}&offset=${offset}`;
        const response = await serverApiRequest(url, {
            method: "GET",
            forwardCookies: true
        });

        if (!response.ok) {
            return {success: false, status: response.status, error: response.message};
        }

        return { success: true, data: response.data };

    } catch (error) {
        console.error("Error fetching group members: ", error);
        return { success: false, error: error.message };
    }
}
