"use client";

import { useEffect, useState, useRef } from "react";
import { getAllGroups } from "@/actions/groups/get-all-groups";
import { getUserGroups } from "@/actions/groups/get-user-groups";
import { searchGroups } from "@/actions/groups/search-groups";
import { Users, Plus, Search, Loader2, Globe } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import CreateGroup from "@/components/groups/CreateGroup";
import GroupCard from "@/components/groups/GroupCard";
import GroupsPagination from "@/components/groups/GroupsPagination";
import { motion } from "motion/react";

export default function GroupsContent() {
    const router = useRouter();
    const [userGroups, setUserGroups] = useState([]);
    const [allGroups, setAllGroups] = useState([]);
    const [isCreateGroupOpen, setIsCreateGroupOpen] = useState(false);

    // Separate loading states for each section
    const [isLoadingUserGroups, setIsLoadingUserGroups] = useState(true);
    const [isLoadingAllGroups, setIsLoadingAllGroups] = useState(true);
    const [initialLoadDone, setInitialLoadDone] = useState(false);

    // Pagination state
    const [userGroupsPage, setUserGroupsPage] = useState(1);
    const [allGroupsPage, setAllGroupsPage] = useState(1);
    const [userGroupsTotal, setUserGroupsTotal] = useState(0);
    const [allGroupsTotal, setAllGroupsTotal] = useState(0);
    const groupsPerPage = 8;

    // Search state
    const [searchQuery, setSearchQuery] = useState("");
    const [searchResults, setSearchResults] = useState([]);
    const [isSearching, setIsSearching] = useState(false);
    const [showSearchResults, setShowSearchResults] = useState(false);
    const searchRef = useRef(null);

    useEffect(() => {
        fetchUserGroups();
    }, [userGroupsPage]);

    useEffect(() => {
        fetchAllGroups();
    }, [allGroupsPage]);

    const fetchUserGroups = async () => {
        setIsLoadingUserGroups(true);
        try {
            const offset = (userGroupsPage - 1) * groupsPerPage;
            // Fetch one extra to check if there are more
            const userGroupsRes = await getUserGroups({ limit: groupsPerPage + 1, offset });

            if (userGroupsRes.success && userGroupsRes.data) {
                const hasMore = userGroupsRes.data.length > groupsPerPage;
                // Only show the first 8 groups
                const displayGroups = hasMore ? userGroupsRes.data.slice(0, groupsPerPage) : userGroupsRes.data;
                setUserGroups(displayGroups);

                // Calculate total: if we have more, add one more page worth, otherwise exact count
                if (hasMore) {
                    setUserGroupsTotal(offset + groupsPerPage + 1); // At least one more item exists
                } else {
                    setUserGroupsTotal(offset + userGroupsRes.data.length);
                }
            }
        } catch (error) {
            console.error("Error fetching user groups:", error);
        } finally {
            setIsLoadingUserGroups(false);
        }
    };

    const fetchAllGroups = async () => {
        setIsLoadingAllGroups(true);
        try {
            const offset = (allGroupsPage - 1) * groupsPerPage;
            // Fetch one extra to check if there are more
            const allGroupsRes = await getAllGroups({ limit: groupsPerPage + 1, offset });

            if (allGroupsRes.success && allGroupsRes.data) {
                const hasMore = allGroupsRes.data.length > groupsPerPage;
                // Only show the first 8 groups
                const displayGroups = hasMore ? allGroupsRes.data.slice(0, groupsPerPage) : allGroupsRes.data;
                setAllGroups(displayGroups);

                // Calculate total: if we have more, add one more page worth, otherwise exact count
                if (hasMore) {
                    setAllGroupsTotal(offset + groupsPerPage + 1); // At least one more item exists
                } else {
                    setAllGroupsTotal(offset + allGroupsRes.data.length);
                }
            }
        } catch (error) {
            console.error("Error fetching all groups:", error);
        } finally {
            setIsLoadingAllGroups(false);
            setInitialLoadDone(true);
        }
    };

    // Close search results when clicking outside
    useEffect(() => {
        function handleClickOutside(event) {
            if (searchRef.current && !searchRef.current.contains(event.target)) {
                setShowSearchResults(false);
            }
        }

        document.addEventListener("mousedown", handleClickOutside);
        return () => {
            document.removeEventListener("mousedown", handleClickOutside);
        };
    }, []);

    // Debounced search
    useEffect(() => {
        const timer = setTimeout(async () => {
            if (searchQuery.trim().length >= 2) {
                setIsSearching(true);
                try {
                    const response = await searchGroups({ query: searchQuery, limit: 10 });
                    if (response.success && response.data?.groups) {
                        setSearchResults(response.data.groups);
                        setShowSearchResults(true);
                    } else {
                        setSearchResults([]);
                    }
                } catch (error) {
                    console.error("Search error:", error);
                    setSearchResults([]);
                } finally {
                    setIsSearching(false);
                }
            } else {
                setSearchResults([]);
                setShowSearchResults(false);
            }
        }, 300); // 300ms debounce

        return () => clearTimeout(timer);
    }, [searchQuery]);

    // const handleCreateGroupSuccess = (groupId) => {
    //     // Redirect to the newly created group page
    //     router.push(`/groups/${groupId}`);
    // };

    return (
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-12">
            {/* Header */}
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                <div>
                    <h1 className="text-3xl font-bold text-foreground tracking-tight">Communities</h1>
                    <p className="text-(--muted) mt-1">Discover specific interest groups and connect with like-minded people.</p>
                </div>
                <button
                    onClick={() => setIsCreateGroupOpen(true)}
                    className="flex items-center gap-2 bg-(--accent) text-white px-5 py-2.5 rounded-full font-medium text-sm hover:bg-(--accent-hover) transition-all shadow-lg shadow-black/5 cursor-pointer"
                >
                    <Plus className="w-4 h-4" />
                    Create Group
                </button>
            </div>

            {/* Search Bar */}
            <div className="max-w-lg" ref={searchRef}>
                <div className="relative w-full group">
                    <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                        {isSearching ? (
                            <Loader2 className="h-5 w-5 text-(--muted) animate-spin" />
                        ) : (
                            <Search className="h-5 w-5 text-(--muted) group-focus-within:text-(--accent) transition-colors" />
                        )}
                    </div>
                    <input
                        type="text"
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        onFocus={() => {
                            if (searchResults.length > 0) setShowSearchResults(true);
                        }}
                        className="block w-full pl-12 pr-4 py-3 border border-(--border) rounded-full text-sm bg-(--muted)/5 text-foreground placeholder-(--muted) hover:border-foreground focus:outline-none focus:border-(--accent) focus:ring-2 focus:ring-(--accent)/10 transition-all"
                        placeholder="Search groups..."
                    />

                    {/* Search Results Dropdown */}
                    {showSearchResults && (
                        <div className="absolute top-full left-0 right-0 mt-2 bg-background border border-(--border) rounded-2xl shadow-xl overflow-hidden animate-in fade-in zoom-in-95 duration-200 max-h-96 overflow-y-auto z-50">
                            {searchResults.length > 0 ? (
                                <div className="py-2">
                                    {searchResults.map((group) => (
                                        <Link
                                            key={group.group_id}
                                            href={`/groups/${group.group_id}`}
                                            onClick={() => {
                                                setShowSearchResults(false);
                                                setSearchQuery("");
                                            }}
                                            className="flex items-center gap-3 px-4 py-3 hover:bg-(--muted)/5 transition-colors"
                                        >
                                            <div className="w-12 h-12 rounded-lg bg-(--muted)/10 flex items-center justify-center overflow-hidden shrink-0">
                                                {group.group_image_url ? (
                                                    <img src={group.group_image_url} alt={group.group_title || "Group"} className="w-full h-full object-cover" />
                                                ) : (
                                                    <Users className="w-6 h-6 text-(--muted)" />
                                                )}
                                            </div>
                                            <div className="flex-1 min-w-0">
                                                <p className="text-sm font-semibold text-foreground truncate">
                                                    {group.group_title}
                                                </p>
                                                <p className="text-xs text-(--muted) truncate">
                                                    {group.members_count || 0} members
                                                </p>
                                            </div>
                                            {group.is_owner && (
                                                <span className="text-xs font-medium bg-(--accent)/10 text-(--accent) px-2 py-1 rounded-md">
                                                    Owner
                                                </span>
                                            )}
                                            {!group.is_owner && group.is_member && (
                                                <span className="text-xs font-medium bg-green-500/10 text-green-600 dark:text-green-400 px-2 py-1 rounded-md">
                                                    Member
                                                </span>
                                            )}
                                        </Link>
                                    ))}
                                </div>
                            ) : (
                                <div className="p-4 text-center text-sm text-(--muted)">
                                    No groups found
                                </div>
                            )}
                        </div>
                    )}
                </div>
            </div>

            {/* Initial full page loading */}
            {!initialLoadDone ? (
                <div className="flex items-center justify-center py-20">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-(--accent)"></div>
                </div>
            ) : (
                <>
                    {/* My Groups Section */}
                    {(userGroups.length > 0 || isLoadingUserGroups) && (
                        <div className="space-y-6">
                            <div className="flex items-center gap-2">
                                <Users className="w-5 h-5 text-(--accent)" />
                                <h2 className="text-xl font-bold text-foreground">My Groups</h2>
                            </div>
                            {isLoadingUserGroups ? (
                                <div className="flex items-center justify-center py-12">
                                    <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-(--accent)"></div>
                                </div>
                            ) : (
                                <>
                                    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                                        {userGroups.map((group, index) => (
                                            <motion.div
                                                key={group.group_id}
                                                initial={{ opacity: 0, scale: 0.9 }}
                                                animate={{ opacity: 1, scale: 1 }}
                                                transition={{
                                                    duration: 0.3,
                                                    delay: index * 0.05,
                                                    ease: "easeOut"
                                                }}
                                            >
                                                <GroupCard group={group} />
                                            </motion.div>
                                        ))}
                                    </div>
                                    <GroupsPagination
                                        currentPage={userGroupsPage}
                                        totalItems={userGroupsTotal}
                                        itemsPerPage={groupsPerPage}
                                        onPageChange={setUserGroupsPage}
                                    />
                                </>
                            )}
                        </div>
                    )}

                    {/* All Groups Section */}
                    <div className="space-y-6">
                        <div className="flex items-center gap-2">
                            <Globe className="w-5 h-5 text-(--accent)" />
                            <h2 className="text-xl font-bold text-foreground">Discover</h2>
                        </div>
                        {isLoadingAllGroups ? (
                            <div className="flex items-center justify-center py-12">
                                <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-(--accent)"></div>
                            </div>
                        ) : allGroups.length > 0 ? (
                            <>
                                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                                    {allGroups.map((group, index) => (
                                        <motion.div
                                            key={group.group_id}
                                            initial={{ opacity: 0, scale: 0.9 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            transition={{
                                                duration: 0.3,
                                                delay: index * 0.05,
                                                ease: "easeOut"
                                            }}
                                        >
                                            <GroupCard group={group} />
                                        </motion.div>
                                    ))}
                                </div>
                                <GroupsPagination
                                    currentPage={allGroupsPage}
                                    totalItems={allGroupsTotal}
                                    itemsPerPage={groupsPerPage}
                                    onPageChange={setAllGroupsPage}
                                />
                            </>
                        ) : (
                            <div className="text-center py-12 bg-(--muted)/5 rounded-2xl border border-dashed border-(--border)">
                                <Users className="w-12 h-12 text-(--muted) mx-auto mb-3 opacity-20" />
                                <p className="text-(--muted)">No groups found. Be the first to create one!</p>
                            </div>
                        )}
                    </div>
                </>
            )}

            {/* Create Group Modal */}
            <CreateGroup
                isOpen={isCreateGroupOpen}
                onClose={() => setIsCreateGroupOpen(false)}
            />
        </div>
    );
}
