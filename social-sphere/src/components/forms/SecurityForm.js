"use client";

import { useActionState, useState } from "react";
import { updatePasswordAction, updateEmailAction } from "@/actions/profile/settings";
import { Eye, EyeOff, Loader2 } from "lucide-react";

const initialState = {
    success: false,
    message: null,
};

export default function SecurityForm({ user }) {
    const [emailState, emailAction, isEmailPending] = useActionState(updateEmailAction, initialState);
    const [passwordState, passwordAction, isPasswordPending] = useActionState(updatePasswordAction, initialState);

    const [showCurrentPassword, setShowCurrentPassword] = useState(false);
    const [showNewPassword, setShowNewPassword] = useState(false);
    const [showConfirmPassword, setShowConfirmPassword] = useState(false);

    return (
        <div className="space-y-12 animate-in fade-in slide-in-from-bottom-4 duration-500">
            {/* Email Section */}
            <div className="space-y-6">
                <div>
                    <h3 className="text-lg font-semibold">Email Address</h3>
                    <p className="text-sm text-(--muted)">Update the email address associated with your account.</p>
                </div>

                <form action={emailAction} className="space-y-4 max-w-md">
                    <div className="form-group">
                        <label className="form-label pl-4">Email</label>
                        <input
                            type="email"
                            name="email"
                            defaultValue={user?.email || "user@example.com"} // Mock default
                            className="form-input"
                            placeholder="your@email.com"
                            required
                        />
                    </div>

                    {emailState.message && (
                        <div className={`p-3 rounded-xl text-sm ${emailState.success ? 'bg-purple-500/10 text-(--accent)' : 'bg-red-500/10 text-red-600'}`}>
                            {emailState.message}
                        </div>
                    )}

                    <div className="flex justify-end">
                        <button
                            type="submit"
                            disabled={isEmailPending}
                            className="btn btn-primary px-6 flex items-center gap-2"
                        >
                            {isEmailPending && <Loader2 className="w-4 h-4 animate-spin" />}
                            {isEmailPending ? "Updating..." : "Update Email"}
                        </button>
                    </div>
                </form>
            </div>

            <div className="h-px bg-(--muted)/10" />

            {/* Password Section */}
            <div className="space-y-6">
                <div>
                    <h3 className="text-lg font-semibold">Change Password</h3>
                    <p className="text-sm text-(--muted)">Ensure your account is secure by using a strong password.</p>
                </div>

                <form action={passwordAction} className="space-y-4 max-w-md">
                    <div className="form-group">
                        <label className="form-label pl-4">Current Password</label>
                        <div className="relative">
                            <input
                                type={showCurrentPassword ? "text" : "password"}
                                name="currentPassword"
                                className="form-input pr-10"
                                placeholder="••••••••"
                                required
                            />
                            <button
                                type="button"
                                onClick={() => setShowCurrentPassword(!showCurrentPassword)}
                                className="form-toggle-btn"
                            >
                                {showCurrentPassword ? <EyeOff size={20} /> : <Eye size={20} />}
                            </button>
                        </div>
                    </div>
                    <div className="form-group">
                        <label className="form-label pl-4">New Password</label>
                        <div className="relative">
                            <input
                                type={showNewPassword ? "text" : "password"}
                                name="newPassword"
                                className="form-input pr-10"
                                placeholder="••••••••"
                                required
                            />
                            <button
                                type="button"
                                onClick={() => setShowNewPassword(!showNewPassword)}
                                className="form-toggle-btn"
                            >
                                {showNewPassword ? <EyeOff size={20} /> : <Eye size={20} />}
                            </button>
                        </div>
                    </div>
                    <div className="form-group">
                        <label className="form-label pl-4">Confirm New Password</label>
                        <div className="relative">
                            <input
                                type={showConfirmPassword ? "text" : "password"}
                                name="confirmPassword"
                                className="form-input pr-10"
                                placeholder="••••••••"
                                required
                            />
                            <button
                                type="button"
                                onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                                className="form-toggle-btn"
                            >
                                {showConfirmPassword ? <EyeOff size={20} /> : <Eye size={20} />}
                            </button>
                        </div>
                    </div>

                    {passwordState.message && (
                        <div className={`p-3 rounded-xl text-sm ${passwordState.success ? 'bg-purple-500/10 text-(--accent)' : 'bg-red-500/10 text-red-600'}`}>
                            {passwordState.message}
                        </div>
                    )}

                    <div className="flex justify-end">
                        <button
                            type="submit"
                            disabled={isPasswordPending}
                            className="btn btn-primary px-6 flex items-center gap-2"
                        >
                            {isPasswordPending && <Loader2 className="w-4 h-4 animate-spin" />}
                            {isPasswordPending ? "Updating..." : "Update Password"}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
}
