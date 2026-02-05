"use server";

import { cookies } from "next/headers";

const API_BASE = process.env.GATEWAY

export async function serverApiRequest(endpoint, options = {}) {
    try {
        const cookieStore = await cookies();
        const jwt = cookieStore.get("jwt")?.value;

        const headers = { ...(options.headers || {}) };
        if (jwt) headers["Cookie"] = `jwt=${jwt}`;

        const res = await fetch(`${API_BASE}${endpoint}`, {
            ...options,
            headers,
            cache: "no-store"
        });


        if (options.forwardCookies) {
            // Handle multiple Set-Cookie headers
            const setCookieHeaders = res.headers.getSetCookie ? res.headers.getSetCookie() : [];

            // Fallback for environments where getSetCookie might not be available
            if (setCookieHeaders.length === 0) {
                const header = res.headers.get('Set-Cookie');
                if (header) setCookieHeaders.push(header);
            }

            if (setCookieHeaders.length > 0) {
                setCookieHeaders.forEach(cookieStr => {

                    const parts = cookieStr.split(';');
                    const [nameValue, ...optionsParts] = parts;
                    const [name, ...valueParts] = nameValue.split('=');
                    const value = valueParts.join('=');

                    if (name && value !== undefined) {
                        const cookieOptions = {
                            secure: true,
                            httpOnly: true,
                            path: '/',
                            sameSite: 'lax', // Default safe value
                        };

                        optionsParts.forEach(part => {
                            const [optKey, optVal] = part.trim().split('=');
                            const keyLower = optKey.toLowerCase();
                            if (keyLower === 'path') cookieOptions.path = optVal;
                            if (keyLower === 'httponly') cookieOptions.httpOnly = true;
                            if (keyLower === 'secure') cookieOptions.secure = true;
                            if (keyLower === 'samesite') cookieOptions.sameSite = optVal.toLowerCase();
                            if (keyLower === 'max-age') cookieOptions.maxAge = parseInt(optVal);
                            if (keyLower === 'expires') cookieOptions.expires = new Date(optVal);
                        });

                        cookieStore.set(name.trim(), value, cookieOptions);
                    }
                });
            }
        }

        if (!res.ok) {
            const err = await res.json().catch(() => ({}));
            // show error
            console.log("ERROR: ", err);

            // show response status
            console.log("STATUS: ", res.status);
            if (res.status === 403) {
                return {ok: false, status: res.status, message: err.error || err.message || "Forbidden"}
            }
            if (res.status === 400) {
                return {ok: false, status: res.status, message: err.error || err.message || "Bad request"}
            }

            return {ok: false, status: res.status, message: err.error || err.message || "Unknown error"}
        }

        // Handle empty response bodies (like delete endpoints)
        const text = await res.text();
        if (!text || text.trim() === '') {
            return {ok: true, data: null};
        } else {
            console.log("Data: ", JSON.parse(text))
            return {ok: true, data: JSON.parse(text)};
        }
        
    } catch (e) {
        console.error('Failed to parse JSON response:', text);
        //throw new Error('Invalid JSON response from server');
    }
}
