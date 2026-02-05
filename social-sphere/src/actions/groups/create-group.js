"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function createGroup({
    group_title,
    group_description,
    group_image_name,
    group_image_size,
    group_image_type
}) {
    try {
        const requestBody = {
            group_title,
            group_description,
        };

        // Add image fields if image is being uploaded
        if (group_image_size && group_image_name && group_image_type) {
            requestBody.group_image_name = group_image_name;
            requestBody.group_image_size = group_image_size;
            requestBody.group_image_type = group_image_type;
        }

        const response = await serverApiRequest(`/groups`, {
            method: "POST",
            body: JSON.stringify(requestBody),
            forwardCookies: true,
            headers: {
                "Content-Type": "application/json"
            }
        });

        if (!response.ok) {
            return {success: false, status: response.status, error: response.message};
        }

        return {
            success: true,
            GroupId: response.data.group_id,
            FileId: response.data.FileId,
            UploadUrl: response.data.UploadUrl
        };

    } catch (error) {
        console.error("Error creating group:", error);
        return { success: false, error: error.message };
    }
}
