"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function getImageUrl({fileId, variant}) {
    try {
        const url = `/files/images/${fileId}/${variant}`;
        const res = await serverApiRequest(url, {
            method: "GET",
            forwardCookies: true,
        });

        if (!res.ok) {
            return {success: false, status: res.status, error: res.message};
        }
        return { success: true, data: res.data };

    } catch (error) {
        console.error("Error fetching post:", error);
        return {success: false, error: error.message};
    }
}