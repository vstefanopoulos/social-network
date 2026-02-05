import { getProfileInfo } from "@/actions/profile/get-profile-info";
import { redirect } from "next/navigation";
import SettingsClient from "./SettingsClient";

export default async function SettingsPage({ params }) {
    const { id } = await params;

    // get user's info
    const result = await getProfileInfo(id);

    if (!result.success || result.data?.user_id !== id) { redirect(`/profile/${id}`); }

    return <SettingsClient user={result.data} />;
}
