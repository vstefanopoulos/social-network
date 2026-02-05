"use server";

import { serverApiRequest } from "@/lib/server-api";

export async function updateProfileInfo(data) {
    try {
        const url = `/my/profile`;
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

        return { success: true, data: response.data };
    } catch (error) {
        console.error("Error updating profile info:", error);
        return { success: false, error: error.message };
    }
}

export async function updateProfilePrivacy({ bool }) {
    try {
        const url = `/my/profile/privacy`;
        const response = await serverApiRequest(url, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                public: bool,
            }),
            forwardCookies: true
        });

        if (!response.ok) {
            return {success: false, status: response.status, error: response.message};
        }

        return { success: true, data: response.data };
    } catch (error) {
        console.error("Error updating profile privacy:", error);
        return { success: false, error: error.message };
    }
}

export async function updateProfileEmail({ email }) {
    try {
        const url = `/my/profile/email`;
        const response = await serverApiRequest(url, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                email: email,
            }),
            forwardCookies: true
        });

        if (!response.ok) {
            return {success: false, status: response.status, error: response.message};
        }

        return { success: true, data: response.data };
    } catch (error) {
        console.error("Error updating profile email:", error);
        return { success: false, error: error.message };
    }
}

export async function updateProfilePassword({ oldPassword, newPassword }) {
    try {
        const url = `/my/profile/password`;
        const response = await serverApiRequest(url, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                old_password: oldPassword,
                new_password: newPassword,
            }),
            forwardCookies: true
        });

        if (!response.ok) {
            return {success: false, status: response.status, error: response.message};
        }

        return { success: true, data: response.data };
    } catch (error) {
        console.error("Error updating profile password:", error);
        return { success: false, error: error.message };
    }
}
