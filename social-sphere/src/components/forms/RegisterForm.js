"use client";

import { useState, useEffect } from "react";
import { motion, AnimatePresence } from "motion/react";
import { Eye, EyeOff, Upload, X } from "lucide-react";
import { useFormValidation } from "@/hooks/useFormValidation";
import { register } from "@/actions/auth/register";
import { validateUpload } from "@/actions/auth/validate-upload";
import { useStore } from "@/store/store";
import LoadingThreeDotsJumping from '@/components/ui/LoadingDots';
import { validateRegistrationForm, validateImage } from "@/lib/validation";

export default function RegisterForm() {
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState(null);
    const [imageError, setImageError] = useState(null);
    const [warning, setWarning] = useState("");
    const [showPassword, setShowPassword] = useState(false);
    const [showConfirmPassword, setShowConfirmPassword] = useState(false);
    const [avatarPreview, setAvatarPreview] = useState(null);
    const [avatarName, setAvatarName] = useState(null);
    const [avatarFile, setAvatarFile] = useState(null);
    const setUser = useStore((state) => state.setUser);
    const clearUser = useStore((state) => state.clearUser);

    // Clear any stale user data when register page loads
    useEffect(() => {
        clearUser();
    }, [clearUser]);

    // Real-time validation state
    const { errors: fieldErrors, validateField } = useFormValidation();
    const [aboutCount, setAboutCount] = useState(0);

    async function handleSubmit(event) {
        event.preventDefault();
        setIsLoading(true);
        setError("");

        const rawFormData = new FormData(event.currentTarget);
        const userData = {
            username: rawFormData.get("username"),
            first_name: rawFormData.get("first_name"),
            last_name: rawFormData.get("last_name"),
            date_of_birth: rawFormData.get("date_of_birth"),
            about: rawFormData.get("about"),
            email: rawFormData.get("email"),
            password: rawFormData.get("password"),
        };

        // Add avatar metadata if present
        if (avatarFile) {
            userData.avatar_name = avatarFile.name;
            userData.avatar_size = avatarFile.size;
            userData.avatar_type = avatarFile.type;
        }
        try {

            const validation = await validateRegistrationForm(rawFormData, avatarFile);
            if (!validation.valid) {
                setError(validation.error || "Registration Failed");
                setIsLoading(false);
                return;
            }
            // Register with metadata
            const resp = await register(userData);

            if (!resp.success || resp.error) {
                setError(resp?.error || "Registration failed")
                setIsLoading(false);
                return;
            }

            // Prepare store data
            const userStoreData = {
                id: resp.data.UserId,
                fileId: resp.data.FileId,
                username: resp.data.Username,
                avatar_url: ""
            };

            let imageUploadFailed = false;

            // Step 2: Upload avatar if needed (non-blocking)
            if (avatarFile && resp.data.UploadUrl) {
                try {

                    const uploadRes = await fetch(resp.data.UploadUrl, {
                        method: "PUT",
                        body: avatarFile
                    });

                    if (uploadRes.ok) {
                        // Step 3: Validate upload
                        const validateResp = await validateUpload(resp.data.FileId);
                        if (validateResp.success && validateResp.data?.download_url) {
                            userStoreData.avatar_url = validateResp.data.download_url;
                        } else {
                            imageUploadFailed = true;
                        }
                    } else {
                        const errorText = await uploadRes.text();
                        console.error(`Storage error (${uploadRes.status}): ${errorText}`);
                        imageUploadFailed = true;
                    }
                } catch (uploadError) {
                    console.error("Avatar upload error:", uploadError);
                    imageUploadFailed = true;
                }
            }

            // Step 4: Store user data and redirect
            setUser(userStoreData);

            // Show warning if avatar upload failed (non-blocking)
            if (imageUploadFailed) {
                setWarning("Avatar failed to upload. You can update it later in your profile.");
                // Redirect after a short delay so user can see the warning
                setTimeout(() => {
                    window.location.href = "/feed/public";
                }, 2000);
            } else {
                window.location.href = "/feed/public";
            }

        } catch (error) {
            console.error("Registration exception:", error);
            setError("An unexpected error occurred");
            setIsLoading(false);
        }
    }

    async function handleAvatarChange(event) {
        const file = event.target.files[0];
        if (!file) return;

        // Validate image file (type, size, dimensions)
        const validation = await validateImage(file);
        if (!validation.valid) {
            setImageError(validation.error);
            return;
        }

        setImageError(null);
        setAvatarName(file.name);
        setAvatarFile(file);
        const reader = new FileReader();
        reader.onloadend = () => {
            setAvatarPreview(reader.result);
        };
        reader.readAsDataURL(file);
    }

    function removeAvatar() {
        setAvatarPreview(null);
        setAvatarName("");
        setAvatarFile(null);
    }

    // Real-time validation handlers
    function handleFieldValidation(name, value) {
        switch (name) {
            case "first_name":
                validateField("first_name", value, (val) => {
                    if (!val.trim()) return "First name is required.";
                    if (val.trim().length < 2) return "First name must be at least 2 characters.";
                    return null;
                });
                break;

            case "last_name":
                validateField("last_name", value, (val) => {
                    if (!val.trim()) return "Last name is required.";
                    if (val.trim().length < 2) return "Last name must be at least 2 characters.";
                    return null;
                });
                break;

            case "email":
                validateField("email", value, (val) => {
                    const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
                    if (!emailPattern.test(val.trim())) return "Please enter a valid email address.";
                    return null;
                });
                break;

            case "password":
                validateField("password", value, (val) => {
                    const strongPattern = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[^\w\s]).+$/;
                    if (val.length < 8) return "Password must be at least 8 characters.";
                    if (!strongPattern.test(val)) return "Password needs 1 lowercase, 1 uppercase, 1 number, and 1 symbol.";
                    return null;
                });
                break;

            case "confirmPassword":
                validateField("confirmPassword", value, (val) => {
                    const passwordField = document.querySelector('input[name="password"]');
                    if (passwordField && val !== passwordField.value) return "Passwords do not match.";
                    return null;
                });
                break;

            case "date_of_birth":
                validateField("date_of_birth", value, (val) => {
                    if (!val) return "Date of birth is required.";
                    const today = new Date();
                    const birthDate = new Date(val);
                    let age = today.getFullYear() - birthDate.getFullYear();
                    const monthDiff = today.getMonth() - birthDate.getMonth();
                    if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < birthDate.getDate())) {
                        age--;
                    }
                    if (age < 13 || age > 111) return "You must be between 13 and 111 years old.";
                    return null;
                });
                break;

            case "username":
                validateField("username", value, (val) => {
                    if (val.trim()) {
                        const usernamePattern = /^[A-Za-z0-9_.-]+$/;
                        if (val.trim().length < 4) return "Username must be at least 4 characters.";
                        if (!usernamePattern.test(val.trim())) return "Username can only use letters, numbers, dots, underscores, or dashes.";
                    }
                    return null;
                });
                break;

            case "about":
                setAboutCount(value.length);
                validateField("about", value, (val) => {
                    if (val.length > 400) return "About me must be at most 400 characters.";
                    return null;
                });
                break;
        }
    }

    return (
        <>
            {/* Warning Message - Fixed top banner */}
            <AnimatePresence>
                {warning && (
                    <motion.div
                        initial={{ opacity: 0, y: -20 }}
                        animate={{ opacity: 1, y: 0 }}
                        exit={{ opacity: 0, y: -20 }}
                        className="fixed top-4 left-1/2 -translate-x-1/2 z-50 px-4 py-2 bg-amber-500/90 text-white text-sm rounded-lg shadow-lg"
                    >
                        {warning}
                    </motion.div>
                )}
            </AnimatePresence>

            <form onSubmit={handleSubmit} className="w-full">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-12">
                {/* LEFT COLUMN - Account Info */}
                <div className="space-y-6">
                    <h3 className="text-lg font-semibold text-foreground mb-4">Account Information</h3>

                    {/* Name Fields */}
                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <label htmlFor="first_name" className="form-label pl-4">First Name <span className="text-red-500">*</span></label>
                            <input
                                id="first_name"
                                name="first_name"
                                type="text"
                                defaultValue="hello"
                                required
                                className="form-input"
                                placeholder="Jane"
                                onChange={(e) => handleFieldValidation("first_name", e.target.value)}
                            />
                            {fieldErrors.first_name && (
                                <div className="form-error">{fieldErrors.first_name}</div>
                            )}
                        </div>
                        <div>
                            <label htmlFor="last_name" className="form-label pl-4">Last Name <span className="text-red-500">*</span></label>
                            <input
                                id="last_name"
                                name="last_name"
                                type="text"
                                defaultValue="world"
                                required
                                className="form-input"
                                placeholder="Doe"
                                onChange={(e) => handleFieldValidation("last_name", e.target.value)}
                            />
                            {fieldErrors.last_name && (
                                <div className="form-error">{fieldErrors.last_name}</div>
                            )}
                        </div>
                    </div>

                    {/* Email */}
                    <div>
                        <label htmlFor="email" className="form-label pl-4">Email <span className="text-red-500">*</span></label>
                        <input
                            id="email"
                            name="email"
                            type="email"
                            defaultValue="hello@world.com"
                            required
                            className="form-input"
                            placeholder="jane@example.com"
                            onChange={(e) => handleFieldValidation("email", e.target.value)}
                        />
                        {fieldErrors.email && (
                            <div className="form-error">{fieldErrors.email}</div>
                        )}
                    </div>

                    {/* Password */}
                    <div>
                        <label htmlFor="password" className="form-label pl-4">Password <span className="text-red-500">*</span></label>
                        <div className="relative">
                            <input
                                id="password"
                                name="password"
                                type={showPassword ? "text" : "password"}
                                defaultValue="Hello12!"
                                required
                                className="form-input pr-12"
                                placeholder="HelloWorld123!"
                                minLength={8}
                                onChange={(e) => handleFieldValidation("password", e.target.value)}
                            />
                            <button
                                type="button"
                                onClick={() => setShowPassword(!showPassword)}
                                className="form-toggle-btn p-2"
                            >
                                {showPassword ? <EyeOff size={20} className="rounded-full" /> : <Eye size={20} className="rounded-full" />}
                            </button>
                        </div>
                        {fieldErrors.password && (
                            <div className="form-error">{fieldErrors.password}</div>
                        )}
                    </div>

                    {/* Confirm Password */}
                    <div>
                        <label htmlFor="confirmPassword" className="form-label pl-4">Confirm Password <span className="text-red-500">*</span></label>
                        <div className="relative">
                            <input
                                id="confirmPassword"
                                name="confirmPassword"
                                type={showConfirmPassword ? "text" : "password"}
                                defaultValue="Hello12!"
                                required
                                className="form-input pr-12"
                                placeholder="Confirm password"
                                minLength={8}
                                onChange={(e) => handleFieldValidation("confirmPassword", e.target.value)}
                            />
                            <button
                                type="button"
                                onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                                className="form-toggle-btn p-2"
                            >
                                {showConfirmPassword ? <EyeOff size={20} /> : <Eye size={20} />}
                            </button>
                        </div>
                        {fieldErrors.confirmPassword && (
                            <div className="form-error">{fieldErrors.confirmPassword}</div>
                        )}
                    </div>

                    {/* Date of Birth */}
                    <div>
                        <label htmlFor="date_of_birth" className="form-label pl-4">Date of Birth <span className="text-red-500">*</span></label>
                        <input
                            id="date_of_birth"
                            name="date_of_birth"
                            type="date"
                            defaultValue="2000-01-01"
                            required
                            className="form-input focus:outline-none"
                            onChange={(e) => handleFieldValidation("date_of_birth", e.target.value)}
                        />
                        {fieldErrors.date_of_birth && (
                            <div className="form-error">{fieldErrors.date_of_birth}</div>
                        )}
                    </div>
                </div>

                {/* RIGHT COLUMN - Profile Info */}
                <div className="space-y-6">
                    <h3 className="text-lg font-semibold text-foreground mb-4">Profile Details</h3>

                    {/* Avatar Upload */}
                    <div className="flex flex-col items-center gap-3">
                        <div className="relative">
                            <div className="avatar-container">
                                {avatarPreview ? (
                                    <img src={avatarPreview} alt="Avatar preview" className="avatar-image" />
                                ) : (
                                    <Upload className="text-(--muted) w-8 h-8" />
                                )}
                                <input
                                    type="file"
                                    name="avatar"
                                    accept="image/jpeg,image/png,image/gif,image/webp"
                                    onChange={handleAvatarChange}
                                    className="avatar-input"
                                />
                            </div>

                            {/* X button - only show when avatar is uploaded */}
                            {avatarPreview && (
                                <button
                                    type="button"
                                    onClick={removeAvatar}
                                    className="absolute -top-1 -right-1 w-6 h-6 bg-red-500 text-white rounded-full flex items-center justify-center hover:bg-red-600 transition-colors z-10"
                                >
                                    <X size={14} />
                                </button>
                            )}
                        </div>

                        {!imageError ? (
                            <span className="text-sm text-muted">
                                {avatarName || "Upload Avatar (Optional)"}
                            </span>
                        ) : (
                            <span className="text-sm text-red-500">
                                {imageError}
                            </span>
                        )}

                    </div>

                    {/* Username */}
                    <div>
                        <label htmlFor="username" className="form-label pl-4">Username (Optional)</label>
                        <input
                            id="username"
                            name="username"
                            type="text"
                            className="form-input"
                            placeholder="@janed"
                            onChange={(e) => handleFieldValidation("username", e.target.value)}
                        />
                        {fieldErrors.username && (
                            <div className="form-error">{fieldErrors.username}</div>
                        )}
                    </div>

                    {/* About Me */}
                    <div>
                        <div className="flex items-center justify-between mb-2">
                            <label htmlFor="about" className="form-label pl-4 mb-0">About Me (Optional)</label>
                            <span className="text-xs text-muted">
                                {aboutCount}/400
                            </span>
                        </div>
                        <textarea
                            id="about"
                            name="about"
                            rows={5}
                            maxLength={400}
                            className="form-input resize-none"
                            placeholder="Tell us a bit about yourself..."
                            onChange={(e) => handleFieldValidation("about", e.target.value)}
                        />
                        {fieldErrors.about && (
                            <div className="form-error">{fieldErrors.about}</div>
                        )}
                    </div>
                </div>
            </div>

            {/* Error Message */}
            {error && (
                <div className="form-error animate-fade-in mt-6 text-center pt-5">
                    {error}
                </div>
            )}

            {/* Submit Button */}
            <button
                type="submit"
                disabled={isLoading}
                className="w-1/3 mx-auto flex justify-center items-center btn btn-primary mt-12"
            >
                {isLoading ? <LoadingThreeDotsJumping /> : "Create Account"}
            </button>

        </form>
        </>
    );
}
