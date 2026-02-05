"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { motion, AnimatePresence } from "motion/react";
import { ProfileHeader } from "@/components/profile/ProfileHeader";
import CreatePost from "@/components/ui/CreatePost";
import PostCard from "@/components/ui/PostCard";
import Container from "@/components/layout/Container";
import { Lock } from "lucide-react";
import { getUserPosts } from "@/actions/posts/get-user-posts";

export default function ProfileContent({ result, posts: initialPosts }) {
    const [posts, setPosts] = useState(initialPosts || []);
    const [offset, setOffset] = useState(10); // Start after the initial 10 posts
    // Only hasMore if we got a full batch of 10 posts
    const [hasMore, setHasMore] = useState((initialPosts || []).length >= 10);
    const [loading, setLoading] = useState(false);
    const observerTarget = useRef(null);
    const [isPublic, setIsPublic] = useState(result.data?.public);
    const [isFollowing, setIsFollowing] = useState(result.data?.viewer_is_following);

    const handleNewPost = (newPost) => {
        setPosts(prev => [newPost, ...prev]);
    }

    const handleUnfollow = ({isPublic, isFollowing}) => {
        setIsPublic(isPublic);
        setIsFollowing(isFollowing);
    }

    const user = result.data;

    // Handle error state
    if (!result.success) {
        return (
            <div className="flex flex-col items-center justify-center py-50 animate-fade-in">
          <h3 className="text-lg font-semibold text-foreground mb-2">
            User not found
          </h3>
          <p className="text-(--muted) text-center max-w-md px-4">
            User not found or does not exist.
          </p>
        </div>
        );
    }

    // Handle no user
    if (!user) {
        return (
            <div className="flex items-center justify-center min-h-screen px-4">
                <div className="text-(--muted) text-lg">User not found</div>
            </div>
        );
    }

    // Check if viewer can see the profile content
    const canViewProfile = user.own_profile || isPublic || isFollowing;

    const loadMorePosts = useCallback(async () => {
        if (loading || !hasMore || !canViewProfile) return;

        setLoading(true);
        try {
            const result = await getUserPosts({ creatorId: user.user_id, limit: 5, offset });
            const newPosts = result.success ? result.data : [];

            if (newPosts && newPosts.length > 0) {
                setPosts((prevPosts) => [...prevPosts, ...newPosts]);
                setOffset((prevOffset) => prevOffset + 5);

                // If we got fewer than 5 posts, we've reached the end
                if (newPosts.length < 5) {
                    setHasMore(false);
                }
            } else {
                setHasMore(false);
            }
        } catch (error) {
            console.error("Failed to load more posts:", error);
        } finally {
            setLoading(false);
        }
    }, [offset, loading, hasMore, canViewProfile, user.user_id]);

    useEffect(() => {
        if (!canViewProfile) return;

        const observer = new IntersectionObserver(
            (entries) => {
                if (entries[0].isIntersecting && hasMore && !loading) {
                    loadMorePosts();
                }
            },
            { threshold: 0.1 }
        );

        if (observerTarget.current) {
            observer.observe(observerTarget.current);
        }

        return () => {
            if (observerTarget.current) {
                observer.unobserve(observerTarget.current);
            }
        };
    }, [loadMorePosts, hasMore, loading, canViewProfile]);

    // Render profile
    return (
        <div className="w-full">
            <ProfileHeader user={user} onUnfollow={handleUnfollow}/>

            {user.own_profile ? (
                <div>
                    <Container className="pt-6 md:pt-10">
                        <CreatePost onPostCreated={handleNewPost} />
                    </Container>
                    <div className="mt-8 mb-6">
                        <h1 className="text-center feed-title px-4">My Feed</h1>
                        <p className="text-center feed-subtitle px-4">What's happening in my sphere?</p>
                    </div>
                    <div className="section-divider mb-6" />
                </div>
            ) : canViewProfile ? (
                <div>
                    <div className="mt-8 mb-6">
                        <h1 className="text-center feed-title px-4">{user.username}'s Feed</h1>
                        <p className="text-center feed-subtitle px-4">What's happening in {user.username}'s sphere?</p>
                    </div>
                    <div className="section-divider mb-6" />
                </div>
            ) : null}

            <Container className="pt-6 pb-12">
                {canViewProfile ? (
                    posts?.length > 0 ? (
                        <div className="flex flex-col">
                            <AnimatePresence mode="popLayout">
                                {posts.map((post, index) => (
                                    <motion.div
                                        key={post.post_id + index}
                                        initial={{ opacity: 0, scale: 0.8 }}
                                        animate={{ opacity: 1, scale: 1 }}
                                        exit={{ opacity: 0, scale: 0.95 }}
                                        transition={{
                                            duration: 0.3,
                                            delay: index * 0.1,
                                            ease: "easeOut"
                                        }}
                                        layout
                                    >
                                        <PostCard
                                            post={post}
                                            onDelete={(postId) => setPosts(prev => prev.filter(p => p.post_id !== postId))}
                                        />
                                    </motion.div>
                                ))}
                            </AnimatePresence>

                            {/* Loading indicator */}
                            {hasMore && (
                                <div ref={observerTarget} className="flex justify-center py-8">
                                    {loading && (
                                        <div className="text-sm text-(--muted)">Loading more posts...</div>
                                    )}
                                </div>
                            )}

                            {/* End of feed message */}
                            {!hasMore && posts.length > 0 && (
                                <div className="text-center py-8 text-xl font-bold text-(--muted)">
                                    .
                                </div>
                            )}
                        </div>
                    ) : (
                        <div className="flex flex-col items-center justify-center py-20 animate-fade-in">
                            <p className="text-muted text-center max-w-md px-4">
                                Nothing.
                            </p>
                        </div>
                    )
                ) : (
                    <div className="flex flex-col items-center justify-center py-20 animate-fade-in">
                        <div className="w-16 h-16 rounded-full bg-(--muted)/10 flex items-center justify-center mb-4">
                            <Lock className="w-8 h-8 text-(--muted)" />
                        </div>
                        <h3 className="text-lg font-semibold text-foreground mb-2">
                            This profile is private
                        </h3>
                        <p className="text-(--muted) text-center max-w-md px-4">
                            Follow @{user.username} to see their posts and profile details.
                        </p>
                    </div>
                )}
            </Container>
        </div>
    );
}