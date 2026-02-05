import { getPost } from "@/actions/posts/get-post";
import SinglePostCard from "@/components/ui/SinglePostCard";
import { Lock } from "lucide-react";

export default async function PostPage({ params }) {
    const { id } = await params;
    const result = await getPost(id);

    const notAllowed = result.status === 403;
    const notFound = result.status === 400;

    return (
        <>
            {notAllowed ? (
                 <div className="flex flex-col items-center justify-center py-50 animate-fade-in">
                        <div className="w-16 h-16 rounded-full bg-(--muted)/10 flex items-center justify-center mb-4">
                            <Lock className="w-8 h-8 text-(--muted)" />
                        </div>
                        <h3 className="text-lg font-semibold text-foreground mb-2">
                            This post is private
                        </h3>
                        <p className="text-(--muted) text-center max-w-md px-4">
                            You do not have permission to view this post.
                        </p>
                    </div>
            ) : notFound ? (
                <div className="flex flex-col items-center justify-center py-50 animate-fade-in">
                        <h3 className="text-lg font-semibold text-foreground mb-2">
                            Page not found
                        </h3>
                        <p className="text-(--muted) text-center max-w-md px-4">
                            Post not found or does not exist.
                        </p>
                    </div>
            ) : (
                <div className="min-h-screen">
                    <div className="max-w-full mx-auto px-60 py-8">
                        <SinglePostCard post={result.data} />
                    </div>
                </div>
            )}
        </>
    );
}
