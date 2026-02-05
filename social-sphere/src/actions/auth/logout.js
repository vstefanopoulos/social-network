"use server";

import { serverApiRequest } from "@/lib/server-api";
import { redirect } from "next/navigation";

export async function logout() {
    try {
        const res = await serverApiRequest("/logout", {
            method: "POST",
            forwardCookies: true
        });

        if (!res.ok) {
            return {success: false, status: res.status, error: res.message}
        }
    } catch (error) {
        console.error("Logout Action Error:", error);
        return { success: false, error: error.message };
    }

    // Redirect after successful logout
    redirect("/login");
}
