"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function SearchUsers({ query, limit }) {
    try {
        const url = `/users/search?query=${query}&limit=${limit}`;
        const response = await serverApiRequest(url, {
            method: "GET",
            forwardCookies: true
        });

        if (!response.ok) {
            return {success: false, status: response.status, error: response.message};
        }

        return { success: true, data: response.data };
    } catch (error) {
        console.error("Error searching users:", error);
        return { success: false, error: error.message };
    }
}
