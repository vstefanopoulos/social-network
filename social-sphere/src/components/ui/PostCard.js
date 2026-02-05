"use client";

import Link from "next/link";
import { useState, useEffect, useRef } from "react";
import { motion, AnimatePresence } from "motion/react";
import { Heart, MessageCircle, Pencil, Trash2, MoreHorizontal, Share2, Globe, Lock, Users, User, ChevronDown } from "lucide-react";
import PostImage from "./PostImage";
import Modal from "./Modal";
import { useStore } from "@/store/store";
import { editPost } from "@/actions/posts/edit-post";
import { deletePost } from "@/actions/posts/delete-post";
import { editComment } from "@/actions/posts/edit-comment";
import { deleteComment } from "@/actions/posts/delete-comment";
import { validateUpload } from "@/actions/auth/validate-upload";
import { validateImage } from "@/lib/validation";
import { getFollowers } from "@/actions/users/get-followers";
import { getPost } from "@/actions/posts/get-post";
import { getRelativeTime } from "@/lib/time";
import { toggleReaction } from "@/actions/posts/toggle-reaction";
import { getComments } from "@/actions/posts/get-comments";
import { createComment } from "@/actions/posts/create-comment";
import Tooltip from "./Tooltip";

export default function PostCard({ post, onDelete }) {
    const user = useStore((state) => state.user);
    const [image, setImage] = useState(post.image_url);
    const [comments, setComments] = useState([]);
    const [loading, setLoading] = useState(true);
    const [loadingMore, setLoadingMore] = useState(false);
    const [hasMore, setHasMore] = useState(true);
    const [isExpanded, setIsExpanded] = useState(false);
    const [draftComment, setDraftComment] = useState("");
    const [postContent, setPostContent] = useState(post.post_body ?? "");
    const [isEditingPost, setIsEditingPost] = useState(false);
    const [postDraft, setPostDraft] = useState(post.post_body ?? "");
    const [editingCommentId, setEditingCommentId] = useState(null);
    const [editingText, setEditingText] = useState("");
    const [error, setError] = useState("");
    const [warning, setWarning] = useState("");
    const [imageFile, setImageFile] = useState(null);
    const [imagePreview, setImagePreview] = useState(null);
    const [showDeleteModal, setShowDeleteModal] = useState(false);
    const [isDeleting, setIsDeleting] = useState(false);
    const [removeExistingImage, setRemoveExistingImage] = useState(false);
    const [privacy, setPrivacy] = useState(post.audience || "everyone");
    const [isPrivacyOpen, setIsPrivacyOpen] = useState(false);
    const [selectedFollowers, setSelectedFollowers] = useState([]);
    const [followers, setFollowers] = useState([]);
    const [isLoadingFollowers, setIsLoadingFollowers] = useState(false);
    const [relativeTime, setRelativeTime] = useState("");
    const [likedByUser, setLikedByUser] = useState(post.liked_by_user);
    const [reactionsCount, setReactionsCount] = useState(post.reactions_count);
    const [isReactionPending, setIsReactionPending] = useState(false);
    const [commentsCount, setCommentsCount] = useState(post.comments_count);
    const [commentImageFile, setCommentImageFile] = useState(null);
    const [commentImagePreview, setCommentImagePreview] = useState(null);
    const [editingCommentImageFile, setEditingCommentImageFile] = useState(null);
    const [editingCommentImagePreview, setEditingCommentImagePreview] = useState(null);
    const [errorEditComImage, setErrorEditComImage] = useState(null);
    const [errorCreateComImage, setErrorCreateComImage] = useState(null);
    const [removeCommentExistingImage, setRemoveCommentExistingImage] = useState(false);
    const [showDeleteCommentModal, setShowDeleteCommentModal] = useState(false);
    const [commentToDelete, setCommentToDelete] = useState(null);
    const [isDeletingComment, setIsDeletingComment] = useState(false);
    const [commentReactions, setCommentReactions] = useState({});
    const composerRef = useRef(null);
    const cardRef = useRef(null);
    const fileInputRef = useRef(null);
    const commentFileInputRef = useRef(null);
    const editingCommentFileInputRef = useRef(null);
    const dropdownRef = useRef(null);

    const isOwnPost = Boolean(
        user &&
        post?.post_user?.id &&
        String(post.post_user.id) === String(user.id)
    );

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

    // Update relative time on client side to avoid hydration mismatch
    useEffect(() => {
        setRelativeTime(getRelativeTime(post.created_at));

        // Optional: Update every minute to keep it current
        const interval = setInterval(() => {
            setRelativeTime(getRelativeTime(post.created_at));
        }, 60000); // Update every 60 seconds

        return () => clearInterval(interval);
    }, [post.created_at]);

    useEffect(() => {
        if (isExpanded && composerRef.current) {
            composerRef.current.focus();
        }
    }, [isExpanded]);

    // Fetch comments when expanded
    useEffect(() => {
        if (isExpanded && comments.length === 0) {
            fetchLastComment();
        }
    }, [isExpanded]);

    const fetchLastComment = async () => {
        setLoading(true);
        try {
            const result = await getComments({
                postId: post.post_id,
                limit: 1,
                offset: 0
            });

            if (result.success && result.data) {
                setComments(result.data);
                // Initialize comment reactions state
                const reactionsState = {};
                result.data.forEach(comment => {
                    reactionsState[comment.comment_id] = {
                        liked: comment.liked_by_user,
                        count: comment.reactions_count,
                        pending: false
                    };
                });
                setCommentReactions(reactionsState);
                // Check if there are more comments to load
                setHasMore(commentsCount > 1);
            }
        } catch (error) {
            console.error("Failed to fetch last comment:", error);
        } finally {
            setLoading(false);
        }
    };

    const handleLoadMore = async (e) => {
        e.preventDefault();
        e.stopPropagation();
        if (loadingMore) return;

        setLoadingMore(true);
        try {
            const result = await getComments({
                postId: post.post_id,
                limit: 3,
                offset: comments.length
            });

            if (result.success && result.data) {
                // Backend returns newest first, so reverse to get oldest first
                // Then prepend them above the existing comments (which has newest at bottom)
                const reversedComments = [...result.data].reverse();
                setComments((prev) => [...reversedComments, ...prev]);
                // Initialize comment reactions state for new comments
                const reactionsState = {};
                result.data.forEach(comment => {
                    reactionsState[comment.comment_id] = {
                        liked: comment.liked_by_user,
                        count: comment.reactions_count,
                        pending: false
                    };
                });
                setCommentReactions(prev => ({ ...prev, ...reactionsState }));
                // Check if there are still more to load
                setHasMore(comments.length + result.data.length < commentsCount);
            }
        } catch (error) {
            console.error("Failed to load more comments:", error);
        } finally {
            setLoadingMore(false);
        }
    };

    const fetchFollowers = async () => {
        if (!user?.id || isLoadingFollowers) return;

        setIsLoadingFollowers(true);
        const followersResult = await getFollowers({
            userId: user.id,
            limit: 100,
            offset: 0
        });
        setFollowers(followersResult.success ? followersResult.data : []);
        setIsLoadingFollowers(false);
    };

    const handlePrivacySelect = (newPrivacy) => {
        setPrivacy(newPrivacy);
        setIsPrivacyOpen(false);
        if (newPrivacy !== "selected") {
            setSelectedFollowers([]);
        } else {
            fetchFollowers();
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

    const handleStartEditPost = async (e) => {
        e.preventDefault();
        e.stopPropagation();

        // Fetch full post data to get selected_audience_users
        const fetchedPost = await getPost(post.post_id);
        if (!fetchedPost.success || !fetchedPost.data) {
            setError("Failed to load post data");
            return;
        }

        setPostDraft(postContent);
        const postPrivacy = fetchedPost.data.audience || "everyone";
        setPrivacy(postPrivacy);

        // If privacy is "selected", load the followers and set the selected ones
        if (postPrivacy === "selected") {
            await fetchFollowers();
            // Set selected followers from fetched post's selected_audience_users
            if (fetchedPost.data.selected_audience_users && Array.isArray(fetchedPost.data.selected_audience_users)) {
                // Ensure IDs are strings for consistent comparison
                setSelectedFollowers(fetchedPost.data.selected_audience_users.map(user => String(user.id)));
            }
        }

        setIsEditingPost(true);
        setError("");
        setRemoveExistingImage(false);
    };

    const handleCancelEditPost = (e) => {
        e.preventDefault();
        e.stopPropagation();
        setPostDraft(postContent);
        setPrivacy(post.audience || "everyone");
        setSelectedFollowers([]);
        setIsEditingPost(false);
        setImageFile(null);
        setImagePreview(null);
        setRemoveExistingImage(false);
        setError("");
        if (fileInputRef.current) {
            fileInputRef.current.value = "";
        }
    };

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

    const handleRemoveExistingImage = () => {
        setRemoveExistingImage(true);
    };

    const handleSaveEditPost = async (e) => {
        e.preventDefault();
        e.stopPropagation();
        if (!postDraft.trim()) return;

        // Validate selected privacy
        if (privacy === "selected" && selectedFollowers.length === 0) {
            setError("Please select at least one follower for selected posts");
            return;
        }

        try {
            setError("");

            const editData = {
                post_id: post.post_id,
                post_body: postDraft.trim(),
                audience: privacy,
                audience_ids: privacy === "selected" ? selectedFollowers : []
            };

            // Handle new image upload
            if (imageFile) {
                editData.image_name = imageFile.name;
                editData.image_size = imageFile.size;
                editData.image_type = imageFile.type;
            }
            // Handle explicit image removal
            else if (removeExistingImage) {
                editData.delete_image = true;
            }

            const resp = await editPost(editData);

            if (!resp.success) {
                setError(resp.error || "Failed to edit post");
                return;
            }

            if (imageFile && resp.data?.FileId && resp.data?.UploadUrl) {
                let imageUploadFailed = false;
                try {
                    const uploadRes = await fetch(resp.data.UploadUrl, {
                        method: "PUT",
                        body: imageFile,
                    });

                    if (uploadRes.ok) {
                        const validateResp = await validateUpload(resp.data.FileId);
                        if (validateResp.success) {
                            setImage(validateResp.data?.download_url);
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

                if (imageUploadFailed) {
                    setWarning("Image failed to upload. You can try again later.");
                    setTimeout(() => setWarning(""), 3000);
                }
            } else if (removeExistingImage) {
                // User removed the existing image
                setImage(null);
                setImagePreview(null);
            }

            setPostContent(postDraft);
            setIsEditingPost(false);
            // window.location.reload();

        } catch (err) {
            console.error("Failed to edit post:", err);
            setError("Failed to edit post. Please try again.");
        }
    };

    const handleDeleteClick = (e) => {
        e.preventDefault();
        e.stopPropagation();
        setShowDeleteModal(true);
    };

    const handleDeleteConfirm = async () => {
        setIsDeleting(true);
        setError("");

        try {
            const resp = await deletePost(post.post_id);

            if (!resp.success) {
                setError(resp.error || "Failed to delete post");
                setIsDeleting(false);
                setShowDeleteModal(false);
                return;
            }

            // Successfully deleted
            setShowDeleteModal(false);
            setIsDeleting(false);

            // If onDelete callback is provided (from feed), use it for smooth animation
            // Otherwise reload the page (for single post page)
            if (onDelete) {
                onDelete(post.post_id);
            } else {
                window.location.reload();
            }

        } catch (err) {
            console.error("Failed to delete post:", err);
            setError("Failed to delete post. Please try again.");
            setIsDeleting(false);
            setShowDeleteModal(false);
        }
    };

    const handleHeartClick = async (e) => {
        e.preventDefault();
        e.stopPropagation();

        if (isReactionPending) return;

        // Optimistic update
        const previousLiked = likedByUser;
        const previousCount = reactionsCount;

        setLikedByUser(!likedByUser);
        setReactionsCount(likedByUser ? reactionsCount - 1 : reactionsCount + 1);
        setIsReactionPending(true);

        try {
            const resp = await toggleReaction(post.post_id);

            if (!resp.success) {
                // Revert on error
                setLikedByUser(previousLiked);
                setReactionsCount(previousCount);
                console.error("Failed to toggle reaction:", resp.error);
            }
        } catch (err) {
            // Revert on error
            setLikedByUser(previousLiked);
            setReactionsCount(previousCount);
            console.error("Failed to toggle reactions:", err);
        } finally {
            setIsReactionPending(false);
        }
    };

    const handleToggleComments = (e) => {
        e.preventDefault();
        e.stopPropagation();
        setIsExpanded(!isExpanded);
    };

    const handleSubmitComment = async (e) => {
        e.preventDefault();
        e.stopPropagation();

        if (!draftComment.trim()) return;

        try {
            const commentData = {
                postId: post.post_id,
                commentBody: draftComment.trim()
            };

            // Handle image upload if present
            if (commentImageFile) {
                commentData.imageName = commentImageFile.name;
                commentData.imageSize = commentImageFile.size;
                commentData.imageType = commentImageFile.type;
            }

            const resp = await createComment(commentData);

            if (!resp.success) {
                setError(resp.error || "Failed to create comment");
                return;
            }

            // If there's an image, upload it (non-blocking)
            let commentImageUploadFailed = false;
            if (commentImageFile && resp.data?.FileId && resp.data?.UploadUrl) {
                try {
                    const uploadRes = await fetch(resp.data.UploadUrl, {
                        method: "PUT",
                        body: commentImageFile,
                    });

                    if (uploadRes.ok) {
                        const validateResp = await validateUpload(resp.data.FileId);
                        if (!validateResp.success) {
                            commentImageUploadFailed = true;
                        }
                    } else {
                        commentImageUploadFailed = true;
                    }
                } catch (uploadErr) {
                    console.error("Comment image upload failed:", uploadErr);
                    commentImageUploadFailed = true;
                }
            }

            // Success! Clear form and refresh
            setDraftComment("");
            setCommentImageFile(null);
            setCommentImagePreview(null);
            if (commentFileInputRef.current) {
                commentFileInputRef.current.value = "";
            }
            // Increment comment count
            setCommentsCount((prev) => prev + 1);
            // Refresh to show the new comment
            setError("");
            setComments([]);
            fetchLastComment();

            // Show warning if image upload failed
            if (commentImageUploadFailed) {
                setWarning("Comment image failed to upload. You can try again later.");
                setTimeout(() => setWarning(""), 3000);
            }
        } catch (err) {
            console.error("Failed to create comment:", err);
            setError("Failed to create comment. Please try again.");
        }
    };

    const handleCancelComposer = (e) => {
        e.preventDefault();
        e.stopPropagation();
        setDraftComment("");
        setErrorCreateComImage(null);
        setCommentImageFile(null);
        setCommentImagePreview(null);
        if (commentFileInputRef.current) {
            commentFileInputRef.current.value = "";
        }
    };

    const handleCommentImageSelect = async (e) => {
        const file = e.target.files?.[0];
        if (!file) return;

        // Validate image file (type, size, dimensions)
        const validation = await validateImage(file);
        if (!validation.valid) {
            setErrorCreateComImage(validation.error);
            return;
        }

        setCommentImageFile(file);
        setErrorCreateComImage(null);

        const reader = new FileReader();
        reader.onloadend = () => {
            setCommentImagePreview(reader.result);
        };
        reader.readAsDataURL(file);
    };

    const handleRemoveCommentImage = (e) => {
        e.preventDefault();
        e.stopPropagation();
        setCommentImageFile(null);
        setCommentImagePreview(null);
        if (commentFileInputRef.current) {
            commentFileInputRef.current.value = "";
        }
    };

    const handleStartEditComment = (comment) => {
        setEditingCommentId(comment.comment_id);
        setEditingText(comment.comment_body);
        setEditingCommentImageFile(null);
        setEditingCommentImagePreview(null);
        setRemoveCommentExistingImage(false);
        setError("");
    };

    const handleCancelEditComment = () => {
        setEditingCommentId(null);
        setEditingText("");
        setEditingCommentImageFile(null);
        setEditingCommentImagePreview(null);
        setRemoveCommentExistingImage(false);
        setError("");
        if (editingCommentFileInputRef.current) {
            editingCommentFileInputRef.current.value = "";
        }
    };

    const handleEditingCommentImageSelect = async (e) => {
        const file = e.target.files?.[0];
        if (!file) return;

        // Validate image file (type, size, dimensions)
        const validation = await validateImage(file);
        if (!validation.valid) {
            setErrorEditComImage(validation.error);
            return;
        }

        setEditingCommentImageFile(file);
        setErrorEditComImage(null);

        const reader = new FileReader();
        reader.onloadend = () => {
            setEditingCommentImagePreview(reader.result);
        };
        reader.readAsDataURL(file);
    };

    const handleRemoveEditingCommentImage = (e) => {
        e.preventDefault();
        e.stopPropagation();
        setEditingCommentImageFile(null);
        setEditingCommentImagePreview(null);
        if (editingCommentFileInputRef.current) {
            editingCommentFileInputRef.current.value = "";
        }
    };

    const handleRemoveCommentExistingImage = (e) => {
        e.preventDefault();
        e.stopPropagation();
        setRemoveCommentExistingImage(true);
    };

    const handleSaveEditComment = async (comment) => {
        if (!editingText.trim()) {
            handleCancelEditComment();
            return;
        }

        try {
            setError("");

            const editData = {
                comment_id: comment.comment_id,
                comment_body: editingText.trim()
            };

            // Handle new image upload
            if (editingCommentImageFile) {
                editData.image_name = editingCommentImageFile.name;
                editData.image_size = editingCommentImageFile.size;
                editData.image_type = editingCommentImageFile.type;
            }
            // Handle explicit image removal
            else if (removeCommentExistingImage) {
                editData.delete_image = true;
            }

            const resp = await editComment(editData);

            if (!resp.success) {
                setError("Failed to edit comment");
                return;
            }

            // If there's a new image, upload it (non-blocking)
            let editCommentImageUploadFailed = false;
            if (editingCommentImageFile && resp.data?.FileId && resp.data?.UploadUrl) {
                try {
                    const uploadRes = await fetch(resp.data.UploadUrl, {
                        method: "PUT",
                        body: editingCommentImageFile,
                    });

                    if (uploadRes.ok) {
                        const validateResp = await validateUpload(resp.data.FileId);
                        if (!validateResp.success) {
                            editCommentImageUploadFailed = true;
                        }
                    } else {
                        editCommentImageUploadFailed = true;
                    }
                } catch (uploadErr) {
                    console.error("Edit comment image upload failed:", uploadErr);
                    editCommentImageUploadFailed = true;
                }

                if (editCommentImageUploadFailed) {
                    setWarning("Comment image failed to upload. You can try again later.");
                    setTimeout(() => setWarning(""), 3000);
                }
            }

            // Update the comment in local state
            setComments((prev) =>
                prev.map((c) =>
                    c.comment_id === comment.comment_id
                        ? { ...c, comment_body: editingText.trim() }
                        : c
                )
            );
            handleCancelEditComment();

            // Refresh to show updated image if changed
            if (editingCommentImageFile || removeCommentExistingImage) {
                setComments([]);
                fetchLastComment();
            }
        } catch (err) {
            console.error("Failed to edit comment:", err);
            setError("Failed to edit comment. Please try again.");
        }
    };

    const handleDeleteCommentClick = (comment, e) => {
        e.preventDefault();
        e.stopPropagation();
        setCommentToDelete(comment);
        setShowDeleteCommentModal(true);
    };

    const handleDeleteCommentConfirm = async () => {
        if (!commentToDelete) return;

        setIsDeletingComment(true);
        setError("");

        try {
            const resp = await deleteComment(commentToDelete.comment_id);

            if (!resp.success) {
                setError(resp.error || "Failed to delete comment");
                setIsDeletingComment(false);
                setShowDeleteCommentModal(false);
                return;
            }

            // Remove comment from local state
            setComments((prev) => prev.filter((c) => c.comment_id !== commentToDelete.comment_id));
            setCommentsCount((prev) => prev - 1);
            setShowDeleteCommentModal(false);
            setCommentToDelete(null);
        } catch (err) {
            console.error("Failed to delete comment:", err);
            setError("Failed to delete comment. Please try again.");
        } finally {
            setIsDeletingComment(false);
        }
    };

    const handleCommentReactionClick = async (commentId, e) => {
        e.preventDefault();
        e.stopPropagation();

        const currentReaction = commentReactions[commentId];
        if (!currentReaction || currentReaction.pending) return;

        // Optimistic update
        const previousLiked = currentReaction.liked;
        const previousCount = currentReaction.count;

        setCommentReactions(prev => ({
            ...prev,
            [commentId]: {
                liked: !previousLiked,
                count: previousLiked ? previousCount - 1 : previousCount + 1,
                pending: true
            }
        }));

        try {
            const resp = await toggleReaction(commentId);

            if (!resp.success) {
                // Revert on error
                setCommentReactions(prev => ({
                    ...prev,
                    [commentId]: {
                        liked: previousLiked,
                        count: previousCount,
                        pending: false
                    }
                }));
                console.error("Failed to toggle comment reaction:", resp.error);
            } else {
                // Update pending state
                setCommentReactions(prev => ({
                    ...prev,
                    [commentId]: {
                        ...prev[commentId],
                        pending: false
                    }
                }));
            }
        } catch (err) {
            // Revert on error
            setCommentReactions(prev => ({
                ...prev,
                [commentId]: {
                    liked: previousLiked,
                    count: previousCount,
                    pending: false
                }
            }));
            console.error("Failed to toggle comment reaction:", err);
        }
    };

    return (
        <div
            ref={cardRef}
            className="group bg-background border border-(--border) rounded-2xl transition-all hover:border-(--muted)/40 hover:shadow-sm mb-6"
        >
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

            {/* Header */}
            <div className="p-5 flex items-start justify-between">
                <div className="flex items-center gap-3">
                    <Link href={`/profile/${post.post_user.id}`} prefetch={false} className="w-10 h-10 rounded-full bg-(--muted)/10 flex items-center justify-center overflow-hidden shrink-0">
                        {post.post_user.avatar_url ? (<div className="w-10 h-10 rounded-full overflow-hidden border border-(--border)">
                            <img
                                src={post.post_user.avatar_url}
                                alt="Avatar"
                                className="w-full h-full object-cover"
                            />
                        </div>) : (
                            <User className="w-5 h-5 text-(--muted)" />)}
                    </Link>
                    <div>
                        <Link href={`/profile/${post.post_user.id}`} prefetch={false}>
                            <h3 className="font-semibold text-foreground hover:underline decoration-2 underline-offset-2">
                                @{post.post_user.username}
                            </h3>
                        </Link>
                        <div className="flex items-center gap-2 text-xs text-(--muted) mt-0.5">
                            <span>{relativeTime || "..."}</span>
                        </div>
                    </div>
                </div>

                {isOwnPost && !isEditingPost && (
                    <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                        <Tooltip content="Edit Post">
                            <button
                                onClick={handleStartEditPost}
                                className="p-2 text-(--muted) hover:text-(--accent) hover:bg-(--accent)/5 rounded-full transition-colors cursor-pointer"
                            >
                                <Pencil className="w-4 h-4" />
                            </button>
                        </Tooltip>

                        <Tooltip content="Delete Post">
                            <button
                                onClick={handleDeleteClick}
                                className="p-2 text-(--muted) hover:text-red-500 hover:bg-red-500/5 rounded-full transition-colors cursor-pointer"
                            >
                                <Trash2 className="w-4 h-4" />
                            </button>
                        </Tooltip>
                    </div>
                )}
            </div>

            {/* Content */}
            <div className="px-5 pb-3">
                {isEditingPost ? (
                    <div className="space-y-3 mb-4">
                        <textarea
                            className="w-full rounded-xl border border-(--muted)/30 px-4 py-3 text-sm bg-(--muted)/5 focus:outline-none focus:border-(--accent) focus:ring-2 focus:ring-(--accent)/10 transition-all resize-none"
                            rows={4}
                            value={postDraft}
                            onChange={(e) => setPostDraft(e.target.value)}
                        />

                        {/* Privacy Selector */}
                        <div className="relative" ref={dropdownRef}>
                            <button
                                type="button"
                                onClick={() => setIsPrivacyOpen(!isPrivacyOpen)}
                                className="flex items-center gap-1.5 bg-(--muted)/5 border border-(--border) rounded-full px-3 py-1.5 text-sm text-foreground hover:border-foreground focus:border-(--accent) transition-colors cursor-pointer"
                            >
                                <span className="capitalize">{privacy}</span>
                                <ChevronDown size={14} className={`transition-transform duration-200 ${isPrivacyOpen ? "rotate-180" : ""}`} />
                            </button>
                            {isPrivacyOpen && (
                                <div className="absolute top-full left-0 mt-1 w-32 bg-background border border-(--border) rounded-xl z-50 shadow-lg">
                                    <div className="flex flex-col p-1">
                                        <button
                                            type="button"
                                            onClick={() => handlePrivacySelect("everyone")}
                                            className={`w-full text-left px-3 py-1.5 text-sm rounded-lg transition-colors ${privacy === "everyone" ? "bg-(--muted)/10 font-medium" : "hover:bg-(--muted)/5 cursor-pointer"}`}
                                        >
                                            Everyone
                                        </button>
                                        <button
                                            type="button"
                                            onClick={() => handlePrivacySelect("followers")}
                                            className={`w-full text-left px-3 py-1.5 text-sm rounded-lg transition-colors ${privacy === "followers" ? "bg-(--muted)/10 font-medium" : "hover:bg-(--muted)/5 cursor-pointer"}`}
                                        >
                                            Followers
                                        </button>
                                        <button
                                            type="button"
                                            onClick={() => handlePrivacySelect("selected")}
                                            className={`w-full text-left px-3 py-1.5 text-sm rounded-lg transition-colors ${privacy === "selected" ? "bg-(--muted)/10 font-medium" : "hover:bg-(--muted)/5 cursor-pointer"}`}
                                        >
                                            Selected
                                        </button>
                                    </div>
                                </div>
                            )}
                        </div>

                        {/* Follower Selection for "Selected" Privacy */}
                        {privacy === "selected" && (
                            <div className="border border-(--border) rounded-xl p-4 space-y-2 bg-(--muted)/5">
                                <p className="text-xs font-medium text-(--muted)">
                                    Select followers who can see this post:
                                </p>
                                <div className="space-y-1.5 max-h-32 overflow-y-auto">
                                    {followers.length > 0 ? (
                                        followers.map((follower, index) => (
                                            <label
                                                key={follower.id || `follower-${index}`}
                                                className="flex items-center gap-2 cursor-pointer hover:bg-(--muted)/10 rounded-lg px-2 py-1.5 transition-colors"
                                            >
                                                <input
                                                    type="checkbox"
                                                    checked={selectedFollowers.includes(String(follower.id))}
                                                    onChange={() => toggleFollower(follower.id)}
                                                    className="rounded border-gray-300"
                                                />
                                                <span className="text-sm">
                                                    @{follower.username}
                                                </span>
                                            </label>
                                        ))
                                    ) : (
                                        <p className="text-xs text-(--muted) text-center py-2">
                                            {isLoadingFollowers ? "Loading followers..." : "No followers to select"}
                                        </p>
                                    )}
                                </div>
                            </div>
                        )}

                        {/* Image Preview for Edit - New Image */}
                        {imagePreview && (
                            <div className="relative inline-block">
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
                                    <Trash2 className="w-4 h-4" />
                                </button>
                            </div>
                        )}

                        {/* Existing Image in Edit Mode */}
                        {!imagePreview && image && !removeExistingImage && (
                            <div className="relative inline-block">
                                <img
                                    src={image}
                                    alt="Post image"
                                    className="max-w-full max-h-64 rounded-xl border border-(--border)"
                                />
                                <button
                                    type="button"
                                    onClick={handleRemoveExistingImage}
                                    className="absolute -top-2 -right-2 bg-background text-(--muted) hover:text-red-500 rounded-full p-1.5 border border-(--border) shadow-sm transition-colors cursor-pointer"
                                >
                                    <Trash2 className="w-4 h-4" />
                                </button>
                            </div>
                        )}

                        {/* Error Message */}
                        {error && (
                            <div className="text-red-500 text-sm bg-background rounded-lg px-4 py-2.5">
                                {error}
                            </div>
                        )}

                        <div className="flex items-center justify-between gap-2">
                            <input
                                ref={fileInputRef}
                                type="file"
                                accept="image/jpeg,image/png,image/gif,image/webp"
                                onChange={handleImageSelect}
                                className="hidden"
                            />
                            <button
                                type="button"
                                onClick={() => fileInputRef.current?.click()}
                                className="px-3 py-1.5 text-xs font-medium text-(--muted) hover:text-foreground hover:bg-(--muted)/10 rounded-full transition-colors cursor-pointer"
                            >
                                {image ? "change image" : "Upload Image"}
                            </button>

                            <div className="flex items-center gap-2">
                                <button
                                    type="button"
                                    className="px-3 py-1.5 text-xs font-medium text-(--muted) hover:text-foreground hover:bg-(--muted)/10 rounded-full transition-colors cursor-pointer"
                                    onClick={handleCancelEditPost}
                                >
                                    Cancel
                                </button>
                                <button
                                    type="button"
                                    className="px-4 py-1.5 text-xs font-medium bg-(--accent) text-white hover:bg-(--accent-hover) rounded-full transition-colors disabled:opacity-50 cursor-pointer"
                                    disabled={!postDraft.trim()}
                                    onClick={handleSaveEditPost}
                                >
                                    Save Changes
                                </button>
                            </div>
                        </div>
                    </div>
                ) : (
                    <Link href={`/posts/${post.post_id}`} prefetch={false}>
                        <p className="text-[15px] leading-relaxed text-(--foreground)/90 whitespace-pre-wrap">
                            {postContent}
                        </p>
                    </Link>
                )}

                {/* Error Message (for non-edit operations like delete) */}
                {error && !isEditingPost && (
                    <div className="px-5 pb-3">
                        <div className="text-red-500 text-sm bg-background rounded-lg px-4 py-2.5">
                            {error}
                        </div>
                    </div>
                )}
            </div>


            {image && !isEditingPost && (
                <PostImage src={image} alt="He" />
            )}

            {/* Actions Footer */}
            {!isEditingPost && (
                <div className="px-5 py-4">
                    <div className="flex items-center justify-between">
                        <div className="flex items-center gap-6">
                            <button
                                onClick={handleHeartClick}
                                disabled={isReactionPending}
                                className="flex items-center gap-2 text-(--muted) hover:text-red-500 cursor-pointer transition-colors group/heart disabled:opacity-50"
                            >
                                <Heart className={`w-5 h-5 transition-transform group-hover/heart:scale-110  ${likedByUser ? "fill-red-500 text-red-500" : ""}`} />
                                <span className="text-sm font-medium">{reactionsCount}</span>
                            </button>

                            <button
                                onClick={handleToggleComments}
                                className={`flex items-center gap-2 transition-colors group/comment cursor-pointer ${isExpanded ? "text-(--accent)" : "text-(--muted) hover:text-(--accent)"}`}
                            >
                                <MessageCircle className={`w-5 h-5 transition-transform group-hover/comment:scale-110 ${isExpanded ? "fill-(--accent)/10" : ""}`} />
                                <span className="text-sm font-medium">{commentsCount}</span>
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* Expanded Section: Comments + Composer */}
            {isExpanded && (
                <div className="animate-in fade-in slide-in-from-top-2 duration-200">
                    {/* Comments List */}
                    {loading ? (
                        <div className="bg-(--muted)/5 border-t border-(--border) px-5 py-8 text-center">
                            <p className="text-sm text-(--muted)">Loading comments...</p>
                        </div>
                    ) : comments.length > 0 ? (
                        <div className="bg-(--muted)/5 border-t border-(--border) px-5 py-4">
                            {/* Load More Button */}
                            {hasMore && (
                                <button
                                    onClick={handleLoadMore}
                                    disabled={loadingMore}
                                    className="w-full text-left text-xs font-medium text-(--muted) hover:text-(--accent) mb-4 pl-11 transition-colors disabled:opacity-50 cursor-pointer"
                                >
                                    {loadingMore ? "Loading..." : "Load more comments"}
                                </button>
                            )}

                            <div className="flex flex-col gap-4">
                                {comments.map((comment) => {
                                    const isOwner = user && String(comment.user?.id) === String(user.id);
                                    const isEditing = editingCommentId === comment.comment_id;
                                    return (
                                        <div key={comment.comment_id} className="flex gap-3 group/comment-item">
                                            <Link href={`/profile/${comment.user.id}`} prefetch={false} className="shrink-0">
                                                <div className="w-8 h-8 rounded-full overflow-hidden border border-(--border) bg-(--muted)/10 flex items-center justify-center">
                                                    {comment.user.avatar_url ? (
                                                        <img src={comment.user.avatar_url} alt={comment.user.username} className="w-full h-full object-cover" />
                                                    ) : (
                                                        <User className="w-4 h-4 text-(--muted)" />
                                                    )}
                                                </div>
                                            </Link>
                                            <div className="flex-1 min-w-0">
                                                {isEditing ? (
                                                    <div className="space-y-2">
                                                        <textarea
                                                            className="w-full rounded-xl border border-(--muted)/30 px-4 py-3 text-sm bg-(--muted)/5 focus:outline-none focus:border-(--accent) focus:ring-2 focus:ring-(--accent)/10 transition-all resize-none"
                                                            rows={3}
                                                            value={editingText}
                                                            onChange={(e) => setEditingText(e.target.value)}
                                                        />

                                                        {/* Image Preview for Edit - New Image */}
                                                        {editingCommentImagePreview && (
                                                            <div className="relative inline-block">
                                                                <img
                                                                    src={editingCommentImagePreview}
                                                                    alt="Upload preview"
                                                                    className="max-w-full max-h-48 rounded-xl border border-(--border)"
                                                                />
                                                                <button
                                                                    type="button"
                                                                    onClick={handleRemoveEditingCommentImage}
                                                                    className="absolute -top-2 -right-2 bg-background text-(--muted) hover:text-red-500 rounded-full p-1.5 border border-(--border) shadow-sm transition-colors cursor-pointer"
                                                                >
                                                                    <Trash2 className="w-3 h-3" />
                                                                </button>
                                                            </div>
                                                        )}

                                                        {/* Existing Image in Edit Mode */}
                                                        {!editingCommentImagePreview && comment?.image_url && !removeCommentExistingImage && (
                                                            <div className="relative inline-block">
                                                                <img
                                                                    src={comment.image_url}
                                                                    alt="Comment image"
                                                                    className="max-w-full max-h-48 rounded-xl border border-(--border)"
                                                                />
                                                                <button
                                                                    type="button"
                                                                    onClick={handleRemoveCommentExistingImage}
                                                                    className="absolute -top-2 -right-2 bg-background text-(--muted) hover:text-red-500 rounded-full p-1.5 border border-(--border) shadow-sm transition-colors cursor-pointer"
                                                                >
                                                                    <Trash2 className="w-3 h-3" />
                                                                </button>
                                                            </div>
                                                        )}

                                                        <div className="flex items-center justify-between gap-2">
                                                            <input
                                                                ref={editingCommentFileInputRef}
                                                                type="file"
                                                                accept="image/jpeg,image/png,image/gif,image/webp"
                                                                onChange={handleEditingCommentImageSelect}
                                                                className="hidden"
                                                            />
                                                            <button
                                                                type="button"
                                                                onClick={() => editingCommentFileInputRef.current?.click()}
                                                                className="px-3 py-1.5 text-xs font-medium text-(--muted) hover:text-foreground hover:bg-(--muted)/10 rounded-full transition-colors cursor-pointer"
                                                            >
                                                                {editingCommentImagePreview || comment.image_url ? "Change Image" : "Add Image"}
                                                            </button>

                                                            {errorEditComImage ? (
                                                                <span className="text-sm text-red-500">{errorEditComImage}</span>
                                                            ) : <></>}

                                                            <div className="flex items-center gap-2">
                                                                <button
                                                                    type="button"
                                                                    className="px-3 py-1.5 text-xs font-medium text-(--muted) hover:text-foreground hover:bg-(--muted)/10 rounded-full transition-colors cursor-pointer"
                                                                    onClick={handleCancelEditComment}
                                                                >
                                                                    Cancel
                                                                </button>
                                                                <button
                                                                    type="button"
                                                                    className="px-4 py-1.5 text-xs font-medium bg-(--accent) text-white hover:bg-(--accent-hover) rounded-full transition-colors disabled:opacity-50 cursor-pointer"
                                                                    disabled={!editingText.trim()}
                                                                    onClick={() => handleSaveEditComment(comment)}
                                                                >
                                                                    Save
                                                                </button>
                                                            </div>
                                                        </div>
                                                    </div>
                                                ) : (
                                                    <div className="bg-background rounded-2xl rounded-tl-none px-4 py-2 border border-(--border) relative">
                                                        <div className="flex items-center justify-between mb-1">
                                                            <Link href={`/profile/${comment.user.id}`} prefetch={false}>
                                                                <span className="text-xs font-bold hover:underline">@{comment.user.username}</span>
                                                            </Link>
                                                            <div className="flex items-center gap-1">
                                                                <span className="text-[10px] text-(--muted)">{getRelativeTime(comment.created_at)}</span>
                                                                {isOwner && (
                                                                    <div className="flex items-center gap-0.5 opacity-0 group-hover/comment-item:opacity-100 transition-opacity ml-2">
                                                                        <Tooltip content="Edit Comment">
                                                                            <button
                                                                                onClick={() => handleStartEditComment(comment)}
                                                                                className="p-1 text-(--muted) hover:text-(--accent) hover:bg-(--accent)/5 rounded-full transition-colors cursor-pointer"
                                                                            >
                                                                                <Pencil className="w-3 h-3" />
                                                                            </button>
                                                                        </Tooltip>
                                                                        <Tooltip content="Delete Comment">
                                                                            <button
                                                                                onClick={(e) => handleDeleteCommentClick(comment, e)}
                                                                                className="p-1 text-(--muted) hover:text-red-500 hover:bg-red-500/5 rounded-full transition-colors cursor-pointer"
                                                                            >
                                                                                <Trash2 className="w-3 h-3" />
                                                                            </button>
                                                                        </Tooltip>
                                                                    </div>
                                                                )}
                                                            </div>
                                                        </div>
                                                        <p className="text-sm text-(--foreground)/90 leading-relaxed whitespace-pre-wrap">
                                                            {comment.comment_body}
                                                        </p>
                                                        {/* Comment Image */}
                                                        {comment.image_url && (
                                                            <div className="mt-2">
                                                                <img
                                                                    src={comment.image_url}
                                                                    alt="Comment attachment"
                                                                    className="max-w-full max-h-48 rounded-lg border border-(--border) object-cover"
                                                                />
                                                            </div>
                                                        )}
                                                        {/* Comment Reactions */}
                                                        {commentReactions[comment.comment_id] && (
                                                            <div className="mt-2">
                                                                <button
                                                                    onClick={(e) => handleCommentReactionClick(comment.comment_id, e)}
                                                                    disabled={commentReactions[comment.comment_id].pending}
                                                                    className="flex items-center gap-1 text-(--muted) hover:text-red-500 transition-colors group/comment-heart disabled:opacity-50"
                                                                >
                                                                    <Heart className={`w-4 h-4 transition-transform group-hover/comment-heart:scale-110 cursor-pointer ${commentReactions[comment.comment_id].liked ? "fill-red-500 text-red-500" : ""}`} />
                                                                    <span className="text-xs font-medium">{commentReactions[comment.comment_id].count}</span>
                                                                </button>
                                                            </div>
                                                        )}
                                                    </div>
                                                )}
                                            </div>
                                        </div>
                                    );
                                })}
                            </div>
                        </div>
                    ) : (
                        <div className="bg-(--muted)/5 border-t border-(--border) px-5 py-8 text-center">
                            <p className="text-sm text-(--muted)">No comments yet. Be the first to comment!</p>
                        </div>
                    )}

                    {/* Comment Composer */}
                    <div className="border-t border-(--border) p-4 bg-(--muted)/5">
                        <div className="flex gap-3">
                            <div className="w-8 h-8 rounded-full overflow-hidden bg-(--muted)/10 shrink-0 flex items-center justify-center">
                                {user?.avatar_url ? (
                                    <img src={user.avatar_url} alt="My Avatar" className="w-full h-full object-cover" />
                                ) : (
                                    <User className="w-4 h-4 text-(--muted)" />
                                )}
                            </div>
                            <div className="flex-1 space-y-2">
                                <textarea
                                    ref={composerRef}
                                    value={draftComment}
                                    onChange={(e) => setDraftComment(e.target.value)}
                                    rows={1}
                                    className="w-full rounded-2xl border border-(--muted)/30 px-4 py-2.5 text-sm bg-transparent focus:outline-none focus:border-(--accent) focus:ring-2 focus:ring-(--accent)/10 transition-all resize-none min-h-[42px]"
                                    placeholder="Write a comment..."
                                />

                                {/* Image Preview */}
                                {commentImagePreview && (
                                    <div className="relative inline-block">
                                        <img
                                            src={commentImagePreview}
                                            alt="Comment image preview"
                                            className="max-w-full max-h-32 rounded-xl border border-(--border)"
                                        />
                                        <button
                                            type="button"
                                            onClick={handleRemoveCommentImage}
                                            className="absolute -top-2 -right-2 bg-background text-(--muted) hover:text-red-500 rounded-full p-1.5 border border-(--border) shadow-sm transition-colors cursor-pointer"
                                        >
                                            <Trash2 className="w-3 h-3" />
                                        </button>
                                    </div>
                                )}

                                <div className="flex items-center justify-between gap-2">
                                    <input
                                        ref={commentFileInputRef}
                                        type="file"
                                        accept="image/jpeg,image/png,image/gif,image/webp"
                                        onChange={handleCommentImageSelect}
                                        className="hidden"
                                    />
                                    <button
                                        type="button"
                                        onClick={() => commentFileInputRef.current?.click()}
                                        className="px-3 py-1.5 text-xs font-medium text-(--muted) hover:text-foreground hover:bg-(--muted)/10 rounded-full transition-colors cursor-pointer"
                                    >
                                        {commentImageFile ? "Change Image" : "Add Image"}
                                    </button>

                                    {errorCreateComImage ? (
                                        <span className="text-sm text-red-500">{errorCreateComImage}</span>
                                    ) : <></>}

                                    <div className="flex gap-2">
                                        <button
                                            type="button"
                                            onClick={handleCancelComposer}
                                            className="px-3 py-1.5 text-xs font-medium text-(--muted) hover:text-foreground hover:bg-(--muted)/10 rounded-full transition-colors cursor-pointer"
                                        >
                                            Cancel
                                        </button>
                                        <button
                                            type="button"
                                            disabled={!draftComment.trim()}
                                            onClick={handleSubmitComment}
                                            className="px-4 py-1.5 text-xs font-medium bg-(--accent) text-white hover:bg-(--accent-hover) rounded-full transition-colors disabled:opacity-50 cursor-pointer"
                                        >
                                            Reply
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            )}

            {/* Delete Post Confirmation Modal */}
            <Modal
                isOpen={showDeleteModal}
                onClose={() => setShowDeleteModal(false)}
                title="Delete Post"
                description="Are you sure you want to delete this post? This action cannot be undone."
                onConfirm={handleDeleteConfirm}
                confirmText="Delete"
                cancelText="Cancel"
                isLoading={isDeleting}
                loadingText="Deleting..."
            />

            {/* Delete Comment Confirmation Modal */}
            <Modal
                isOpen={showDeleteCommentModal}
                onClose={() => {
                    setShowDeleteCommentModal(false);
                    setCommentToDelete(null);
                }}
                title="Delete Comment"
                description="Are you sure you want to delete this comment? This action cannot be undone."
                onConfirm={handleDeleteCommentConfirm}
                confirmText="Delete"
                cancelText="Cancel"
                isLoading={isDeletingComment}
                loadingText="Deleting..."
            />
        </div>
    );
}
