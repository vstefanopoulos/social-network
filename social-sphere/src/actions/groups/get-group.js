"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function getGroup(groupId) {
    try {
        const url = `/groups/${groupId}`;
        const response = await serverApiRequest(url, {
            method: "GET"
        });

        if (!response.ok) {
            return {success: false, status: response.status, error: response.message};
        }

        return { success: true, data: response.data };

    } catch (error) {
        console.error("Error getting group:", error);
        return { success: false, error: error.message };
    }
}
