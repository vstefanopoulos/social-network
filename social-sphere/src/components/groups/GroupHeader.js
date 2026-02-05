"use client";

import { useState, useRef, useCallback, useEffect } from "react";
import { Users, UserPlus, Settings, LogOut, Clock, UserRoundPlus, User, Check, Loader2, X, UserCheck, Trash2 } from "lucide-react";
import { motion, AnimatePresence } from "motion/react";
import Modal from "@/components/ui/Modal";
import Container from "@/components/layout/Container";
import { requestJoinGroup } from "@/actions/groups/request-join-group";
import { cancelJoinRequest } from "@/actions/groups/cancel-join-request";
import { leaveGroup } from "@/actions/groups/leave-group";
import { inviteToGroup } from "@/actions/groups/invite-to-group";
import { respondToGroupInvite } from "@/actions/groups/respond-to-invite";
import { getPendingRequests } from "@/actions/groups/get-pending-requests";
import { getPendingRequestsCount } from "@/actions/groups/get-pening-count";
import { handleJoinRequest as handleJoinRequestAction } from "@/actions/groups/handle-join-request";
import { getNotInvited } from "@/actions/users/get-not-invited";
import { getGroupMembers } from "@/actions/groups/group-members";
import { removeFromGroup } from "@/actions/groups/remove-from-group";
import Tooltip from "../ui/Tooltip";
import UpdateGroupModal from "./UpdateGroupModal";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useStore } from "@/store/store";

