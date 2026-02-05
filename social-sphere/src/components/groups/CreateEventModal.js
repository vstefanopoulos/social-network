"use client";

import { useState, useRef } from "react";
import { motion, AnimatePresence } from "motion/react";
import { X, Image as ImageIcon, Calendar } from "lucide-react";
import { createEvent } from "@/actions/events/create-event";
import { validateUpload } from "@/actions/auth/validate-upload";
import { validateImage } from "@/lib/validation";
import Modal from "@/components/ui/Modal";
import { useStore } from "@/store/store";

export default function CreateEventModal({ isOpen, onClose, onSuccess, groupId }) {
    const user = useStore((state) => state.user);
    const [title, setTitle] = useState("");
    const [body, setBody] = useState("");
    const [eventDate, setEventDate] = useState("");
    const [imageFile, setImageFile] = useState(null);
    const [imagePreview, setImagePreview] = useState(null);
    const [error, setError] = useState("");
    const [warning, setWarning] = useState("");
    const [isSubmitting, setIsSubmitting] = useState(false);
    const fileInputRef = useRef(null);

    const MAX_TITLE_CHARS = 100;
    const MAX_BODY_CHARS = 1000;

    const handleImageSelect = async (e) => {
        const file = e.target.files?.[0];
        if (!file) return;

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
            setError("Please enter an event title");
            return;
        }

        if (!body.trim()) {
            setError("Please enter an event description");
            return;
        }

        if (!eventDate) {
            setError("Please select an event date and time");
            return;
        }

        // Check if date is in the future
        const selectedDate = new Date(eventDate);
        if (selectedDate <= new Date()) {
            setError("Event date must be in the future");
            return;
        }

        setIsSubmitting(true);
        setError("");

        try {
            const eventData = {
                event_title: title.trim(),
                event_body: body.trim(),
                event_date: new Date(eventDate).toISOString(),
            };

            // Add image data if selected
            if (imageFile) {
                eventData.image_name = imageFile.name;
                eventData.image_size = imageFile.size;
                eventData.image_type = imageFile.type;
            }

            const response = await createEvent({groupID: groupId, data:eventData});

            if (!response.success) {
                setError(response.error || "Failed to create event");
                setIsSubmitting(false);
                return;
            }

            // If there's an image, upload it (non-blocking)
            let imageUrl = null;
            let imageUploadFailed = false;
            if (imageFile && response.FileId && response.UploadUrl) {
                try {

                    const uploadRes = await fetch(response.UploadUrl, {
                        method: "PUT",
                        body: imageFile,
                    });

                    if (uploadRes.ok) {
                        const validateResp = await validateUpload(response.FileId);
                        if (validateResp.success) {
                            imageUrl = validateResp.data?.download_url;
                        } else {
                            imageUploadFailed = true;
                        }
                    } else {
                        imageUploadFailed = true;
                    }
                } catch (uploadErr) {
                    console.error("Event image upload failed:", uploadErr);
                    imageUploadFailed = true;
                }
            }

            // Create the new event object for the UI
            const newEvent = {
                event_id: response.data.EventId,
                event_title: title.trim(),
                event_body: body.trim(),
                group_id: groupId,
                event_date: eventDate,
                image_id: imageUploadFailed ? null : response.data.FileId,
                image_url: imageUrl,
                going_count: 0,
                not_going_count: 0,
                user_response: null,
                user: {
                    id: user.id,
                    username: user.username,
                    avatar_url: user.avatar_url,
                },
                created_at: new Date().toISOString(),
            };

            // Success - reset form and close
            setIsSubmitting(false);
            resetForm();
            onClose();

            // Show warning if image upload failed
            if (imageUploadFailed) {
                setWarning("Event image failed to upload. You can try again later.");
                setTimeout(() => setWarning(""), 3000);
            }

            if (onSuccess) {
                onSuccess(newEvent);
            }
        } catch (err) {
            console.error("Failed to create event:", err);
            setError("Failed to create event. Please try again.");
            setIsSubmitting(false);
        }
    };

    const resetForm = () => {
        setTitle("");
        setBody("");
        setEventDate("");
        setImageFile(null);
        setImagePreview(null);
        setError("");
        if (fileInputRef.current) {
            fileInputRef.current.value = "";
        }
    };

    const handleClose = () => {
        if (isSubmitting) return;
        resetForm();
        onClose();
    };

    // Get minimum datetime (now) for the date picker
    const getMinDateTime = () => {
        const now = new Date();
        now.setMinutes(now.getMinutes() - now.getTimezoneOffset());
        return now.toISOString().slice(0, 16);
    };

    const isValid = title.trim() && body.trim() && eventDate;

    return (
        <>
            {/* Warning Message - Fixed top banner (outside Modal so it persists) */}
            <AnimatePresence>
                {warning && (
                    <motion.div
                        initial={{ opacity: 0, y: -20 }}
                        animate={{ opacity: 1, y: 0 }}
                        exit={{ opacity: 0, y: -20 }}
                        className="fixed top-4 left-1/2 -translate-x-1/2 z-100 px-4 py-2 bg-amber-500/90 text-white text-sm rounded-lg shadow-lg"
                    >
                        {warning}
                    </motion.div>
                )}
            </AnimatePresence>

            <Modal
                isOpen={isOpen}
                onClose={handleClose}
                title="Create Event"
                description="Plan a group event for members to attend"
                showCloseButton={!isSubmitting}
            >
                <div className="space-y-4">
                {/* Title Input */}
                <div>
                    <label className="block text-sm font-medium text-foreground mb-2">
                        Event Title <span className="text-red-500">*</span>
                    </label>
                    <input
                        type="text"
                        value={title}
                        onChange={(e) => setTitle(e.target.value)}
                        placeholder="Enter event title"
                        disabled={isSubmitting}
                        className="w-full rounded-xl border border-(--muted)/30 px-4 py-2.5 text-sm bg-(--muted)/5 focus:outline-none focus:border-(--accent) focus:ring-2 focus:ring-(--accent)/10 transition-all disabled:opacity-50"
                        maxLength={MAX_TITLE_CHARS}
                    />
                    <div className="text-xs text-(--muted) mt-1 text-right pr-3">
                        {title.length}/{MAX_TITLE_CHARS}
                    </div>
                </div>

                {/* Description Input */}
                <div>
                    <label className="block text-sm font-medium text-foreground mb-2">
                        Description <span className="text-red-500">*</span>
                    </label>
                    <textarea
                        value={body}
                        onChange={(e) => setBody(e.target.value)}
                        placeholder="Describe your event..."
                        disabled={isSubmitting}
                        rows={3}
                        className="w-full rounded-xl border border-(--muted)/30 px-4 py-2.5 text-sm bg-(--muted)/5 focus:outline-none focus:border-(--accent) focus:ring-2 focus:ring-(--accent)/10 transition-all resize-none disabled:opacity-50"
                        maxLength={MAX_BODY_CHARS}
                    />
                    <div className="text-xs text-(--muted) text-right pr-3">
                        {body.length}/{MAX_BODY_CHARS}
                    </div>
                </div>

                {/* Date and Time Input */}
                <div>
                    <label className="block text-sm font-medium text-foreground mb-2">
                        <span className="flex items-center gap-2">
                            <Calendar className="w-4 h-4" />
                            Event Date & Time <span className="text-red-500">*</span>
                        </span>
                    </label>
                    <input
                        type="datetime-local"
                        value={eventDate}
                        onChange={(e) => setEventDate(e.target.value)}
                        min={getMinDateTime()}
                        disabled={isSubmitting}
                        className="w-full rounded-xl border border-(--muted)/30 px-4 py-2.5 text-sm bg-(--muted)/5 focus:outline-none focus:border-(--accent) focus:ring-2 focus:ring-(--accent)/10 transition-all disabled:opacity-50"
                    />
                </div>

                {/* Image Upload */}
                <div>
                    <label className="block text-sm font-medium text-foreground mb-2">
                        Event Image (Optional)
                    </label>

                    {imagePreview ? (
                        <div className="relative inline-block">
                            <img
                                src={imagePreview}
                                alt="Event preview"
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
                        disabled={isSubmitting || !isValid}
                        className="px-5 py-2 text-sm font-medium bg-(--accent) text-white hover:bg-(--accent-hover) rounded-full transition-colors disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed flex items-center gap-2"
                    >
                        {isSubmitting ? (
                            <>
                                <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                                Creating...
                            </>
                        ) : (
                            "Create Event"
                        )}
                    </button>
                </div>
            </div>
        </Modal>
        </>
    );
}
