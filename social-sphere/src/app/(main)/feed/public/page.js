import { getPublicPosts } from "@/actions/posts/get-public-posts";
import PublicFeedContent from "@/components/feed/PublicFeedContent";

export const metadata = {
    title: "Public Feed",
}

export default async function PublicFeedPage() {
    // call backend for public posts - initial 10 posts
    const limit = 10;
    const offset = 0;
    const result = await getPublicPosts({ limit, offset });

    return <PublicFeedContent initialPosts={result.success ? result.data : []} />;
}