"use client";

import Link from "next/link";
import { Users } from "lucide-react";

export default function GroupCard({ group }) {
    return (
        <Link href={`/groups/${group.group_id}`} className="block group">
            <div className="bg-background border border-(--border) rounded-xl overflow-hidden hover:shadow-lg transition-all duration-300 hover:border-(--accent)/20 h-full flex flex-col">
                <div className="relative h-32 bg-(--muted)/10">
                    {group?.group_image_url ? (
                        <img
                            src={group.group_image_url}
                            alt={group.group_title || "Group image"}
                            className="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105"
                        />
                    ) : (
                        <div className="w-full h-full flex items-center justify-center text-(--muted)">
                            <Users className="w-8 h-8 opacity-20" />
                        </div>
                    )}
                    <div className="absolute inset-0 bg-linear-to-t from-black/60 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300" />

                    {/* Owner/Member Badge */}
                    {group.is_owner && (
                        <div className="absolute top-2 right-2 bg-(--accent) text-white px-2 py-1 rounded-md text-xs font-semibold shadow-lg">
                            Owner
                        </div>
                    )}
                    {!group.is_owner && group.is_member && (
                        <div className="absolute top-2 right-2 bg-green-500 text-white px-2 py-1 rounded-md text-xs font-semibold shadow-lg">
                            Member
                        </div>
                    )}
                </div>
                <div className="p-4 flex-1 flex flex-col">
                    <h3 className="font-semibold text-lg text-foreground mb-1 line-clamp-1 group-hover:text-(--accent) transition-colors">
                        {group.group_title}
                    </h3>
                    <p className="text-sm text-(--muted) line-clamp-2 mb-4 flex-1">
                        {group.group_description}
                    </p>
                    <div className="flex items-center justify-between mt-auto">
                        <span className="text-xs font-medium bg-(--muted)/10 text-(--muted) px-2 py-1 rounded-md">
                            {group.members_count || 0} Members
                        </span>
                    </div>
                </div>
            </div>
        </Link>
    );
}
