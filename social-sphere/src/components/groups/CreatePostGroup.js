"use client";

import { useState, useRef } from "react";
import { motion, AnimatePresence } from "motion/react";
import { X, Image as ImageIcon } from "lucide-react";
import Tooltip from "@/components/ui/Tooltip";
import { validateImage } from "@/lib/validation";
import { createPost } from "@/actions/posts/create-post";
import { validateUpload } from "@/actions/auth/validate-upload";
import { useStore } from "@/store/store";

export default function CreatePostGroup({ onPostCreated=null, groupId=null }) {
    const user = useStore((state) => state.user);
    const [content, setContent] = useState("");
    const [imageFile, setImageFile] = useState(null);
    const [imagePreview, setImagePreview] = useState(null);
    const [error, setError] = useState("");
    const [warning, setWarning] = useState("");
    const fileInputRef = useRef(null);

    const MAX_CHARS = 5000;
    const MIN_CHARS = 1;
    const privacy = "group";

    const handleImageSelect = async (e) => {
        const file = e.target.files?.[0];
        if (!file) return;

        // Validate image file (type, size, dimensions)
        const validation = await validateImage(file);
        if (!validation.valid) {
            setError(validation.error);
            return;
        }

        setImageFile(file);
        setError("");

        // Create preview
        const reader = new FileReader();
        reader.onloadend = () => {
            setImagePreview(reader.result);
        };
        reader.readAsDataURL(file);
    };

    const handleRemoveImage = () => {
        setImageFile(null);
        setImagePreview(null);
        if (fileInputRef.current) {
            fileInputRef.current.value = "";
        }
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError("");

        // Validation
        if (!content.trim()) {
            setError("Post content is required");
            return;
        }

        if (content.length < MIN_CHARS) {
            setError(`Post must be at least ${MIN_CHARS} character`);
            return;
        }

        if (content.length > MAX_CHARS) {
            setError(`Post must be at most ${MAX_CHARS} characters`);
            return;
        }


        try {
            // Prepare post data
            const postData = {
                post_body: content.trim(),
                audience: privacy,
                group_id: groupId
            };

            // Add image metadata if image is selected
            if (imageFile) {
                postData.image_name = imageFile.name;
                postData.image_size = imageFile.size;
                postData.image_type = imageFile.type;
            }

            // Create post with metadata
            const resp = await createPost(postData);

            if (!resp.success) {
                setError(resp.error || "Failed to create post");
                return;
            }

            // Step 2: Upload image if needed (non-blocking)
            let imageUrl = null;
            let imageUploadFailed = false;
            if (imageFile && resp.FileId && resp.UploadUrl) {
                try {
                    const uploadRes = await fetch(resp.UploadUrl, {
                        method: "PUT",
                        body: imageFile,
                    });

                    if (uploadRes.ok) {
                        // Step 3: Validate the upload
                        const validateResp = await validateUpload(resp.FileId);
                        if (validateResp.success) {
                            imageUrl = validateResp.data?.download_url;
                        } else {
                            imageUploadFailed = true;
                        }
                    } else {
                        imageUploadFailed = true;
                    }
                } catch (uploadErr) {
                    console.error("Image upload failed:", uploadErr);
                    imageUploadFailed = true;
                }
            }

            const now = new Date().toISOString();

            const newPost = {
                comments_count: 0,
                image: imageUploadFailed ? null : resp.data.FileId,
                image_url: imageUrl,
                liked_by_user: false,
                post_body: content,
                post_id: resp.data.PostId,
                reactions_count: 0,
                created_at: now,
                post_user: {
                    avatar_url: user.avatar_url,
                    id: user.id,
                    username: user.username
                }
            }

            // Reset form
            setContent("");
            handleRemoveImage();

            // Show warning if image upload failed
            if (imageUploadFailed) {
                setWarning("Image failed to upload. You can try again later.");
                setTimeout(() => setWarning(""), 3000);
            }

            // Refresh the page to show the new post
            if (onPostCreated) {
                onPostCreated(newPost);
            }
            
        } catch (err) {
            console.error("Failed to create post:", err);
            setError("Failed to create post. Please try again.");
        }
    };

    const handleCancel = () => {
        setContent("");
        handleRemoveImage();
        setError("");
        setWarning("");
    };

    const charCount = content.length;
    const isOverLimit = charCount > MAX_CHARS;
    const isValid = content.trim().length >= MIN_CHARS && !isOverLimit;

    return (
        <div className="bg-background rounded-2xl p-3">
            <form onSubmit={handleSubmit}>
                {/* Textarea with character counter */}
                <div className="relative">
                    <textarea
                        value={content}
                        onChange={(e) => setContent(e.target.value)}
                        placeholder="Share something with your group.."
                        rows={3}
                        className="w-full bg-(--muted)/5 border border-(--border) rounded-xl px-2 py-3 pr-20 text-foreground placeholder:text-(--muted)/60 hover:border-foreground focus:outline-none focus:border-(--accent) focus:ring-2 focus:ring-(--accent)/10 transition-all resize-none"
                    />
                    {/* Character counter - bottom right */}
                    <div className="absolute bottom-3 right-3 text-xs">
                        <span
                            className={`font-medium ${isOverLimit
                                ? "text-red-500"
                                : charCount > MAX_CHARS * 0.9
                                    ? "text-orange-500"
                                    : "text-(--muted)/60"
                                }`}
                        >
                            {charCount > 0 && `${charCount}/${MAX_CHARS}`}
                        </span>
                    </div>
                </div>

                {/* Image Preview */}
                {imagePreview && (
                    <div className="relative inline-block mt-3">
                        <img
                            src={imagePreview}
                            alt="Upload preview"
                            className="max-w-full max-h-64 rounded-xl border border-(--border)"
                        />
                        <button
                            type="button"
                            onClick={handleRemoveImage}
                            className="absolute -top-2 -right-2 bg-background text-(--muted) hover:text-red-500 rounded-full p-1.5 border border-(--border) shadow-sm transition-colors cursor-pointer"
                        >
                            <X size={16} />
                        </button>
                    </div>
                )}

                {/* Error Message */}
                {error && (
                    <div className="text-red-500 text-sm bg-red-50 border border-red-200 rounded-lg px-4 py-2.5 animate-fade-in">
                        {error}
                    </div>
                )}

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

               

                {/* Bottom Controls Row */}
                <div className="flex flex-wrap items-center justify-between gap-2 pt-2">
                    {/* Left side: Privacy and Image Upload */}
                    <div className="flex items-center gap-2">
                        {/* Image Upload Button */}
                        <input
                            ref={fileInputRef}
                            type="file"
                            accept="image/jpeg,image/png,image/gif,image/webp"
                            onChange={handleImageSelect}
                            className="hidden"
                        />
                        <Tooltip content="Upload image">
                            <button
                                type="button"
                                onClick={() => fileInputRef.current?.click()}
                                className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-(--muted) border hover:text-foreground hover:bg-(--muted)/10 rounded-full transition-colors cursor-pointer"
                            >
                                <ImageIcon size={18} />
                                <span>Image</span>
                            </button>
                        </Tooltip>
                    </div>

                    {/* Right side: Submit and Cancel Buttons */}
                    <div className="flex items-center gap-2">
                        {(content || imageFile) && (
                            <>
                                <button
                                    type="button"
                                    onClick={handleCancel}
                                    className="px-4 py-1.5 text-sm text-(--muted) hover:text-foreground hover:bg-(--muted)/10 rounded-full transition-colors cursor-pointer"
                                >
                                    Cancel
                                </button>
                                <button
                                    type="submit"
                                    disabled={!isValid}
                                    className="px-5 py-1.5 text-sm font-medium bg-(--accent) text-white hover:bg-(--accent-hover) rounded-full disabled:opacity-50 disabled:cursor-not-allowed transition-colors cursor-pointer"
                                >
                                    Post
                                </button>
                            </>
                        )}
                    </div>
                </div>
            </form>
        </div>
    );
}
