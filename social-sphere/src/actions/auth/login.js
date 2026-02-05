"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function login(credentials) {
    try {
        const apiResp = await serverApiRequest("/login", {
            method: "POST",
            body: JSON.stringify(credentials),
            forwardCookies: true,
            headers: {
                "Content-Type": "application/json"
            }
        });

        if(!apiResp.ok) {
            return { success: false, status: apiResp.status, error: apiResp.message };
        }

        if (!apiResp.data.id) {
            return {
                success: false,
                error: "Login failed - no user ID returned"
            };
        }

        return {
            success: true,
            user_id: apiResp.data.id,
            username: apiResp.data.username,
            avatar_url: apiResp.data.avatar_url || ""
        };

    } catch (error) {
        console.error("Login Action Error:", error);
        return { success: false, error: error.message };
    }
}
