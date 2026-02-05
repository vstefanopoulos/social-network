"use server";

import { serverApiRequest } from "@/lib/server-api";

// na fugei id apo data kai na mpei sto URL
export async function updateGroup({groupId ,data}) {
    try {
        const url = `/groups/${groupId}`;
        const response = await serverApiRequest(url, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(data),
            forwardCookies: true
        });

        if (!response.ok) {
            return {success: false, status: response.status, error: response.message};
        }

        return {
            success: true,
            GroupId: response.data.GroupId,
            FileId: response.data.FileId,
            UploadUrl: response.data.UploadUrl
        };
    } catch (error) {
        console.error("Error updating group:", error);
        return { success: false, error: error.message };
    }
}
