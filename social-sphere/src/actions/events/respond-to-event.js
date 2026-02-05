"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function respondToEvent({id, going}) {
    try {

        
        const url = `/events/${id}/response`;

        console.log("answering: ", url , going)
        const apiResp = await serverApiRequest(url, {
            method: "POST",
            body: JSON.stringify({ going }),
            headers: {
                "Content-Type": "application/json"
            }
        });

        if (!apiResp.ok) {
            return {success: false, status: apiResp.status, error: apiResp.message};
        }

        return { success: true };

    } catch (error) {
        console.error("Respond to Event Action Error:", error);
        return { success: false, error: error.message };
    }
}
