import { getNotifs } from "@/actions/notifs/get-user-notifs";
import NotificationsContent from "@/components/notifications/NotificationsContent";

export const metadata = {
    title: "Notifications",
};

export default async function NotificationsPage() {
    const result = await getNotifs({ limit: 20, offset: 0 });
    console.log("NOTIFICATIONS: ", result.data)

    return <NotificationsContent initialNotifications={result.success ? result.data : []} />;
}