export function GroupHeader({ group }) {
    const router = useRouter();
    const user = useStore((state) => state.user);
    const [isMember, setIsMember] = useState(group?.is_member);
    const [isOwner] = useState(group?.is_owner);
    const [pendingRequest, setPendingRequest] = useState(group?.pending_request);
    const [pendingInvite, setPendingInvite] = useState(group?.pending_invite);
    const [isLoading, setIsLoading] = useState(false);
    const [showLeaveModal, setShowLeaveModal] = useState(false);
    const [showInviteModal, setShowInviteModal] = useState(false);
    const [showUpdateModal, setShowUpdateModal] = useState(false);
    const [membersCount, setMembersCount] = useState(group?.members_count);

    // Invite modal state
    const [followers, setFollowers] = useState([]);
    const [selectedUsers, setSelectedUsers] = useState([]);
    const [isLoadingFollowers, setIsLoadingFollowers] = useState(false);
    const [isLoadingMore, setIsLoadingMore] = useState(false);
    const [isInviting, setIsInviting] = useState(false);
    const [inviteSuccess, setInviteSuccess] = useState(false);
    const [hasMore, setHasMore] = useState(true);
    const [offset, setOffset] = useState(0);
    const FOLLOWERS_LIMIT = 10;
    const scrollContainerRef = useRef(null);

    // Pending requests modal state
    const [showPendingModal, setShowPendingModal] = useState(false);
    const [pendingUsers, setPendingUsers] = useState([]);
    const [isLoadingPending, setIsLoadingPending] = useState(false);
    const [processingUser, setProcessingUser] = useState(null);
    const [isLoadingMorePending, setIsLoadingMorePending] = useState(false);
    const [hasMorePending, setHasMorePending] = useState(true);
    const [pendingOffset, setPendingOffset] = useState(0);
    const PENDING_LIMIT = 10;
    const pendingScrollRef = useRef(null);
    const [pendingCount, setPendingCount] = useState(null);

    // Members modal state
    const [showMembersModal, setShowMembersModal] = useState(false);
    const [members, setMembers] = useState([]);
    const [isLoadingMembers, setIsLoadingMembers] = useState(false);
    const [isLoadingMoreMembers, setIsLoadingMoreMembers] = useState(false);
    const [hasMoreMembers, setHasMoreMembers] = useState(true);
    const [membersOffset, setMembersOffset] = useState(0);
    const [removingMember, setRemovingMember] = useState(null);
    const MEMBERS_LIMIT = 10;
    const membersScrollRef = useRef(null);

    // Scroll-aware mini header
    const [showMiniHeader, setShowMiniHeader] = useState(false);
    const headerRef = useRef(null);

    useEffect(() => {
        const handleScroll = () => {
            if (headerRef.current) {
                const headerBottom = headerRef.current.getBoundingClientRect().bottom;
                setShowMiniHeader(headerBottom < 0);
            }
        };

        window.addEventListener("scroll", handleScroll, { passive: true });
        return () => window.removeEventListener("scroll", handleScroll);
    }, []);

    // Fetch count on mount (for owners only)
    useEffect(() => {
        if (!isOwner) return;

        const fetchPendingCount = async () => {
            try {
                const result = await getPendingRequestsCount({
                    groupId: group.group_id
                });
                if (result.success && result.data) {
                    setPendingCount(result.data);
                } else {
                    setPendingCount(null);
                }
            } catch (error) {
                console.error("Error fetching pending count:", error);
                setPendingCount(null);
            }
        };

        fetchPendingCount();
    }, [group.group_id, isOwner]);


    const handleJoinRequest = async () => {
        if (isLoading) return;
        setIsLoading(true);

        try {
            const response = await requestJoinGroup({ groupId: group.group_id });
            if (response.success) {
                setPendingRequest(true);
            } else {
                console.error("Error requesting to join:", response.error);
            }
        } catch (error) {
            console.error("Error requesting to join:", error);
        } finally {
            setIsLoading(false);
        }
    };

    const handleCancelRequest = async () => {
        if (isLoading) return;
        setIsLoading(true);

        try {
            const response = await cancelJoinRequest({ groupId: group.group_id });
            if (response.success) {
                setPendingRequest(false);
            } else {
                console.error("Error canceling request:", response.error);
            }
        } catch (error) {
            console.error("Error canceling request:", error);
        } finally {
            setIsLoading(false);
        }
    };

    const handleInviteResponse = async (accept) => {
        if (isLoading) return;
        setIsLoading(true);

        try {
            const response = await respondToGroupInvite({
                groupId: group.group_id,
                accept
            });
            if (response.success) {
                setPendingInvite(false);
                if (accept) {
                    setMembersCount(prev => prev + 1);
                    setIsMember(true);
                }
            } else {
                console.error("Error responding to invite:", response.error);
            }
        } catch (error) {
            console.error("Error responding to invite:", error);
        } finally {
            setIsLoading(false);
        }
    };

    const handleLeaveGroup = async () => {
        if (isLoading) return;
        setIsLoading(true);

        try {
            const response = await leaveGroup({ groupId: group.group_id });
            if (response.success) {
                setIsMember(false);
                setMembersCount(prev => prev - 1);
                setShowLeaveModal(false);
            } else {
                console.error("Error leaving group:", response.error);
            }
        } catch (error) {
            console.error("Error leaving group:", error);
        } finally {
            setIsLoading(false);
        }
    };

    const handleInviteMembers = async () => {
        setShowInviteModal(true);
        setInviteSuccess(false);
        setSelectedUsers([]);
        setFollowers([]);
        setOffset(0);
        setHasMore(true);

        if (user) {
            setIsLoadingFollowers(true);
            try {
                const result = await getNotInvited({
                    groupId: group.group_id,
                    limit: FOLLOWERS_LIMIT,
                    offset: 0
                });
                if (Array.isArray(result.data?.users)) {
                    setFollowers(result.data.users);
                    setHasMore(result.data.users.length === FOLLOWERS_LIMIT);
                    setOffset(FOLLOWERS_LIMIT);
                } else {
                    setFollowers([]);
                    setHasMore(false);
                }
            } catch (error) {
                console.error("Error fetching followers:", error);
                setFollowers([]);
                setHasMore(false);
            } finally {
                setIsLoadingFollowers(false);
            }
        }
    };

    const handleCloseInviteModal = () => {
        setShowInviteModal(false);
        setSelectedUsers([]);
        setFollowers([]);
        setInviteSuccess(false);
        setOffset(0);
        setHasMore(true);
    };

    const loadMoreFollowers = useCallback(async () => {
        if (isLoadingMore || !hasMore) return;

        setIsLoadingMore(true);
        try {
            const result = await getNotInvited({
                groupId: group.group_id,
                limit: FOLLOWERS_LIMIT,
                offset: offset
            });
            if (Array.isArray(result.data?.users) && result.data.users.length > 0) {
                setFollowers((prev) => [...prev, ...result.data.users]);
                setHasMore(result.data.users.length === FOLLOWERS_LIMIT);
                setOffset((prev) => prev + FOLLOWERS_LIMIT);
            } else {
                setHasMore(false);
            }
        } catch (error) {
            console.error("Error loading more followers:", error);
        } finally {
            setIsLoadingMore(false);
        }
    }, [isLoadingMore, hasMore, offset, group.group_id]);

    const handleScroll = useCallback((e) => {
        const { scrollTop, scrollHeight, clientHeight } = e.target;
        if (scrollHeight - scrollTop - clientHeight < 50 && hasMore && !isLoadingMore) {
            loadMoreFollowers();
        }
    }, [hasMore, isLoadingMore, loadMoreFollowers]);

    const toggleUserSelection = (userId) => {
        setSelectedUsers((prev) =>
            prev.includes(userId)
                ? prev.filter((id) => id !== userId)
                : [...prev, userId]
        );
    };

    const handleSendInvites = async () => {
        if (selectedUsers.length === 0 || isInviting) return;

        setIsInviting(true);
        try {
            const response = await inviteToGroup({
                groupId: group.group_id,
                invitedIds: selectedUsers,
            });

            if (response.success) {
                setInviteSuccess(true);
                setTimeout(() => {
                    handleCloseInviteModal();
                }, 1500);
            } else {
                console.error("Error inviting users:", response.error);
            }
        } catch (error) {
            console.error("Error inviting users:", error);
        } finally {
            setIsInviting(false);
        }
    };

    const handleUpdateSuccess = () => {
        // Refresh the page to get updated group data
        router.refresh();
    };

    // Pending requests handlers
    const handleOpenPendingModal = async () => {
        setShowPendingModal(true);
        setIsLoadingPending(true);
        setPendingUsers([]);
        setPendingOffset(0);
        setHasMorePending(true);
        try {
            const result = await getPendingRequests({
                groupId: group.group_id,
                limit: PENDING_LIMIT,
                offset: 0
            });
            if (Array.isArray(result.data?.users)) {
                setPendingUsers(result.data.users);
                setHasMorePending(result.data.users.length === PENDING_LIMIT);
                setPendingOffset(PENDING_LIMIT);
            } else {
                setPendingUsers([]);
                setHasMorePending(false);
            }
        } catch (error) {
            console.error("Error fetching pending requests:", error);
            setPendingUsers([]);
            setHasMorePending(false);
        } finally {
            setIsLoadingPending(false);
        }
    };

    const handleClosePendingModal = () => {
        setShowPendingModal(false);
        setPendingUsers([]);
        setProcessingUser(null);
        setPendingOffset(0);
        setHasMorePending(true);
    };

    const loadMorePendingUsers = useCallback(async () => {
        if (isLoadingMorePending || !hasMorePending) return;

        setIsLoadingMorePending(true);
        try {
            const result = await getPendingRequests({
                groupId: group.group_id,
                limit: PENDING_LIMIT,
                offset: pendingOffset
            });
            if (Array.isArray(result.data?.users) && result.data.users.length > 0) {
                setPendingUsers((prev) => [...prev, ...result.data.users]);
                setHasMorePending(result.data.users.length === PENDING_LIMIT);
                setPendingOffset((prev) => prev + PENDING_LIMIT);
            } else {
                setHasMorePending(false);
            }
        } catch (error) {
            console.error("Error loading more pending users:", error);
        } finally {
            setIsLoadingMorePending(false);
        }
    }, [isLoadingMorePending, hasMorePending, pendingOffset, group.group_id]);

    const handlePendingScroll = useCallback((e) => {
        const { scrollTop, scrollHeight, clientHeight } = e.target;
        if (scrollHeight - scrollTop - clientHeight < 50 && hasMorePending && !isLoadingMorePending) {
            loadMorePendingUsers();
        }
    }, [hasMorePending, isLoadingMorePending, loadMorePendingUsers]);

    const handleRequest = async (requesterId, accepted) => {
        setProcessingUser(requesterId);
        try {
            const response = await handleJoinRequestAction({
                groupId: group.group_id,
                requesterId: requesterId,
                accepted: accepted
            });
            if (response.success) {
                // Remove user from pending list
                setPendingUsers((prev) => prev.filter((u) => u.id !== requesterId));

                setPendingCount((prev) => {
                    if (prev <= 1) return null;
                    return prev - 1;
                });

                if (accepted) {
                    setMembersCount((prev) => prev + 1);
                }

            } else {
                console.error("Error handling request:", response.error);
            }
        } catch (error) {
            console.error("Error handling request:", error);
        } finally {
            setProcessingUser(null);
        }
    };

    // Members modal handlers
    const handleOpenMembersModal = async () => {
        if (!isMember && !isOwner) return; // Non-members cannot open

        setShowMembersModal(true);
        setIsLoadingMembers(true);
        setMembers([]);
        setMembersOffset(0);
        setHasMoreMembers(true);
        try {
            const result = await getGroupMembers({
                group_id: group.group_id,
                limit: MEMBERS_LIMIT,
                offset: 0
            });
            if (result.success && result.data?.group_users) {
                setMembers(result.data.group_users);
                setHasMoreMembers(result.data.group_users.length === MEMBERS_LIMIT);
                setMembersOffset(MEMBERS_LIMIT);
            } else {
                setMembers([]);
                setHasMoreMembers(false);
            }
        } catch (error) {
            console.error("Error fetching members:", error);
            setMembers([]);
            setHasMoreMembers(false);
        } finally {
            setIsLoadingMembers(false);
        }
    };

    const handleCloseMembersModal = () => {
        setShowMembersModal(false);
        setMembers([]);
        setRemovingMember(null);
        setMembersOffset(0);
        setHasMoreMembers(true);
    };

    const loadMoreMembers = useCallback(async () => {
        if (isLoadingMoreMembers || !hasMoreMembers) return;

        setIsLoadingMoreMembers(true);
        try {
            const result = await getGroupMembers({
                group_id: group.group_id,
                limit: MEMBERS_LIMIT,
                offset: membersOffset
            });
            if (result.success && result.data?.group_users?.length > 0) {
                setMembers((prev) => [...prev, ...result.data.group_users]);
                setHasMoreMembers(result.data.group_users.length === MEMBERS_LIMIT);
                setMembersOffset((prev) => prev + MEMBERS_LIMIT);
            } else {
                setHasMoreMembers(false);
            }
        } catch (error) {
            console.error("Error loading more members:", error);
        } finally {
            setIsLoadingMoreMembers(false);
        }
    }, [isLoadingMoreMembers, hasMoreMembers, membersOffset, group.group_id]);

    const handleMembersScroll = useCallback((e) => {
        const { scrollTop, scrollHeight, clientHeight } = e.target;
        if (scrollHeight - scrollTop - clientHeight < 50 && hasMoreMembers && !isLoadingMoreMembers) {
            loadMoreMembers();
        }
    }, [hasMoreMembers, isLoadingMoreMembers, loadMoreMembers]);

    const handleRemoveMember = async (memberId) => {
        setRemovingMember(memberId);
        try {
            const response = await removeFromGroup({
                groupId: group.group_id,
                memberId: memberId
            });
            if (response.success) {
                setMembers((prev) => prev.filter((m) => m.user_id !== memberId));
                setMembersCount((prev) => prev - 1);
            } else {
                console.error("Error removing member:", response.error);
            }
        } catch (error) {
            console.error("Error removing member:", error);
        } finally {
            setRemovingMember(null);
        }
    };

    return (
        <>
            {/* Mini Header - appears on scroll */}
            <AnimatePresence>
                {showMiniHeader && (
                    <motion.div
                        initial={{ y: -100, opacity: 0 }}
                        animate={{ y: 0, opacity: 1 }}
                        exit={{ y: -100, opacity: 0 }}
                        transition={{ type: "spring", damping: 30, stiffness: 300 }}
                        className="fixed top-16 left-0 right-0 z-40 bg-background/80 backdrop-blur-xl border-b border-(--border)/50"
                    >
                        <Container>
                            <div className="flex items-center justify-between py-3">
                                <div className="flex items-center gap-3">
                                    {/* Small Group Image */}
                                    <div className="w-9 h-9 rounded-lg overflow-hidden bg-(--muted)/10 border border-(--border)">
                                        {group.group_image_url ? (
                                            <img
                                                src={group.group_image_url}
                                                alt={group.group_title}
                                                className="w-full h-full object-cover"
                                            />
                                        ) : (
                                            <div className="w-full h-full flex items-center justify-center">
                                                <Users className="w-4 h-4 text-(--muted)" />
                                            </div>
                                        )}
                                    </div>
                                    {/* Title & Member Count */}
                                    <div>
                                        <p className="font-semibold text-foreground text-sm">
                                            {group.group_title}
                                        </p>
                                        <p className="text-xs text-(--muted)">
                                            {membersCount} {membersCount === 1 ? "member" : "members"}
                                        </p>
                                    </div>
                                </div>

                                {/* Action Buttons */}
                                <div className="flex items-center gap-2">
                                    {isOwner ? (
                                        <>
                                            <button
                                                onClick={handleInviteMembers}
                                                className="flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium bg-(--accent) text-white cursor-pointer"
                                            >
                                                <UserPlus className="w-3.5 h-3.5" />
                                                Invite
                                            </button>
                                            <button
                                                onClick={() => setShowUpdateModal(true)}
                                                className="p-1.5 rounded-full border border-(--border) text-(--muted) hover:text-foreground transition-colors cursor-pointer"
                                            >
                                                <Settings className="w-3.5 h-3.5" />
                                            </button>
                                        </>
                                    ) : isMember ? (
                                        <>
                                            <button
                                                onClick={handleInviteMembers}
                                                className="flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium border border-(--accent) text-(--accent) cursor-pointer"
                                            >
                                                <UserPlus className="w-3.5 h-3.5" />
                                                Invite
                                            </button>
                                            <button
                                                onClick={() => setShowLeaveModal(true)}
                                                className="p-1.5 rounded-full border border-(--border) text-(--muted) hover:text-red-500 transition-colors cursor-pointer"
                                            >
                                                <LogOut className="w-3.5 h-3.5" />
                                            </button>
                                        </>
                                    ) : pendingInvite ? (
                                        <div className="flex items-center gap-2">
                                            <button
                                                onClick={() => handleInviteResponse(false)}
                                                disabled={isLoading}
                                                className="p-1.5 rounded-full border border-(--border) text-(--muted) hover:text-red-500 transition-colors cursor-pointer"
                                            >
                                                <X className="w-3.5 h-3.5" />
                                            </button>
                                            <button
                                                onClick={() => handleInviteResponse(true)}
                                                disabled={isLoading}
                                                className="flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium bg-(--accent) text-white cursor-pointer"
                                            >
                                                <Check className="w-3.5 h-3.5" />
                                                Accept
                                            </button>
                                        </div>
                                    ) : pendingRequest ? (
                                        <button
                                            onClick={handleCancelRequest}
                                            disabled={isLoading}
                                            className="flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium bg-(--muted)/10 text-(--muted) border border-(--border) cursor-pointer"
                                        >
                                            <Clock className="w-3.5 h-3.5" />
                                            Pending
                                        </button>
                                    ) : (
                                        <button
                                            onClick={handleJoinRequest}
                                            disabled={isLoading}
                                            className="flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium bg-(--accent) text-white cursor-pointer"
                                        >
                                            <UserRoundPlus className="w-3.5 h-3.5" />
                                            Join
                                        </button>
                                    )}
                                </div>
                            </div>
                        </Container>
                    </motion.div>
                )}
            </AnimatePresence>

            {/* Main Header */}
            <div ref={headerRef} className="w-full border-b border-(--border)">
                <Container>
                    <div className="py-8">
                        {/* Top Section: Group Image, Title, Actions */}
                        <div className="flex flex-col sm:flex-row gap-3 items-start sm:items-center mb-6">
                            {/* Group Image */}
                            <div className="relative">
                                <div className="w-24 h-24 sm:w-28 sm:h-28 rounded-2xl overflow-hidden bg-(--muted)/10 border-2 border-(--border) ring-4 ring-background shadow-lg">
                                    {group.group_image_url ? (
                                        <img
                                            src={group.group_image_url}
                                            alt={group.group_title}
                                            className="w-full h-full object-cover"
                                        />
                                    ) : (
                                        <div className="w-full h-full flex items-center justify-center bg-linear-to-br from-gray-100 to-gray-200">
                                            <Users className="w-12 h-12 text-(--muted)" />
                                        </div>
                                    )}
                                </div>
                            </div>

                            {/* Title & Actions */}
                            <div className="flex-1 min-w-0 flex flex-col sm:flex-row justify-between items-start gap-4">
                                <div className="flex-1 min-w-0">
                                    <div className="flex items-center gap-3 mb-2">
                                        {isOwner && (
                                            <span className="inline-flex items-center gap-1 px-1 py-0.5 rounded-full text-[10px] bg-(--accent) text-white shadow-sm">
                                                {/* <Shield className="w-3 h-3" /> */}
                                                Owner
                                            </span>
                                        )}
                                        {!isOwner && isMember && (
                                            <span className="inline-flex items-center gap-1 px-1 py-0.5 rounded-full text-xs bg-green-500 text-white shadow-sm">
                                                Member
                                            </span>
                                        )}
                                    </div>
                                    <div className="flex items-center gap-3 mb-2">
                                        <h1 className="text-2xl sm:text-3xl font-bold text-foreground tracking-tight">
                                            {group.group_title}
                                        </h1>

                                    </div>

                                    {(isMember || isOwner) ? (
                                        <button
                                            onClick={handleOpenMembersModal}
                                            className="flex items-center gap-2 text-(--muted) hover:text-(--accent) transition-colors cursor-pointer"
                                        >
                                            <Users className="w-4 h-4" />
                                            <span className="text-base">
                                                {membersCount || 0} {membersCount === 1 ? "Member" : "Members"}
                                            </span>
                                        </button>
                                    ) : (
                                        <div className="flex items-center gap-2 text-(--muted)">
                                            <Users className="w-4 h-4" />
                                            <span className="text-base">
                                                {membersCount || 0} {membersCount === 1 ? "Member" : "Members"}
                                            </span>
                                        </div>
                                    )}
                                </div>

                                {/* Action Buttons */}
                                <div className="flex items-center gap-2 shrink-0 flex-wrap">
                                    {isOwner ? (
                                        <>
                                            {/* Pending Requests */}
                                            <Tooltip content="View Requests">
                                                <button
                                                    onClick={handleOpenPendingModal}
                                                    className="flex items-center gap-2 px-4 py-2 rounded-full text-sm font-medium border border-(--border) text-foreground hover:bg-(--muted)/5 transition-colors cursor-pointer"
                                                >
                                                    <UserCheck className="w-4 h-4" />
                                                    {pendingCount && (
                                                        <span className="absolute -top-0.5 -right-0.5 min-w-4 h-4 sm:min-w-[18px] sm:h-[18px] px-1 text-[9px] sm:text-[10px] font-bold text-white bg-red-500 rounded-full flex items-center justify-center border-2 border-background">
                                                            {pendingCount}
                                                        </span>
                                                    )}
                                                </button>
                                            </Tooltip>
                                            {/* Invite Members */}
                                            <Tooltip content="Invite members">
                                                <button
                                                    onClick={handleInviteMembers}
                                                    className="flex items-center gap-2 px-4 py-2 rounded-full text-sm font-medium bg-(--accent) text-white hover:bg-(--accent-hover) shadow-lg shadow-(--accent)/20 transition-colors cursor-pointer"
                                                >
                                                    <UserPlus className="w-4 h-4" />
                                                </button>
                                            </Tooltip>
                                            {/* Settings/Edit Group */}
                                            <Tooltip content="Settings">
                                                <button
                                                    onClick={() => setShowUpdateModal(true)}
                                                    className="flex items-center gap-2 px-4 py-2 rounded-full text-sm font-medium border border-(--border) text-foreground hover:bg-(--muted)/5 transition-colors cursor-pointer"
                                                >
                                                    <Settings className="w-4 h-4" />
                                                </button>
                                            </Tooltip>
                                        </>
                                    ) : isMember ? (
                                        <>
                                            {/* Invite Members */}
                                            <Tooltip content="Invite Members">
                                                <button
                                                    onClick={handleInviteMembers}
                                                    className="flex items-center gap-2 px-4 py-2 rounded-full text-sm font-medium border border-(--accent) text-(--accent) hover:bg-(--accent)/5 transition-colors cursor-pointer"
                                                >
                                                    <UserPlus className="w-4 h-4" />
                                                </button>
                                            </Tooltip>
                                            {/* Leave Group */}
                                            <Tooltip content="Leave">
                                                <button
                                                    onClick={() => setShowLeaveModal(true)}
                                                    className="flex items-center gap-2 px-4 py-2 rounded-full text-sm font-medium border border-(--border) text-(--muted) hover:bg-red-500/10 hover:text-red-500 hover:border-red-500/20 transition-colors cursor-pointer"
                                                >
                                                    <LogOut className="w-4 h-4" />
                                                </button>
                                            </Tooltip>
                                        </>
                                    ) : pendingInvite ? (
                                        /* Non-member with pending invite: Accept/Decline buttons */
                                        <div className="flex flex-col items-end gap-1">
                                            <span className="text-[15px] text-(--muted)">You have been invited</span>
                                            <div className="flex items-center gap-2">
                                                <Tooltip content="Decline">
                                                    <button
                                                        onClick={() => handleInviteResponse(false)}
                                                        disabled={isLoading}
                                                        className={`flex items-center gap-2 px-4 py-2 rounded-full text-sm font-medium transition-all cursor-pointer ${isLoading
                                                            ? "opacity-70 cursor-wait"
                                                            : "bg-(--muted)/10 text-(--muted) border border-(--border) hover:bg-red-500/10 hover:text-red-500 hover:border-red-500/20"
                                                            }`}
                                                    >
                                                        <X className="w-4 h-4" />
                                                    </button>
                                                </Tooltip>
                                                <Tooltip content="Accept">
                                                    <button
                                                        onClick={() => handleInviteResponse(true)}
                                                        disabled={isLoading}
                                                        className={`flex items-center gap-2 px-4 py-2 rounded-full text-sm font-medium transition-all cursor-pointer ${isLoading
                                                            ? "opacity-70 cursor-wait"
                                                            : "bg-(--accent) text-white hover:bg-(--accent-hover) shadow-lg shadow-(--accent)/20"
                                                            }`}
                                                    >
                                                        <Check className="w-4 h-4" />
                                                    </button>
                                                </Tooltip>
                                            </div>
                                        </div>
                                    ) : pendingRequest ? (
                                        /* Non-member with pending request: Show pending state */
                                        <Tooltip content="Cancel Request">
                                            <button
                                                onClick={handleCancelRequest}
                                                disabled={isLoading}
                                                className={`flex items-center gap-2 px-4 py-2 rounded-full text-sm font-medium transition-all cursor-pointer ${isLoading
                                                    ? "opacity-70 cursor-wait"
                                                    : "bg-(--muted)/10 text-(--muted) border border-(--border) hover:bg-red-500/10 hover:text-red-500 hover:border-red-500/20"
                                                    }`}
                                            >
                                                <Clock className="w-4 h-4" />
                                                <span className="hidden sm:inline">Pending</span>
                                            </button>
                                        </Tooltip>
                                    ) : (
                                        /* Non-member: Request to Join */
                                        <Tooltip content="Request to Join">
                                            <button
                                                onClick={handleJoinRequest}
                                                disabled={isLoading}
                                                className={`flex items-center gap-2 px-4 py-2 rounded-full text-sm font-medium transition-all cursor-pointer ${isLoading
                                                    ? "opacity-70 cursor-wait"
                                                    : "bg-(--accent) text-white hover:bg-(--accent-hover) shadow-lg shadow-(--accent)/20"
                                                    }`}
                                            >
                                                <UserRoundPlus className="w-4 h-4" />
                                            </button>
                                        </Tooltip>
                                    )}
                                </div>
                            </div>
                        </div>

                        {/* Description Section */}
                        {group.group_description && (
                            <div className="mb-6">
                                <p className="text-(--foreground)/90 leading-relaxed whitespace-pre-wrap text-[15px]">
                                    {group.group_description}
                                </p>
                            </div>
                        )}
                    </div>
                </Container>
            </div>

            {/* Leave Group Modal */}
            <Modal
                isOpen={showLeaveModal}
                onClose={() => setShowLeaveModal(false)}
                title="Leave Group?"
                description={`Are you sure you want to leave "${group.group_title}"? You will need to request to join again if you change your mind.`}
                onConfirm={handleLeaveGroup}
                confirmText="Leave Group"
                cancelText="Cancel"
                isLoading={isLoading}
                loadingText="Leaving..."
            />

            {/* Invite Members Modal */}
            {showInviteModal && (
                <Modal
                    isOpen={showInviteModal}
                    onClose={handleCloseInviteModal}
                    title="Invite Members"
                    description={inviteSuccess ? "" : "Select followers to invite to this group."}
                    onConfirm={inviteSuccess ? undefined : handleSendInvites}
                    confirmText={`Send Invites${selectedUsers.length > 0 ? ` (${selectedUsers.length})` : ""}`}
                    cancelText="Cancel"
                    isLoading={isInviting}
                    loadingText="Sending..."
                >
                    {inviteSuccess ? (
                        <div className="py-8 text-center">
                            <div className="w-12 h-12 mx-auto mb-4 rounded-full bg-(--accent)/5 flex items-center justify-center">
                                <Check className="w-6 h-6 text-(--accent)" />
                            </div>
                            <p className="text-(--accent) font-medium">Users have been successfully invited!</p>
                        </div>
                    ) : isLoadingFollowers ? (
                        <div className="py-8 flex flex-col items-center gap-3">
                            <Loader2 className="w-6 h-6 text-(--accent) animate-spin" />
                            <p className="text-sm text-(--muted)">Loading followers...</p>
                        </div>
                    ) : followers.length === 0 ? (
                        <div className="py-8 text-center text-(--muted)">
                            <p>No followers to invite.</p>
                        </div>
                    ) : (
                        <div
                            ref={scrollContainerRef}
                            onScroll={handleScroll}
                            className="max-h-60 overflow-y-auto -mx-5 px-5"
                        >
                            <div className="space-y-1">
                                {followers.map((follower) => {
                                    const isSelected = selectedUsers.includes(follower.id);
                                    return (
                                        <button
                                            key={follower.id}
                                            type="button"
                                            onClick={() => toggleUserSelection(follower.id)}
                                            className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-xl transition-colors cursor-pointer ${isSelected
                                                ? "bg-(--accent)/10 border border-(--accent)/30"
                                                : "hover:bg-(--muted)/5 border border-transparent"
                                                }`}
                                        >
                                            <Link
                                                href={`/profile/${follower.id}`}
                                                onClick={(e) => e.stopPropagation()}
                                                className="w-10 h-10 rounded-full bg-(--muted)/10 flex items-center justify-center overflow-hidden shrink-0 hover:ring-2 hover:ring-(--accent)/50 transition-all"
                                            >
                                                {follower.avatar_url ? (
                                                    <img
                                                        src={follower.avatar_url}
                                                        alt={follower.username}
                                                        className="w-full h-full object-cover"
                                                    />
                                                ) : (
                                                    <User className="w-5 h-5 text-(--muted)" />
                                                )}
                                            </Link>
                                            <Link
                                                href={`/profile/${follower.id}`}
                                                onClick={(e) => e.stopPropagation()}
                                                className="flex-1 min-w-0 text-left"
                                            >
                                                <p className="text-sm font-medium text-foreground truncate hover:text-(--accent) transition-colors">
                                                    {follower.username}
                                                </p>
                                            </Link>
                                            <div className={`w-5 h-5 rounded-full border-2 flex items-center justify-center shrink-0 transition-colors ${isSelected
                                                ? "bg-(--accent) border-(--accent)"
                                                : "border-(--border)"
                                                }`}>
                                                {isSelected && <Check className="w-3 h-3 text-white" />}
                                            </div>
                                        </button>
                                    );
                                })}
                                {isLoadingMore && (
                                    <div className="py-3 flex justify-center">
                                        <Loader2 className="w-5 h-5 text-(--accent) animate-spin" />
                                    </div>
                                )}
                            </div>
                        </div>
                    )}
                </Modal>
            )}

            {/* Update Group Modal */}
            <UpdateGroupModal
                isOpen={showUpdateModal}
                onClose={() => setShowUpdateModal(false)}
                onSuccess={handleUpdateSuccess}
                group={group}
            />

            {/* Pending Requests Modal */}
            {showPendingModal && (
                <Modal
                    isOpen={showPendingModal}
                    onClose={handleClosePendingModal}
                    title="Join Requests"
                    description="Review pending requests to join this group."
                    cancelText="Close"
                >
                    {isLoadingPending ? (
                        <div className="py-8 flex flex-col items-center gap-3">
                            <Loader2 className="w-6 h-6 text-(--accent) animate-spin" />
                            <p className="text-sm text-(--muted)">Loading requests...</p>
                        </div>
                    ) : pendingUsers.length === 0 ? (
                        <div className="py-8 text-center text-(--muted)">
                            <p>No pending requests.</p>
                        </div>
                    ) : (
                        <div
                            ref={pendingScrollRef}
                            onScroll={handlePendingScroll}
                            className="h-20 overflow-y-auto overflow-x-visible -mx-5 px-5"
                        >
                            <div className="space-y-1">
                                {pendingUsers.map((pendingUser) => {
                                    const isProcessing = processingUser === pendingUser.id;
                                    return (
                                        <div
                                            key={pendingUser.id}
                                            className="flex items-center gap-3 px-3 py-2.5 rounded-xl border border-transparent hover:bg-(--muted)/5"
                                        >
                                            <Link
                                                href={`/profile/${pendingUser.id}`}
                                                className="w-10 h-10 rounded-full bg-(--muted)/10 flex items-center justify-center overflow-hidden shrink-0 hover:ring-2 hover:ring-(--accent)/50 transition-all"
                                            >
                                                {pendingUser.avatar_url ? (
                                                    <img
                                                        src={pendingUser.avatar_url}
                                                        alt={pendingUser.username}
                                                        className="w-full h-full object-cover"
                                                    />
                                                ) : (
                                                    <User className="w-5 h-5 text-(--muted)" />
                                                )}
                                            </Link>
                                            <Link
                                                href={`/profile/${pendingUser.id}`}
                                                className="flex-1 min-w-0 text-left"
                                            >
                                                <p className="text-sm font-medium text-foreground truncate hover:text-(--accent) transition-colors">
                                                    {pendingUser.username}
                                                </p>
                                            </Link>
                                            <div className="flex items-center gap-1">
                                                <Tooltip content="Decline">
                                                    <button
                                                        onClick={() => handleRequest(pendingUser.id, false)}
                                                        disabled={isProcessing}
                                                        className={`p-2 rounded-full transition-colors cursor-pointer ${isProcessing
                                                            ? "opacity-50 cursor-wait"
                                                            : "text-(--muted) hover:bg-red-500/10 hover:text-red-500"
                                                            }`}
                                                    >
                                                        <X className="w-4 h-4" />
                                                    </button>
                                                </Tooltip>
                                                <Tooltip content="Accept">
                                                    <button
                                                        onClick={() => handleRequest(pendingUser.id, true)}
                                                        disabled={isProcessing}
                                                        className={`p-2 rounded-full transition-colors cursor-pointer ${isProcessing
                                                            ? "opacity-50 cursor-wait"
                                                            : "text-(--muted) hover:bg-green-500/10 hover:text-green-500"
                                                            }`}
                                                    >
                                                        <Check className="w-4 h-4" />
                                                    </button>
                                                </Tooltip>
                                            </div>
                                        </div>
                                    );
                                })}
                                {isLoadingMorePending && (
                                    <div className="py-3 flex justify-center">
                                        <Loader2 className="w-5 h-5 text-(--accent) animate-spin" />
                                    </div>
                                )}
                            </div>
                        </div>
                    )}
                </Modal>
            )}

            {/* Members Modal */}
            {showMembersModal && (
                <Modal
                    isOpen={showMembersModal}
                    onClose={handleCloseMembersModal}
                    title="Group Members"
                    description={`${membersCount || 0} ${membersCount === 1 ? "member" : "members"} in this group.`}
                    cancelText="Close"
                >
                    {isLoadingMembers ? (
                        <div className="py-8 flex flex-col items-center gap-3">
                            <Loader2 className="w-6 h-6 text-(--accent) animate-spin" />
                            <p className="text-sm text-(--muted)">Loading members...</p>
                        </div>
                    ) : members.length === 0 ? (
                        <div className="py-8 text-center text-(--muted)">
                            <p>No members found.</p>
                        </div>
                    ) : (
                        <div
                            ref={membersScrollRef}
                            onScroll={handleMembersScroll}
                            className="max-h-64 overflow-y-auto -mx-5 px-5"
                        >
                            <div className="space-y-1">
                                {members.map((member) => {
                                    const isRemoving = removingMember === member.user_id;
                                    const isCurrentUser = member.user_id === user?.id;
                                    const isMemberOwner = member.group_role === "owner";
                                    return (
                                        <div
                                            key={member.user_id}
                                            className="flex items-center gap-3 px-3 py-2.5 rounded-xl border border-transparent hover:bg-(--muted)/5"
                                        >
                                            <Link
                                                href={`/profile/${member.user_id}`}
                                                className="w-10 h-10 rounded-full bg-(--muted)/10 flex items-center justify-center overflow-hidden shrink-0 hover:ring-2 hover:ring-(--accent)/50 transition-all"
                                            >
                                                {member.avatar_url ? (
                                                    <img
                                                        src={member.avatar_url}
                                                        alt={member.username}
                                                        className="w-full h-full object-cover"
                                                    />
                                                ) : (
                                                    <User className="w-5 h-5 text-(--muted)" />
                                                )}
                                            </Link>
                                            <Link
                                                href={`/profile/${member.user_id}`}
                                                className="flex-1 min-w-0 text-left"
                                            >
                                                <div className="flex items-center gap-2">
                                                    <p className="text-sm font-medium text-foreground truncate hover:text-(--accent) transition-colors">
                                                        {member.username}
                                                    </p>
                                                    {isMemberOwner && (
                                                        <span className="inline-flex items-center px-1.5 py-0.5 rounded-full text-[10px] bg-(--accent) text-white">
                                                            Owner
                                                        </span>
                                                    )}
                                                </div>
                                            </Link>
                                            {/* Owner can remove members (but not themselves or other owners) */}
                                            {isOwner && !isCurrentUser && !isMemberOwner && (
                                                <Tooltip content="Remove member">
                                                    <button
                                                        onClick={() => handleRemoveMember(member.user_id)}
                                                        disabled={isRemoving}
                                                        className={`p-2 rounded-full transition-colors cursor-pointer ${isRemoving
                                                            ? "opacity-50 cursor-wait"
                                                            : "text-(--muted) hover:bg-red-500/10 hover:text-red-500"
                                                            }`}
                                                    >
                                                        {isRemoving ? (
                                                            <Loader2 className="w-4 h-4 animate-spin" />
                                                        ) : (
                                                            <Trash2 className="w-4 h-4" />
                                                        )}
                                                    </button>
                                                </Tooltip>
                                            )}
                                        </div>
                                    );
                                })}
                                {isLoadingMoreMembers && (
                                    <div className="py-3 flex justify-center">
                                        <Loader2 className="w-5 h-5 text-(--accent) animate-spin" />
                                    </div>
                                )}
                            </div>
                        </div>
                    )}
                </Modal>
            )}
        </>
    );
}
