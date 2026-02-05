import { getFriendsPosts } from "@/actions/posts/get-friends-posts";
import FriendsFeedContent from "@/components/feed/FriendsFeedContent";

export const metadata = {
    title: "Friends Feed",
}

export default async function FriendsFeedPage() {
    // Fetch initial 10 posts
    const result = await getFriendsPosts({ limit: 10, offset: 0 });

    return <FriendsFeedContent initialPosts={result.success ? result.data : []} />;
}