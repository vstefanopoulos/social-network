"use client";

import { useState, useRef } from "react";
import { X, Image as ImageIcon } from "lucide-react";
import { createGroup } from "@/actions/groups/create-group";
import { validateUpload } from "@/actions/auth/validate-upload";
import { validateImage } from "@/lib/validation";
import Modal from "@/components/ui/Modal";
import { useRouter } from "next/navigation";

export default function CreateGroup({ isOpen, onClose }) {
    const [title, setTitle] = useState("");
    const [description, setDescription] = useState("");
    const [imageFile, setImageFile] = useState(null);
    const [imagePreview, setImagePreview] = useState(null);
    const [error, setError] = useState("");
    const [isSubmitting, setIsSubmitting] = useState(false);
    const fileInputRef = useRef(null);
    const router = useRouter();

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

    const handleSubmit = async () => {
        // Validation
        if (!title.trim()) {
            setError("Please enter a group title");
            return;
        }

        if (!description.trim()) {
            setError("Please enter a group description");
            return;
        }

        setIsSubmitting(true);
        setError("");

        try {
            const groupData = {
                group_title: title.trim(),
                group_description: description.trim(),
            };

            // Add image data if image is selected
            if (imageFile) {
                groupData.group_image_name = imageFile.name;
                groupData.group_image_size = imageFile.size;
                groupData.group_image_type = imageFile.type;
            }

            const response = await createGroup(groupData);

            console.log("RESPONSE", response)

            if (!response.success) {
                setError(response.error || "Failed to create group");
                setIsSubmitting(false);
                return;
            }


            // If there's an image, try to upload it (non-blocking)
            let imageUploadFailed = false;
            if (imageFile && response?.FileId && response?.UploadUrl) {
                try {
                    const uploadRes = await fetch(response.UploadUrl, {
                        method: "PUT",
                        body: imageFile,
                    });

                    if (uploadRes.ok) {
                        const validateResp = await validateUpload(response.FileId);
                        if (!validateResp.success) {
                            imageUploadFailed = true;
                            console.log("no success")
                        }
                        console.log("IMAGE SUCCESS")
                    } else {
                        console.log("not ok")
                        imageUploadFailed = true;
                    }
                } catch (uploadErr) {
                    console.error("Image upload failed:", uploadErr);
                    imageUploadFailed = true;
                }
            }

            console.log("No image")

            // Navigate to group page (group was created successfully)
            setIsSubmitting(false);
            const url = imageUploadFailed
                ? `/groups/${response.GroupId}?imageUploadFailed=true`
                : `/groups/${response.GroupId}`;
            console.log("Navigating to:", url, "GroupId:", response.GroupId);
            router.push(url);
        } catch (err) {
            console.error("Failed to create group:", err);
            setError("Failed to create group. Please try again.");
            setIsSubmitting(false);
        }
    };

    const handleClose = () => {
        if (isSubmitting) return;
        setTitle("");
        setDescription("");
        setImageFile(null);
        setImagePreview(null);
        setError("");
        onClose();
    };

    return (
        <Modal
            isOpen={isOpen}
            onClose={handleClose}
            title="Create Group"
            description="Start a new community around your interests"
            showCloseButton={!isSubmitting}
        >
            <div className="space-y-4">
                {/* Title Input */}
                <div>
                    <label className="block text-sm font-medium text-foreground mb-2">
                        Group Title <span className="text-red-500">*</span>
                    </label>
                    <input
                        type="text"
                        value={title}
                        onChange={(e) => setTitle(e.target.value)}
                        placeholder="Enter group title"
                        disabled={isSubmitting}
                        className="w-full rounded-xl border border-(--muted)/30 px-4 py-2.5 text-sm bg-(--muted)/5 focus:outline-none focus:border-(--accent) focus:ring-2 focus:ring-(--accent)/10 transition-all disabled:opacity-50"
                        maxLength={100}
                    />
                </div>

                {/* Description Input */}
                <div>
                    <label className="block text-sm font-medium text-foreground mb-2">
                        Description <span className="text-red-500">*</span>
                    </label>
                    <textarea
                        value={description}
                        onChange={(e) => setDescription(e.target.value)}
                        placeholder="Describe your group..."
                        disabled={isSubmitting}
                        rows={4}
                        className="w-full rounded-xl border border-(--muted)/30 px-4 py-2.5 text-sm bg-(--muted)/5 focus:outline-none focus:border-(--accent) focus:ring-2 focus:ring-(--accent)/10 transition-all resize-none disabled:opacity-50"
                        maxLength={500}
                    />
                </div>

                {/* Image Upload */}
                <div>
                    <label className="block text-sm font-medium text-foreground mb-2 cursor-pointer">
                        Group Image (Optional)
                    </label>

                    {imagePreview ? (
                        <div className="relative inline-block">
                            <img
                                src={imagePreview}
                                alt="Group preview"
                                className="max-w-full max-h-48 rounded-xl border border-(--border)"
                            />
                            <button
                                type="button"
                                onClick={handleRemoveImage}
                                disabled={isSubmitting}
                                className="absolute -top-2 -right-2 bg-background text-(--muted) hover:text-red-500 rounded-full p-1.5 border border-(--border) shadow-sm transition-colors disabled:opacity-50 cursor-pointer"
                            >
                                <X className="w-4 h-4" />
                            </button>
                        </div>
                    ) : (
                        <div>
                            <input
                                ref={fileInputRef}
                                type="file"
                                accept="image/jpeg,image/png,image/gif,image/webp"
                                onChange={handleImageSelect}
                                disabled={isSubmitting}
                                className="hidden"
                            />
                            <button
                                type="button"
                                onClick={() => fileInputRef.current?.click()}
                                disabled={isSubmitting}
                                className="flex items-center gap-2 px-4 py-2.5 text-sm font-medium text-(--muted) hover:text-foreground hover:bg-(--muted)/10 border border-(--border) rounded-xl transition-colors disabled:opacity-50 cursor-pointer"
                            >
                                <ImageIcon className="w-4 h-4" />
                                Upload Image
                            </button>
                        </div>
                    )}
                </div>

                {/* Error Message */}
                {error && (
                    <div className="text-red-500 text-sm bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg px-4 py-2.5">
                        {error}
                    </div>
                )}

                {/* Action Buttons */}
                <div className="flex items-center justify-end gap-3 pt-2">
                    <button
                        onClick={handleClose}
                        disabled={isSubmitting}
                        className="px-4 py-2 text-sm font-medium text-(--muted) hover:text-foreground hover:bg-(--muted)/10 rounded-full transition-colors disabled:opacity-50 cursor-pointer"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleSubmit}
                        disabled={isSubmitting || !title.trim() || !description.trim()}
                        className="px-5 py-2 text-sm font-medium bg-(--accent) text-white hover:bg-(--accent-hover) rounded-full transition-colors disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed flex items-center gap-2"
                    >
                        {isSubmitting ? (
                            <>
                                <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                                Creating...
                            </>
                        ) : (
                            "Create Group"
                        )}
                    </button>
                </div>
            </div>
        </Modal>
    );
}
