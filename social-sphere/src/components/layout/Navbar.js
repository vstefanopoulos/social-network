"use client";

import { usePathname, useRouter } from "next/navigation";
import { Activity, Users, Send, Bell, User, LogOut, Settings, HeartPulse, Search, Loader2, MessageCircle } from "lucide-react";
import { useState, useRef, useEffect, useCallback } from "react";
import Tooltip from "@/components/ui/Tooltip";
import Link from "next/link";
import { useStore } from "@/store/store";
import { logout } from "@/actions/auth/logout";
import { SearchUsers } from "@/actions/search/search-users";
import { getImageUrl } from "@/actions/auth/get-image-url";
import { getConv } from "@/actions/chat/get-conv";
import { useLiveSocket } from "@/context/LiveSocketContext";
import { getUnreadCount } from "@/actions/chat/get-unread-count";
import { getConvByID } from "@/actions/chat/get-conv-by-id";
import { getNotifCount } from "@/actions/notifs/get-notif-count";

export default function Navbar() {
    const pathname = usePathname();
    const router = useRouter();
    const [isDropdownOpen, setIsDropdownOpen] = useState(false);
    const [isMessagesOpen, setIsMessagesOpen] = useState(false);
    const [conversations, setConversations] = useState([]);
    const [isLoadingConversations, setIsLoadingConversations] = useState(false);
    const [isAlerting, setIsAlerting] = useState(false);
    const [isNotifAlerting, setIsNotifAlerting] = useState(false);
    const dropdownRef = useRef(null);
    const messagesRef = useRef(null);
    const searchRef = useRef(null);
    const audioContextRef = useRef(null);
    const audioBufferRef = useRef(null);
    const notifAudioBufferRef = useRef(null);
    const alertTimeoutRef = useRef(null);
    const notifAlertTimeoutRef = useRef(null);
    const pendingSoundRef = useRef(false);
    const pendingNotifSoundRef = useRef(false);
    const user = useStore((state) => state.user);
    const setUser = useStore((state) => state.setUser);
    const unreadCount = useStore((state) => state.unreadCount);
    const setUnreadCount = useStore((state) => state.setUnreadCount);
    const incrementUnreadCount = useStore((state) => state.incrementUnreadCount);
    const decrementUnreadCount = useStore((state) => state.decrementUnreadCount);
    const unreadNotifs = useStore((state) => state.unreadNotifs);
    const setUnreadNotifs = useStore((state) => state.setUnreadNotifs);
    const incrementNotifs = useStore((state) => state.incrementNotifs);

    // Live WebSocket connection
    const {
        addOnPrivateMessage,
        removeOnPrivateMessage,
        addOnNotification,
        removeOnNotification,
        disconnect: disconnectSocket,
    } = useLiveSocket();

    useEffect(() => {
        const getUnread = async () => {
            const result = await getUnreadCount();
            setUnreadCount(result.data?.count ?? 0);

            const unreadNotifs = await getNotifCount();
            setUnreadNotifs(unreadNotifs?.data.value ?? 0);
        }
        getUnread();
    }, [setUnreadCount]);

    // Initialize Web Audio API for message and notification alerts
    useEffect(() => {
        const initAudio = async () => {
            try {
                audioContextRef.current = new (window.AudioContext || window.webkitAudioContext)();

                // Try to resume immediately (works if user interacted with site recently)
                if (audioContextRef.current.state === "suspended") {
                    audioContextRef.current.resume().catch(() => {});
                }

                // Load private message sound
                const msgResponse = await fetch("/alerts/privateMessage.mp3");
                const msgArrayBuffer = await msgResponse.arrayBuffer();
                audioBufferRef.current = await audioContextRef.current.decodeAudioData(msgArrayBuffer);

                // Load notification sound
                const notifResponse = await fetch("/alerts/notification.mp3");
                const notifArrayBuffer = await notifResponse.arrayBuffer();
                notifAudioBufferRef.current = await audioContextRef.current.decodeAudioData(notifArrayBuffer);
            } catch (err) {
                console.error("Failed to load audio:", err);
            }
        };

        initAudio();

        // Resume AudioContext on first user interaction and play pending sounds
        const unlockAudio = async () => {
            if (audioContextRef.current?.state === "suspended") {
                await audioContextRef.current.resume();
                // Play pending message sound if there was one
                if (pendingSoundRef.current && audioBufferRef.current) {
                    pendingSoundRef.current = false;
                    const source = audioContextRef.current.createBufferSource();
                    source.buffer = audioBufferRef.current;
                    source.connect(audioContextRef.current.destination);
                    source.start(0);
                }
                // Play pending notification sound if there was one
                if (pendingNotifSoundRef.current && notifAudioBufferRef.current) {
                    pendingNotifSoundRef.current = false;
                    const source = audioContextRef.current.createBufferSource();
                    source.buffer = notifAudioBufferRef.current;
                    source.connect(audioContextRef.current.destination);
                    source.start(0);
                }
            }
        };

        document.addEventListener("click", unlockAudio);
        document.addEventListener("touchstart", unlockAudio);

        return () => {
            document.removeEventListener("click", unlockAudio);
            document.removeEventListener("touchstart", unlockAudio);
            if (alertTimeoutRef.current) {
                clearTimeout(alertTimeoutRef.current);
            }
            if (notifAlertTimeoutRef.current) {
                clearTimeout(notifAlertTimeoutRef.current);
            }
            if (audioContextRef.current) {
                audioContextRef.current.close();
            }
        };
    }, []);

    // Reset UnreadCount in local state when viewing a conversation via URL
    useEffect(() => {
        if (pathname.startsWith('/messages/')) {
            const viewingId = pathname.split('/messages/')[1];
            if (viewingId) {
                setConversations((prev) =>
                    prev.map((c) =>
                        c.Interlocutor?.id === viewingId
                            ? { ...c, UnreadCount: 0 }
                            : c
                    )
                );
            }
        }
    }, [pathname]);

    // Handle incoming private messages - update conversations in real-time
    const handleNewMessage = useCallback(async (msg) => {
        console.log("New message received:", msg);

        // if the message is not mine, sound alert and glow
        if (msg.sender.id !== user.id) {
            // Play sound using Web Audio API
            if (audioContextRef.current && audioBufferRef.current) {
                if (audioContextRef.current.state === "suspended") {
                    // Queue sound to play after user interaction
                    pendingSoundRef.current = true;
                } else {
                    const source = audioContextRef.current.createBufferSource();
                    source.buffer = audioBufferRef.current;
                    source.connect(audioContextRef.current.destination);
                    source.start(0);
                }
            }

            // Trigger glow for 3 seconds
            if (alertTimeoutRef.current) {
                clearTimeout(alertTimeoutRef.current);
            }
            setIsAlerting(true);
            alertTimeoutRef.current = setTimeout(() => {
                setIsAlerting(false);
            }, 4000);
        }

        const senderId = msg.sender?.id;

        // Check if conversation exists in current state
        const existingConv = conversations.find(
            (conv) => conv.Interlocutor?.id === senderId
        );

        if (existingConv) {
            // Increment unread count if this conversation had no unread messages before
            const isViewingConv = window.location.pathname === `/messages/${existingConv.Interlocutor?.id}`;
            if (existingConv.UnreadCount === 0 && !isViewingConv || !existingConv.UnreadCount && !isViewingConv) {
                incrementUnreadCount();
            }

            // Update existing conversation
            setConversations((prev) => {
                const existingIndex = prev.findIndex(
                    (conv) => conv.Interlocutor?.id === senderId
                );
                if (existingIndex === -1) return prev;

                const updated = [...prev];
                updated[existingIndex] = {
                    ...updated[existingIndex],
                    LastMessage: {
                        id: msg.id,
                        message_text: msg.message_text,
                        sender: msg.sender,
                    },
                    UpdatedAt: msg.created_at,
                    UnreadCount:
                        msg.sender?.id !== user?.id
                            ? updated[existingIndex].UnreadCount + 1
                            : updated[existingIndex].UnreadCount,
                };
                // Sort by most recent
                return updated.sort((a, b) => new Date(b.UpdatedAt) - new Date(a.UpdatedAt));
            });
        } else {
            if (msg.sender.id === user.id) {
                return;
            }
            // New conversation - increment unread count
            incrementUnreadCount();

            const result = await getConvByID({
                interlocutorId: msg.sender.id,
                convId: msg.conversation_id,
            });

            const newConv = {
                ConversationId: msg.conversation_id,
                Interlocutor: result?.data?.Interlocutor || null,
                LastMessage: {
                    id: msg.id,
                    message_text: msg.message_text,
                    sender: msg.sender,
                },
                UpdatedAt: msg.created_at,
                UnreadCount: 1,
            };

            setConversations((prev) => [newConv, ...prev]);
        }
    }, [user?.id, conversations, incrementUnreadCount]);

    // Register message handler
    useEffect(() => {
        addOnPrivateMessage(handleNewMessage);
        return () => removeOnPrivateMessage(handleNewMessage);
    }, [addOnPrivateMessage, removeOnPrivateMessage, handleNewMessage]);

    // Handle incoming notifications - play sound and glow
    const handleNewNotification = useCallback(() => {
        // Increment notification count
        incrementNotifs();

        // Play notification sound using Web Audio API
        if (audioContextRef.current && notifAudioBufferRef.current) {
            if (audioContextRef.current.state === "suspended") {
                // Queue sound to play after user interaction
                pendingNotifSoundRef.current = true;
            } else {
                const source = audioContextRef.current.createBufferSource();
                source.buffer = notifAudioBufferRef.current;
                source.connect(audioContextRef.current.destination);
                source.start(0);
            }
        }

        // Trigger glow for 4 seconds
        if (notifAlertTimeoutRef.current) {
            clearTimeout(notifAlertTimeoutRef.current);
        }
        setIsNotifAlerting(true);
        notifAlertTimeoutRef.current = setTimeout(() => {
            setIsNotifAlerting(false);
        }, 4000);
    }, [incrementNotifs]);

    // Register notification handler
    useEffect(() => {
        addOnNotification(handleNewNotification);
        return () => removeOnNotification(handleNewNotification);
    }, [addOnNotification, removeOnNotification, handleNewNotification]);

    // Avatar state - allows refreshing when original expires
    const [avatarSrc, setAvatarSrc] = useState(user?.avatar_url);

    // Sync avatarSrc when user changes (e.g., after profile update)
    useEffect(() => {
        setAvatarSrc(user?.avatar_url);
    }, [user?.avatar_url]);

    // Search State
    const [searchQuery, setSearchQuery] = useState("");
    const [searchResults, setSearchResults] = useState([]);
    const [isSearching, setIsSearching] = useState(false);
    const [showSearchResults, setShowSearchResults] = useState(false);

    // Handle image error - fetch fresh variant URL when original expires
    const handleImageError = async () => {
        if (!user?.fileId) return;

        const result = await getImageUrl({ fileId: user.fileId, variant: "thumb" });

        if (!result?.success || !result.data?.download_url) return;

        // Update local state for immediate UI update
        setAvatarSrc(result.data.download_url);

        // Update store so it persists
        setUser({
            ...user,
            avatar_url: result.data.download_url
        });
    };

    // Close dropdowns when clicking outside
    useEffect(() => {
        function handleClickOutside(event) {
            if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
                setIsDropdownOpen(false);
            }
            if (messagesRef.current && !messagesRef.current.contains(event.target)) {
                setIsMessagesOpen(false);
            }
            if (searchRef.current && !searchRef.current.contains(event.target)) {
                setShowSearchResults(false);
            }
        }

        document.addEventListener("mousedown", handleClickOutside);
        return () => {
            document.removeEventListener("mousedown", handleClickOutside);
        };
    }, []);

    // Fetch conversations when messages dropdown opens
    const handleMessagesClick = async () => {
        const willOpen = !isMessagesOpen;
        setIsMessagesOpen(willOpen);

        if (willOpen) {
            setIsLoadingConversations(true);
            try {
                const result = await getConv({ first: true, limit: 5 });
                if (result.success && result.data) {
                    setConversations(result.data);
                }
            } catch (error) {
                console.error("Error fetching conversations:", error);
            } finally {
                setIsLoadingConversations(false);
            }
        }
    };

    const clicked = (conv) => {
        if (hasUnreadMessages(conv)) {
            decrementUnreadCount();
            // Reset UnreadCount in local state so future messages increment correctly
            setConversations((prev) =>
                prev.map((c) =>
                    c.ConversationId === conv.ConversationId
                        ? { ...c, UnreadCount: 0 }
                        : c
                )
            );
        }
        setIsMessagesOpen(false);
        router.push(`/messages/${conv.Interlocutor.id}`);
    }

    // Format relative time
    const formatRelativeTime = (dateString) => {
        const date = new Date(dateString);
        const now = new Date();
        const diffMs = now - date;
        const diffMins = Math.floor(diffMs / 60000);
        const diffHours = Math.floor(diffMs / 3600000);
        const diffDays = Math.floor(diffMs / 86400000);

        if (diffMins < 1) return "now";
        if (diffMins < 60) return `${diffMins}m`;
        if (diffHours < 24) return `${diffHours}h`;
        if (diffDays < 7) return `${diffDays}d`;
        return date.toLocaleDateString();
    };

    // Truncate message text
    const truncateMessage = (text, maxLength = 30) => {
        if (!text) return "";
        return text.length > maxLength ? text.substring(0, maxLength) + "..." : text;
    };

    // Check if conversation has unread messages for current user
    // Only unread if last message sender is NOT me (someone else sent it)
    const hasUnreadMessages = (conv) => {
        return conv.UnreadCount > 0 && conv.LastMessage?.sender?.id !== user?.id;
    };

    // Debounced Search
    useEffect(() => {
        const timer = setTimeout(async () => {
            if (searchQuery.trim().length >= 2) {
                setIsSearching(true);
                try {
                    const response = await SearchUsers({ query: searchQuery, limit: 5 });
                    if (response && response.data) {
                        setSearchResults(response.data.users);
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

    const handleResultClick = (userId, e) => {
        e.preventDefault();
        e.stopPropagation();
        setShowSearchResults(false);
        setSearchQuery("");
        router.push(`/profile/${userId}`);
    };

    const handleLogout = async () => {
        try {
            // Disconnect WebSocket before logout
            disconnectSocket();

            // Logout and redirect happens on the server
            await logout();
        } catch (error) {
            // redirect() in server actions throws a NEXT_REDIRECT error
            // which is handled by Next.js, so we can safely ignore it here
            if (!error?.message?.includes('NEXT_REDIRECT')) {
                console.error('Logout error:', error);
            }
        }
    }

    if (!user) {
        return <div></div>;
    }

    const navItems = [
        {
            label: "Public",
            href: "/feed/public",
            icon: Activity,
        },
        {
            label: "Friends",
            href: "/feed/friends",
            icon: HeartPulse,
        },
        {
            label: "Groups",
            href: "/groups",
            icon: Users,
        },
    ];

    const isActive = (path) => pathname === path;

    return (
        <nav className="sticky top-0 z-50 w-full border-b border-(--border) bg-(--background)/95 backdrop-blur-md">
            <div className="w-full px-4 sm:px-6 lg:px-8">
                <div className="flex items-center justify-between h-16 gap-2 sm:gap-3">
                    {/* Left Section: Logo */}
                    <div className="flex items-center shrink-0">
                        <Link
                            href="/feed/public"
                            className="flex items-center"
                            prefetch={false}
                        >
                            <span className="text-sm sm:text-base font-medium tracking-tight text-foreground hover:text-(--muted) transition-colors">
                                SocialSphere
                            </span>
                        </Link>
                    </div>

                    {/* Center Section: Search Bar - Grows to fill available space */}
                    <div className="hidden md:flex flex-1 max-w-md mx-4 ml-50" ref={searchRef}>
                        <div className="relative w-full group">
                            <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                                {isSearching ? (
                                    <Loader2 className="h-4 w-4 text-(--muted) animate-spin" />
                                ) : (
                                    <Search className="h-4 w-4 text-(--muted) group-focus-within:text-(--accent) transition-colors" />
                                )}
                            </div>
                            <input
                                type="text"
                                value={searchQuery}
                                onChange={(e) => setSearchQuery(e.target.value)}
                                onFocus={() => {
                                    if (searchResults.length > 0) setShowSearchResults(true);
                                }}
                                className="block w-full pl-11 pr-4 py-2.5 border border-(--border) rounded-full text-sm bg-(--muted)/5 text-foreground placeholder-(--muted) hover:border-foreground focus:outline-none focus:border-(--accent) focus:ring-2 focus:ring-(--accent)/10 transition-all"
                                placeholder="Search users..."
                            />

                            {/* Search Results Dropdown */}
                            {showSearchResults && (
                                <div className="absolute top-full left-0 right-0 mt-2 bg-background border border-(--border) rounded-2xl shadow-xl overflow-hidden animate-in fade-in zoom-in-95 duration-200 max-h-96 overflow-y-auto z-200">
                                    {searchResults.length > 0 ? (
                                        <div className="py-2">
                                            {searchResults.map((result) => (
                                                <button
                                                    type="button"
                                                    key={result.id}
                                                    onMouseDown={(e) => handleResultClick(result.id, e)}
                                                    className="w-full flex items-center gap-3 px-4 py-3 hover:bg-(--muted)/5 transition-colors cursor-pointer text-left"
                                                >
                                                    <div className="w-10 h-10 rounded-full bg-(--muted)/10 flex items-center justify-center overflow-hidden shrink-0">
                                                        {result.avatar_url ? (
                                                            <img src={result.avatar_url} alt={result.username || "User"} className="w-full h-full object-cover" />
                                                        ) : (
                                                            <User className="w-5 h-5 text-(--muted)" />
                                                        )}
                                                    </div>
                                                    <div className="flex-1 min-w-0">
                                                        <p className="text-sm font-medium text-foreground truncate">
                                                            {result.username}
                                                        </p>
                                                    </div>
                                                </button>
                                            ))}
                                        </div>
                                    ) : (
                                        <div className="p-4 text-center text-sm text-(--muted)">
                                            No users found
                                        </div>
                                    )}
                                </div>
                            )}
                        </div>
                    </div>

                    {/* Right Section: Nav + Actions */}
                    <div className="flex items-center gap-1 sm:gap-1.5 shrink-0">
                        {/* Desktop Navigation */}
                        <div className="hidden lg:flex items-center gap-1">
                            {navItems.map((item) => {
                                const Icon = item.icon;
                                const active = isActive(item.href);
                                return (
                                    <Tooltip key={item.href} content={item.label}>
                                        <Link
                                            href={item.href}
                                            prefetch={false}
                                            className={`flex items-center gap-2 px-3 py-2 rounded-full text-sm font-medium transition-all ${active
                                                ? "bg-(--accent)/10 text-(--accent)"
                                                : "text-(--muted) hover:text-foreground hover:bg-(--muted)/10"
                                                }`}
                                        >
                                            <Icon className="w-[18px] h-[18px]" strokeWidth={active ? 2.5 : 2} />
                                        </Link>
                                    </Tooltip>
                                );
                            })}
                        </div>

                        {/* Tablet/Mobile Navigation - Icon only */}
                        <div className="flex lg:hidden items-center gap-0.5">
                            {navItems.map((item) => {
                                const Icon = item.icon;
                                const active = isActive(item.href);
                                return (
                                    <Tooltip key={item.href} content={item.label}>
                                        <Link
                                            href={item.href}
                                            prefetch={false}
                                            className={`p-2 sm:p-2.5 rounded-full transition-all ${active
                                                ? "bg-(--accent)/10 text-(--accent)"
                                                : "text-(--muted) hover:text-foreground hover:bg-(--muted)/10"
                                                }`}
                                        >
                                            <Icon className="w-4 h-4 sm:w-5 sm:h-5" strokeWidth={active ? 2.5 : 2} />
                                        </Link>
                                    </Tooltip>
                                );
                            })}
                        </div>

                        {/* Divider */}
                        <div className="h-6 w-px bg-(--border) mx-0.5 sm:mx-1" />

                        {/* Messages Dropdown */}
                        <div className="relative" ref={messagesRef}>
                            <Tooltip content={"Messages"} active={!isMessagesOpen}>
                                <button
                                    onClick={handleMessagesClick}
                                    className={`relative p-2 sm:p-2.5 rounded-full transition-all cursor-pointer ${isMessagesOpen
                                        ? "bg-(--accent)/10 text-(--accent)"
                                        : "text-(--muted) hover:text-foreground hover:bg-(--muted)/10"
                                        }`}
                                >
                                    <Send className={`w-4 h-4 sm:w-5 sm:h-5`} strokeWidth={isMessagesOpen ? 2.5 : 2} />
                                    {unreadCount > 0 && (
                                        <span className={`absolute top-0.5 right-0.5 min-w-1 h-1 sm:min-w-1.5 sm:h-1.5  bg-(--accent) rounded-full flex items-center justify-center border-2 border-(--accent) ${isAlerting ? "animate-pulse-glow" : ""}`}>
                                        </span>
                                    )}
                                </button>
                            </Tooltip>

                            {/* Messages Dropdown Menu */}
                            {isMessagesOpen && (
                                <div className="absolute right-0 top-full mt-3 w-80 sm:w-96 rounded-2xl border border-(--border) bg-background shadow-xl overflow-hidden animate-in fade-in zoom-in-95 duration-200 z-100">
                                    <div className="p-4 border-b border-(--border)">
                                        <h3 className="text-sm font-semibold text-foreground">Messages</h3>
                                    </div>

                                    <div className="max-h-80 overflow-y-auto">
                                        {isLoadingConversations ? (
                                            <div className="flex items-center justify-center py-8">
                                                <Loader2 className="w-5 h-5 text-(--muted) animate-spin" />
                                            </div>
                                        ) : conversations.length > 0 ? (
                                            <div className="py-1">
                                                {conversations.map((conv) => (
                                                    <button
                                                        key={conv.ConversationId}
                                                        onClick={() => {
                                                            clicked(conv);
                                                        }}
                                                        className="w-full flex items-start gap-3 px-4 py-3 hover:bg-(--muted)/5 transition-colors cursor-pointer text-left"
                                                    >
                                                        {/* Avatar */}
                                                        <div className="relative shrink-0">
                                                            <div className="w-11 h-11 rounded-full bg-(--muted)/10 flex items-center justify-center overflow-hidden border border-(--border)">
                                                                {conv.Interlocutor?.avatar_url ? (
                                                                    <img
                                                                        src={conv.Interlocutor.avatar_url}
                                                                        alt={conv.Interlocutor.username || "User"}
                                                                        className="w-full h-full object-cover"
                                                                    />
                                                                ) : (
                                                                    <User className="w-5 h-5 text-(--muted)" />
                                                                )}
                                                            </div>
                                                            {hasUnreadMessages(conv) && (
                                                                <span className="absolute -top-1 -right-1 min-w-5 h-5 px-1.5 text-[10px] font-bold text-white bg-red-500 rounded-full flex items-center justify-center border-2 border-background">
                                                                    {conv.UnreadCount}
                                                                </span>
                                                            )}
                                                        </div>

                                                        {/* Content */}
                                                        <div className="flex-1 min-w-0">
                                                            <div className="flex items-center justify-between gap-2">
                                                                <p className={`text-sm truncate ${hasUnreadMessages(conv) ? "font-semibold text-foreground" : "font-medium text-foreground"}`}>
                                                                    {conv.Interlocutor?.username || "Unknown User"}
                                                                </p>
                                                                <span className="text-xs text-(--muted) shrink-0">
                                                                    {formatRelativeTime(conv.UpdatedAt)}
                                                                </span>
                                                            </div>
                                                            <p className={`text-sm mt-0.5 truncate ${hasUnreadMessages(conv) ? "text-foreground" : "text-(--muted)"}`}>
                                                                {truncateMessage(conv.LastMessage?.message_text)}
                                                            </p>
                                                        </div>
                                                    </button>
                                                ))}
                                            </div>
                                        ) : (
                                            <div className="py-8 text-center">
                                                <MessageCircle className="w-8 h-8 text-(--muted) mx-auto mb-2" />
                                                <p className="text-sm text-(--muted)">No conversations yet</p>
                                            </div>
                                        )}
                                    </div>

                                    {/* See all messages link */}
                                    <div className="border-t border-(--border)">
                                        <Link
                                            href="/messages"
                                            onClick={() => setIsMessagesOpen(false)}
                                            className="flex items-center justify-center gap-2 px-4 py-3 text-sm font-medium text-(--accent) hover:bg-(--muted)/5 transition-colors"
                                        >
                                            See all messages
                                        </Link>
                                    </div>
                                </div>
                            )}
                        </div>

                        {/* Notifications */}
                        <Tooltip content="Notifications">
                            <Link
                                href="/notifications"
                                className={`relative p-2 sm:p-2.5 rounded-full transition-all ${isActive('/notifications')
                                    ? "bg-(--accent)/10 text-(--accent)"
                                    : "text-(--muted) hover:text-foreground hover:bg-(--muted)/10"
                                    }`}
                            >
                                <Bell className="w-4 h-4 sm:w-5 sm:h-5" strokeWidth={isActive('/notifications') ? 2.5 : 2} />
                                {unreadNotifs > 0 && (
                                        <span className={`absolute top-0.5 right-0.5 min-w-1 h-1 sm:min-w-1 sm:h-1 bg-foreground rounded-full flex items-center justify-center border-2 border-foreground ${isNotifAlerting ? "animate-pulse-glow-foreground" : ""}`}>
                                        </span>
                                )}
                            </Link>
                        </Tooltip>

                        {/* User Dropdown */}
                        {user && (
                            <div className="relative ml-0.5 sm:ml-1.5 pl-1 sm:pl-2.5 border-l border-(--border)" ref={dropdownRef}>
                                <button
                                    onClick={() => setIsDropdownOpen(!isDropdownOpen)}
                                    className="flex items-center gap-1.5 sm:gap-2 hover:opacity-80 transition-opacity cursor-pointer"
                                >
                                    <div className="w-7 h-7 sm:w-8 sm:h-8 rounded-full bg-(--muted)/10 border border-(--border) flex items-center justify-center overflow-hidden hover:border-(--accent) transition-colors">
                                        {avatarSrc ? (
                                            <img src={avatarSrc} alt={user.username?.[0] || "U"} className="w-full h-full object-cover" onError={handleImageError} />
                                        ) : (
                                            <User className="w-4 h-4 text-(--muted)" />
                                        )}
                                    </div>
                                    {user.username}
                                    <svg
                                        className={`hidden sm:block w-3.5 h-3.5 text-(--muted) transition-transform ${isDropdownOpen ? "rotate-180" : ""
                                            }`}
                                        fill="none"
                                        stroke="currentColor"
                                        viewBox="0 0 24 24"
                                    >
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                                    </svg>    
                                </button>

                                {/* Dropdown Menu */}
                                {isDropdownOpen && (
                                    <div className="absolute right-0 top-full mt-3 w-52 rounded-2xl border border-(--border) bg-background shadow-xl overflow-hidden animate-in fade-in zoom-in-95 duration-200 z-100">
                                        <div className="p-1.5">
                                            <Link
                                                href={`/profile/${user.id}`}
                                                prefetch={false}
                                                onClick={() => setIsDropdownOpen(false)}
                                                className="flex items-center gap-3 px-3.5 py-2.5 text-sm font-medium rounded-xl hover:bg-(--muted)/10 transition-colors text-foreground"
                                            >
                                                <User className="w-4 h-4 text-(--muted)" />
                                                Profile
                                            </Link>
                                            <Link
                                                href={`/profile/${user.id}/settings`}
                                                prefetch={false}
                                                onClick={() => setIsDropdownOpen(false)}
                                                className="flex items-center gap-3 px-3.5 py-2.5 text-sm font-medium rounded-xl hover:bg-(--muted)/10 transition-colors text-foreground"
                                            >
                                                <Settings className="w-4 h-4 text-(--muted)" />
                                                Settings
                                            </Link>
                                            <div className="h-px bg-(--border) my-1.5" />
                                            <button
                                                onClick={handleLogout}
                                                className="w-full flex items-center gap-3 px-3.5 py-2.5 text-sm font-medium rounded-xl text-red-500 hover:bg-red-500/10 transition-colors text-left cursor-pointer"
                                            >
                                                <LogOut className="w-4 h-4" />
                                                Sign Out
                                            </button>
                                        </div>
                                    </div>
                                )}
                            </div>
                        )}
                    </div>
                </div>

                {/* Mobile Search Bar - Below main nav, fully functional */}
                <div className="md:hidden pb-3 pt-1" ref={searchRef}>
                    <div className="relative w-full group">
                        <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                            {isSearching ? (
                                <Loader2 className="h-4 w-4 text-(--muted) animate-spin" />
                            ) : (
                                <Search className="h-4 w-4 text-(--muted) group-focus-within:text-(--accent) transition-colors" />
                            )}
                        </div>
                        <input
                            type="text"
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            onFocus={() => {
                                if (searchResults.length > 0) setShowSearchResults(true);
                            }}
                            className="block w-full pl-11 pr-4 py-2.5 border border-(--border) rounded-full text-sm bg-(--muted)/5 text-foreground placeholder-(--muted) hover:border-foreground focus:outline-none focus:border-(--accent) focus:ring-2 focus:ring-(--accent)/10 transition-all"
                            placeholder="Search users..."
                        />

                        {/* Mobile Search Results Dropdown */}
                        {showSearchResults && (
                            <div className="absolute top-full left-0 right-0 mt-2 bg-background border border-(--border) rounded-2xl shadow-xl overflow-hidden animate-in fade-in zoom-in-95 duration-200 max-h-96 overflow-y-auto z-200">
                                {searchResults.length > 0 ? (
                                    <div className="py-2">
                                        {searchResults.map((result) => (
                                            <button
                                                type="button"
                                                key={result.id}
                                                onMouseDown={(e) => handleResultClick(result.id, e)}
                                                className="w-full flex items-center gap-3 px-4 py-3 hover:bg-(--muted)/5 transition-colors cursor-pointer text-left"
                                            >
                                                <div className="w-10 h-10 rounded-full bg-(--muted)/10 flex items-center justify-center overflow-hidden shrink-0">
                                                    {result.avatar_url ? (
                                                        <img src={result.avatar_url} alt={result.username || "User"} className="w-full h-full object-cover" />
                                                    ) : (
                                                        <User className="w-5 h-5 text-(--muted)" />
                                                    )}
                                                </div>
                                                <div className="flex-1 min-w-0">
                                                    <p className="text-sm font-medium text-foreground truncate">
                                                        {result.username}
                                                    </p>
                                                </div>
                                            </button>
                                        ))}
                                    </div>
                                ) : (
                                    <div className="p-4 text-center text-sm text-(--muted)">
                                        No users found
                                    </div>
                                )}
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </nav>
    );
}