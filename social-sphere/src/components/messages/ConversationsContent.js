"use client";

import { useRouter, usePathname } from "next/navigation";
import { useStore } from "@/store/store";
import { User, MessageCircle, Loader2 } from "lucide-react";
import { ConversationsProvider, useConversations } from "@/context/ConversationsContext";

function ConversationsSidebar() {
    const router = useRouter();
    const pathname = usePathname();
    const user = useStore((state) => state.user);
    const decrementUnreadCount = useStore((state) => state.decrementUnreadCount);

    const {
        conversations,
        markAsRead,
        loadMore,
        isLoadingMore,
        hasMore,
    } = useConversations();

    // Extract selected ID from pathname (/messages/[id])
    const selectedId = pathname.startsWith("/messages/")
        ? pathname.split("/messages/")[1]
        : null;

    // Determine if we're on a conversation page (for mobile view)
    const isOnConversation = !!selectedId;

    // Format relative time
    const formatRelativeTime = (dateString) => {
        if (!dateString) return "";
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
    const truncateMessage = (text, maxLength = 40) => {
        if (!text) return "";
        return text.length > maxLength ? text.substring(0, maxLength) + "..." : text;
    };

    // Check if conversation has unread messages
    const hasUnreadMessages = (conv) => {
        return conv.UnreadCount > 0 && conv.LastMessage?.sender?.id !== user?.id;
    };

    // Handle conversation click
    const handleSelectConversation = (conv) => {
        const id = conv.Interlocutor?.id;

        // Mark as read ONLY if it has unread messages
        if (hasUnreadMessages(conv)) {
            markAsRead(id);
            decrementUnreadCount();
        }

        router.push(`/messages/${id}`);
    };

    return (
        <div
            className={`w-full md:w-80 lg:w-96 border-r border-(--border) flex flex-col ${
                isOnConversation ? "hidden md:flex" : "flex"
            }`}
        >
            {/* Header */}
            <div className="p-4 border-b border-(--border) flex items-center justify-between">
                <h1 className="text-xl font-bold text-foreground">Messages</h1>
            </div>

            {/* Conversations List */}
            <div className="flex-1 overflow-y-auto">
                {conversations.length > 0 ? (
                    <div>
                        {conversations.map((conv) => (
                            <button
                                key={conv.Interlocutor?.id}
                                onClick={() => handleSelectConversation(conv)}
                                className={`w-full flex items-start gap-3 px-4 py-3 transition-colors cursor-pointer text-left border-b border-(--border)/50 ${
                                    selectedId === conv.Interlocutor?.id
                                        ? "bg-(--accent)/10"
                                        : "hover:bg-(--muted)/5"
                                }`}
                            >
                                {/* Avatar */}
                                <div className="relative shrink-0">
                                    <div className="w-12 h-12 rounded-full bg-(--muted)/10 flex items-center justify-center overflow-hidden border border-(--border)">
                                        {conv.Interlocutor?.avatar_url ? (
                                            <img
                                                src={conv.Interlocutor.avatar_url}
                                                alt={conv.Interlocutor.username || "User"}
                                                className="w-full h-full object-cover"
                                            />
                                        ) : (
                                            <User className="w-6 h-6 text-(--muted)" />
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
                                        <p
                                            className={`text-sm truncate ${
                                                hasUnreadMessages(conv)
                                                    ? "font-semibold text-foreground"
                                                    : "font-medium text-foreground"
                                            }`}
                                        >
                                            {conv.Interlocutor?.username || "Unknown User"}
                                        </p>
                                        {conv?.UpdatedAt && (
                                            <span className="text-xs text-(--muted) shrink-0">
                                                {formatRelativeTime(conv.UpdatedAt)}
                                            </span>
                                        )}
                                    </div>
                                    <p
                                        className={`text-sm mt-0.5 truncate ${
                                            hasUnreadMessages(conv) ? "text-foreground" : "text-(--muted)"
                                        }`}
                                    >
                                        {conv.LastMessage?.sender?.id === user?.id ? "You: " : ""}
                                        {truncateMessage(conv.LastMessage?.message_text)}
                                    </p>
                                </div>
                            </button>
                        ))}

                        {/* Load More Button */}
                        {hasMore && (
                            <button
                                onClick={loadMore}
                                disabled={isLoadingMore}
                                className="w-full py-3 text-sm text-(--accent) hover:bg-(--muted)/5 transition-colors disabled:opacity-50"
                            >
                                {isLoadingMore ? (
                                    <Loader2 className="w-4 h-4 mx-auto animate-spin" />
                                ) : (
                                    "Load more"
                                )}
                            </button>
                        )}
                    </div>
                ) : (
                    <div className="flex flex-col items-center justify-center py-12 px-4">
                        <MessageCircle className="w-12 h-12 text-(--muted) mb-3 opacity-30" />
                        <p className="text-(--muted) text-center">No conversations yet</p>
                        <p className="text-(--muted) text-sm text-center mt-1">
                            Start chatting with someone!
                        </p>
                    </div>
                )}
            </div>
        </div>
    );
}

export default function ConversationsContent({ conversations = [], children }) {
    return (
        <ConversationsProvider initialConversations={conversations}>
            <div className="h-[calc(100vh-5rem)] flex bg-background">
                <ConversationsSidebar />
                {children}
            </div>
        </ConversationsProvider>
    );
}
