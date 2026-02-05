"use client";

import { useState, useTransition } from "react";
import { updatePrivacyAction } from "@/actions/profile/settings";
import { Lock, Globe, Loader2 } from "lucide-react";

export default function PrivacyForm({ user }) {
    // Note: user.publicProf is confusing naming (public vs private).
    // If publicProf is true, isPrivate is false.
    // Assuming user.publicProf exists.
    const [isPrivate, setIsPrivate] = useState(!user?.public);
    const [isPending, startTransition] = useTransition();
    const [message, setMessage] = useState(null);

    function handleToggle() {
        setMessage(null);
        const newState = !isPrivate;

        // Optimistic UI update
        setIsPrivate(newState);

        startTransition(async () => {
            try {
                // Pass the NEW state of "public" or "private"?
                // updatePrivacyAction takes "bool". updateProfilePrivacy doc says "public: bool".
                // So if we want to be private, public = false.
                // newState is "isPrivate". So public = !newState.
                const result = await updatePrivacyAction(!newState);

                if (result.success) {
                    setMessage({ type: "success", text: result.message });
                } else {
                    // Revert on failure
                    setIsPrivate(!newState);
                    setMessage({ type: "error", text: result.message });
                }
            } catch (error) {
                setIsPrivate(!newState); // Revert
                setMessage({ type: "error", text: "Something went wrong" });
            }
        });
    }

    return (
        <div className="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
            <div>
                <h3 className="text-lg font-semibold">Account Privacy</h3>
                <p className="text-sm text-(--muted)">Control who can see your profile and posts.</p>
            </div>

            <div className="p-6 rounded-2xl border border-(--muted)/20 bg-(--muted)/5">
                <div className="flex items-start justify-between gap-4">
                    <div className="flex gap-4">
                        <div className={`p-3 rounded-xl ${isPrivate ? 'bg-(--muted)/10 text-(--accent)' : 'bg-(--muted)/10 text-(--accent'}`}>
                            {isPrivate ? <Lock className="w-6 h-6" /> : <Globe className="w-6 h-6 text-(--accent)" />}
                        </div>
                        <div>
                            <h4 className="font-medium text-lg">
                                {isPrivate ? "Private Account" : "Public Account"}
                            </h4>
                            <p className="text-sm text-(--muted) mt-1 max-w-md leading-relaxed">
                                {isPrivate
                                    ? "Only people you approve can see your photos and videos. Your existing followers won't be affected."
                                    : "Anyone on or off the platform can see your photos and videos."
                                }
                            </p>
                        </div>
                    </div>

                    <button
                        onClick={handleToggle}
                        disabled={isPending}
                        className={`relative inline-flex h-7 w-12 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-foreground/20 focus:ring-offset-2 ${isPrivate ? 'bg-(--muted)/30' : 'bg-(--accent)/20'
                            }`}
                    >
                        <span
                            className={`${isPrivate ? 'translate-x-4 h-5 w-5 transform rounded-full bg-background transition-transform duration-200 flex items-center justify-center' : 'translate-x-1 h-5 w-5 transform rounded-full bg-(--accent) transition-transform duration-200 flex items-center justify-center z-index-1'
                                } `}
                        >
                            {isPending && <Loader2 className="w-3 h-3 animate-spin text-foreground" />}
                        </span>
                    </button>
                </div>
            </div>

            {message && (
                <div className={`p-4 rounded-xl text-sm ${message.type === 'success' ? 'bg-purple-500/10 text-(--accent)' : 'bg-red-500/10 text-red-600'}`}>
                    {message.text}
                </div>
            )}
        </div>
    );
}
