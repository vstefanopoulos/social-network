"use client";

import { useState, useRef, useEffect, useCallback } from "react";
import { motion, AnimatePresence } from "motion/react";
import { X, Image as ImageIcon, ChevronDown, User, Check } from "lucide-react";
import Tooltip from "@/components/ui/Tooltip";
import { validateImage } from "@/lib/validation";
import { createPost } from "@/actions/posts/create-post";
import { validateUpload } from "@/actions/auth/validate-upload";
import { getFollowers } from "@/actions/users/get-followers";
import { useStore } from "@/store/store";

export default function CreatePost({ onPostCreated=null }) {
    const user = useStore((state) => state.user);
    const [content, setContent] = useState("");
    const [privacy, setPrivacy] = useState("everyone");
    const [isPrivacyOpen, setIsPrivacyOpen] = useState(false);
    const [selectedFollowers, setSelectedFollowers] = useState([]);
    const [imageFile, setImageFile] = useState(null);
    const [imagePreview, setImagePreview] = useState(null);
    const [error, setError] = useState("");
    const [warning, setWarning] = useState("");
    const [followers, setFollowers] = useState([]);
    const [followersOffset, setFollowersOffset] = useState(0);
    const [hasMoreFollowers, setHasMoreFollowers] = useState(true);
    const [isLoadingFollowers, setIsLoadingFollowers] = useState(false);
    const [followersFetched, setFollowersFetched] = useState(false);
    const fileInputRef = useRef(null);
    const dropdownRef = useRef(null);
    const followersListRef = useRef(null);

    const FOLLOWERS_LIMIT = 20;

    // Fetch initial followers
    const fetchInitialFollowers = async () => {
        if (!user?.id || followersFetched) return;

        setIsLoadingFollowers(true);
        const followersResult = await getFollowers({
            userId: user.id,
            limit: FOLLOWERS_LIMIT,
            offset: 0
        });
        const followersData = followersResult.success ? followersResult.data : [];

        setFollowers(followersData);
        setFollowersOffset(FOLLOWERS_LIMIT);
        setHasMoreFollowers(followersData && followersData.length === FOLLOWERS_LIMIT);
        setFollowersFetched(true);
        setIsLoadingFollowers(false);
    };

    // Load more followers
    const loadMoreFollowers = useCallback(async () => {
        if (!user?.id || isLoadingFollowers || !hasMoreFollowers) return;

        setIsLoadingFollowers(true);
        const followersResult = await getFollowers({
            userId: user.id,
            limit: FOLLOWERS_LIMIT,
            offset: followersOffset
        });
        const moreFollowers = followersResult.success ? followersResult.data : [];

        if (moreFollowers && moreFollowers.length > 0) {
            setFollowers(prev => [...prev, ...moreFollowers]);
            setFollowersOffset(prev => prev + FOLLOWERS_LIMIT);
            setHasMoreFollowers(moreFollowers.length === FOLLOWERS_LIMIT);
        } else {
            setHasMoreFollowers(false);
        }
        setIsLoadingFollowers(false);
    }, [user?.id, isLoadingFollowers, hasMoreFollowers, followersOffset]);

    // Handle scroll for infinite loading
    useEffect(() => {
        const handleScroll = () => {
            if (!followersListRef.current) return;

            const { scrollTop, scrollHeight, clientHeight } = followersListRef.current;
            // Trigger when scrolled to within 10px of bottom
            if (scrollHeight - scrollTop <= clientHeight + 10) {
                loadMoreFollowers();
            }
        };

        const listElement = followersListRef.current;
        if (listElement) {
            listElement.addEventListener('scroll', handleScroll);
            return () => listElement.removeEventListener('scroll', handleScroll);
        }
    }, [loadMoreFollowers]);

    // Close dropdown when clicking outside
    useEffect(() => {
        function handleClickOutside(event) {
            if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
                setIsPrivacyOpen(false);
            }
        }
        document.addEventListener("mousedown", handleClickOutside);
        return () => {
            document.removeEventListener("mousedown", handleClickOutside);
        };
    }, []);

    const MAX_CHARS = 5000;
    const MIN_CHARS = 1;

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

    const handlePrivacySelect = (newPrivacy) => {
        setPrivacy(newPrivacy);
        setIsPrivacyOpen(false);
        if (newPrivacy !== "selected") {
            setSelectedFollowers([]);
        } else {
            // Fetch followers only when "selected" is chosen
            fetchInitialFollowers();
        }
    };

    const toggleFollower = (followerId) => {
        // Ensure followerId is a string for consistent comparison
        const followerIdStr = String(followerId);
        setSelectedFollowers((prev) =>
            prev.includes(followerIdStr)
                ? prev.filter((id) => id !== followerIdStr)
                : [...prev, followerIdStr]
        );
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

        if (privacy === "selected" && selectedFollowers.length === 0) {
            setError("Please select at least one follower for selected posts");
            return;
        }

        try {
            // Prepare post data
            const postData = {
                post_body: content.trim(),
                audience: privacy,
            };

            // Add image metadata if image is selected
            if (imageFile) {
                postData.image_name = imageFile.name;
                postData.image_size = imageFile.size;
                postData.image_type = imageFile.type;
            }

            // Add audience IDs for selected posts
            if (privacy === "selected") {
                postData.audience_ids = selectedFollowers.map(id => id);
            }
            // Step 1: Create post with metadata
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
                    const url = `${resp.UploadUrl}hee`
                    const uploadRes = await fetch(url, {
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
                audience: privacy,
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
            setPrivacy("everyone");
            setSelectedFollowers([]);
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
        setPrivacy("everyone");
        setSelectedFollowers([]);
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
                        placeholder="What's on your mind?"
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
                            className="absolute -top-2 -right-2 bg-background text-(--muted) hover:text-red-500 rounded-full p-1.5 border border-(--border) shadow-sm transition-colors"
                        >
                            <X size={16} />
                        </button>
                    </div>
                )}

                {/* Error Message */}
                {error && (
                    <div className="text-red-500 text-sm bg-background border border-red-200 rounded-lg px-4 py-2.5 animate-fade-in">
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

                {/* Follower Multi-Select for Selected */}
                {privacy === "selected" && (
                    <div className="border border-(--border) rounded-xl p-4 space-y-2 bg-(--muted)/5 mt-3">
                        <p className="text-xs font-medium text-(--muted)">
                            Select followers who can see this post:
                        </p>
                        <div
                            ref={followersListRef}
                            className="space-y-1 max-h-48 overflow-y-auto custom-scrollbar"
                        >
                            {followers.length > 0 ? (
                                <>
                                    {followers.map((follower, index) => {
                                        const isSelected = selectedFollowers.includes(String(follower.id));
                                        return (
                                            <button
                                                key={follower.id || `follower-${index}`}
                                                type="button"
                                                onClick={() => toggleFollower(follower.id)}
                                                className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-xl transition-colors cursor-pointer ${
                                                    isSelected
                                                        ? "bg-(--accent)/10 border border-(--accent)/30"
                                                        : "hover:bg-(--muted)/10 border border-transparent"
                                                }`}
                                            >
                                                <div className="w-9 h-9 rounded-full bg-(--muted)/10 flex items-center justify-center overflow-hidden shrink-0">
                                                    {follower.avatar_url ? (
                                                        <img
                                                            src={follower.avatar_url}
                                                            alt={follower.username}
                                                            className="w-full h-full object-cover"
                                                        />
                                                    ) : (
                                                        <User className="w-4 h-4 text-(--muted)" />
                                                    )}
                                                </div>
                                                <div className="flex-1 min-w-0 text-left">
                                                    <p className="text-sm font-medium text-foreground truncate">
                                                        {follower.username}
                                                    </p>
                                                </div>
                                                <div className={`w-5 h-5 rounded-full border-2 flex items-center justify-center shrink-0 transition-colors ${
                                                    isSelected
                                                        ? "bg-(--accent) border-(--accent)"
                                                        : "border-(--border)"
                                                }`}>
                                                    {isSelected && <Check className="w-3 h-3 text-white" />}
                                                </div>
                                            </button>
                                        );
                                    })}
                                    {isLoadingFollowers && (
                                        <div className="flex justify-center py-2">
                                            <div className="w-4 h-4 border-2 border-(--accent) border-t-transparent rounded-full animate-spin"></div>
                                        </div>
                                    )}
                                    {!isLoadingFollowers && !hasMoreFollowers && followers.length > FOLLOWERS_LIMIT && (
                                        <p className="text-xs text-(--muted) text-center py-1">
                                            All followers loaded
                                        </p>
                                    )}
                                </>
                            ) : (
                                <p className="text-xs text-(--muted) text-center py-2">
                                    {isLoadingFollowers ? "Loading followers..." : "No followers to select"}
                                </p>
                            )}
                        </div>
                    </div>
                )}

                {/* Bottom Controls Row */}
                <div className="flex flex-wrap items-center justify-between gap-2 pt-2">
                    {/* Left side: Privacy and Image Upload */}
                    <div className="flex items-center gap-2">
                        {/* Privacy Selector */}
                        {/* Privacy Dropdown */}
                        <div className="relative" ref={dropdownRef}>
                            <Tooltip content={isPrivacyOpen ? "" : "Select privacy"}>
                                <button
                                    type="button"
                                    onClick={() => setIsPrivacyOpen(!isPrivacyOpen)}
                                    className="flex items-center gap-1.5 bg-(--muted)/5 border border-(--border) rounded-full px-3 py-1.5 text-sm text-foreground hover:border-foreground focus:border-(--accent) transition-colors cursor-pointer"
                                >
                                    <span className="capitalize">{privacy}</span>
                                    <ChevronDown size={14} className={`transition-transform duration-200 ${isPrivacyOpen ? "rotate-180" : ""}`} />
                                </button>
                            </Tooltip>
                            {isPrivacyOpen && (
                                <div className="absolute top-full left-0 mt-1 w-32 bg-background border border-(--border) rounded-xl z-50 animate-fade-in">

                                    <div className="flex flex-col p-1">
                                        <button
                                            type="button"
                                            onClick={() => handlePrivacySelect("everyone")}
                                            className={`w-full text-left px-3 py-1.5 text-sm rounded-lg transition-colors ${privacy === "everyone" ? "bg-(--muted)/10 font-medium" : "hover:bg-(--muted)/5 cursor-pointer"
                                                }`}
                                        >
                                            Everyone
                                        </button>


                                        <button
                                            type="button"
                                            onClick={() => handlePrivacySelect("followers")}
                                            className={`w-full text-left px-3 py-1.5 text-sm rounded-lg transition-colors ${privacy === "followers" ? "bg-(--muted)/10 font-medium" : "hover:bg-(--muted)/5 cursor-pointer"
                                                }`}
                                        >
                                            Followers
                                        </button>


                                        <button
                                            type="button"
                                            onClick={() => handlePrivacySelect("selected")}
                                            className={`w-full text-left px-3 py-1.5 text-sm rounded-lg transition-colors ${privacy === "selected" ? "bg-(--muted)/10 font-medium" : "hover:bg-(--muted)/5 cursor-pointer"
                                                }`}
                                        >
                                            Selected
                                        </button>

                                    </div>
                                </div>
                            )}
                        </div>

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
