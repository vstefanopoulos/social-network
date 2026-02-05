"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function register(userData) {
    try {
        // register with a public profile by default
        userData.public = true;

        const apiResp = await serverApiRequest("/register", {
            method: "POST",
            body: JSON.stringify(userData),
            headers: {
                "Content-Type": "application/json"
            },
            forwardCookies: true
        });


        if (!apiResp.ok) {
            return {success: false, status: apiResp.status, error: apiResp.message};
        }

        return { success: true, data: apiResp.data };

    } catch (error) {
        console.error("Register Action Error:", error);
        return { success: false, error: error.message };
    }
}
