import { getProfileInfo } from "@/actions/profile/get-profile-info";
import ProfileContent from "@/components/profile/ProfileContent";
import { getUserPosts } from "@/actions/posts/get-user-posts";

export async function generateMetadata({ params }) {
    const { id } = await params;
    const result = await getProfileInfo(id);

    if (!result.success || !result.data) {
        return { title: "Profile" };
    }

    return {
        title: `${result.data.username}'s Profile`,
        description: `View ${result.data.first_name} ${result.data.last_name}'s profile`,
    };
}

export default async function ProfilePage({ params }) {
    const { id } = await params;
    const profileResult = await getProfileInfo(id);
    const postsResult = await getUserPosts({ creatorId: id, limit: 10, offset: 0 });

    return <ProfileContent
        result={profileResult}
        posts={postsResult.success ? postsResult.data : []}
    />;
}