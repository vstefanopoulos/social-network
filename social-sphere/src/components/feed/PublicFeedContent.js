"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { motion, AnimatePresence } from "motion/react";
import PostCard from "@/components/ui/PostCard";
import CreatePost from "@/components/ui/CreatePost";
import Container from "@/components/layout/Container";
import { getPublicPosts } from "@/actions/posts/get-public-posts";

export default function PublicFeedContent({ initialPosts }) {
    const [posts, setPosts] = useState(initialPosts || []);
    const [offset, setOffset] = useState(10); // Start after the initial 10 posts
    // Only hasMore if we got a full batch of 10 posts
    const [hasMore, setHasMore] = useState((initialPosts || []).length >= 10);
    const [loading, setLoading] = useState(false);
    const observerTarget = useRef(null);

    const handleNewPost = (newPost) => {
        if (newPost.audience !== "everyone") {
            return;
        }
        setPosts(prev => [newPost, ...prev]);
    }

    const loadMorePosts = useCallback(async () => {
        if (loading || !hasMore) return;

        setLoading(true);
        try {
            const result = await getPublicPosts({ limit: 5, offset });
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
    }, [offset, loading, hasMore]);

    useEffect(() => {
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
    }, [loadMorePosts, hasMore, loading]);

    return (
        <div className="w-full">
            {/* Create Post Section */}
            <Container className="pt-6 md:pt-10">
                <CreatePost onPostCreated={handleNewPost} />
            </Container>

            {/* Feed Header */}
            <div className="mt-8 mb-6">
                <h1 className="text-center feed-title px-4">Public Feed</h1>
                <p className="text-center feed-subtitle px-4">What's happening in global sphere?</p>
            </div>

            <div className="section-divider mb-6" />

            {/* Posts Feed */}
            <Container className="pt-6 pb-12">
                {posts?.length > 0 ? (
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
                            Be the first ever to share something on the public sphere!
                        </p>
                    </div>
                )}
            </Container>
        </div>
    );
}
