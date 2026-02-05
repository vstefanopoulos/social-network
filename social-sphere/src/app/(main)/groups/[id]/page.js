import { getGroup } from "@/actions/groups/get-group";
import { GroupHeader } from "@/components/groups/GroupHeader";
import GroupPageContent from "@/components/groups/GroupPageContent";
import { notFound, redirect } from "next/navigation";
import { getGroupPosts } from "@/actions/groups/get-group-posts";
import { getMostPopular } from "@/actions/groups/get-most-popular";
import GroupPostCard from "@/components/groups/GroupPostCard";
import Container from "@/components/layout/Container";
import { Lock } from "lucide-react";

export default async function GroupPage({ params }) {
  const { id } = await params;
  const result = await getGroup(id);

  let notfound
  if (!result.success) {
    notfound = result.status === 400;
  }

  const group = result?.data;

  let posts = null;

  if (group?.is_owner || group?.is_member) {
    const response = await getGroupPosts({ groupId: id, limit: 10 });
    posts = response.data;
  } else {
    const response = await getMostPopular(id);
    posts = response.data;
  }

  const isVisitor = !group?.is_owner && !group?.is_member;

  return (
    <div className="min-h-screen">
      {notfound ? (
        <div className="flex flex-col items-center justify-center py-50 animate-fade-in">
          <h3 className="text-lg font-semibold text-foreground mb-2">
            Page not found
          </h3>
          <p className="text-(--muted) text-center max-w-md px-4">
            Group not found or does not exist.
          </p>
        </div>
      ) : (
        <div>
          <GroupHeader group={group} />
          {isVisitor ? (
            <Container className="pt-6 pb-12">
              <GroupPostCard post={posts} onDelete={null} allowed={false} />
              <div className="flex flex-col items-center justify-center py-20 animate-fade-in">
                <div className="w-16 h-16 rounded-full bg-(--muted)/10 flex items-center justify-center mb-4">
                  <Lock className="w-8 h-8 text-(--muted)" />
                </div>
                <h3 className="text-lg font-semibold text-foreground mb-2">
                  This group is private
                </h3>
                <p className="text-(--muted) text-center max-w-md px-4">
                  Join @{group?.group_title} to see their posts and event updates.
                </p>
              </div>
            </Container>
          ) : (
            <GroupPageContent group={group} firstPosts={posts} />
          )}
        </div>
      )}

    </div>
  );
}