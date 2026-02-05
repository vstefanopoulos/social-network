"use client";

import { createContext, useContext, useState, useCallback, useRef, useEffect } from "react";
import { useStore } from "@/store/store";
import { getConv } from "@/actions/chat/get-conv";
import { getConvByID } from "@/actions/chat/get-conv-by-id";
import { useLiveSocket } from "@/context/LiveSocketContext";

const ConversationsContext = createContext(null);

export function ConversationsProvider({ initialConversations = [], children }) {
    const user = useStore((state) => state.user);
    const incrementUnreadCount = useStore((state) => state.incrementUnreadCount);
    const [conversations, setConversations] = useState(initialConversations);
    const [isLoadingMore, setIsLoadingMore] = useState(false);
    const [hasMore, setHasMore] = useState(() => initialConversations.length >= 15);

    // Ref to track current conversations for WebSocket callback
    const conversationsRef = useRef(conversations);
    useEffect(() => {
        conversationsRef.current = conversations;
    }, [conversations]);

    const { addOnPrivateMessage, removeOnPrivateMessage } = useLiveSocket();

    // Add a new conversation (for new message flow)
    const addConversation = useCallback((newConv) => {
        setConversations((prev) => {
            // Check if already exists
            const exists = prev.some((c) => c.Interlocutor?.id === newConv.Interlocutor?.id);
            if (exists) return prev;
            return [newConv, ...prev];
        });
    }, []);

    // Update a conversation (e.g., after sending message)
    const updateConversation = useCallback((interlocutorId, updates) => {
        setConversations((prev) =>
            prev.map((conv) =>
                conv.Interlocutor?.id === interlocutorId
                    ? { ...conv, ...updates }
                    : conv
            ).sort((a, b) => new Date(b.UpdatedAt || 0) - new Date(a.UpdatedAt || 0))
        );
    }, []);

    // Mark conversation as read (optimistic)
    const markAsRead = useCallback((interlocutorId) => {
        setConversations((prev) =>
            prev.map((conv) =>
                conv.Interlocutor?.id === interlocutorId
                    ? { ...conv, UnreadCount: 0 }
                    : conv
            )
        );
    }, []);

    // Load more conversations (pagination)
    const loadMore = useCallback(async () => {
        if (isLoadingMore || !hasMore || conversations.length === 0) return;

        setIsLoadingMore(true);
        try {
            const oldestConv = conversations[conversations.length - 1];
            const beforeDate = oldestConv.UpdatedAt;

            const result = await getConv({ first: false, beforeDate, limit: 15 });
            if (result.success && result.data) {
                setConversations((prev) => [...prev, ...result.data]);
                setHasMore(result.data.length >= 15);
            }
        } catch (error) {
            console.error("Error loading more conversations:", error);
        } finally {
            setIsLoadingMore(false);
        }
    }, [isLoadingMore, hasMore, conversations]);

    // Handle incoming WebSocket messages to update conversation list
    const handlePrivateMessage = useCallback(async (msg) => {
        const senderId = msg.sender?.id;
        const isOwnMessage = senderId === user?.id;

        // Only update conversation list for incoming messages (not own messages)
        // Own messages are handled by MessagesContent which calls updateConversation
        if (isOwnMessage) return;

        const currentConvs = conversationsRef.current;
        const existingConv = currentConvs.find((conv) => conv.Interlocutor?.id === senderId);

        if (existingConv) {
            // Update existing conversation
            setConversations((prev) =>
                prev.map((conv) => {
                    if (conv.Interlocutor?.id === senderId) {
                        return {
                            ...conv,
                            LastMessage: {
                                id: msg.id,
                                message_text: msg.message_text,
                                sender: msg.sender,
                            },
                            UpdatedAt: msg.created_at,
                            UnreadCount: (conv.UnreadCount || 0) + 1,
                        };
                    }
                    return conv;
                }).sort((a, b) => new Date(b.UpdatedAt) - new Date(a.UpdatedAt))
            );
            // Increment global unread count
            incrementUnreadCount();
        } else {
            // New conversation - fetch full data from server
            const result = await getConvByID({
                interlocutorId: senderId,
                convId: msg.conversation_id,
            });

            if (result.success && result.data) {
                setConversations((prev) => {
                    const alreadyExists = prev.some((conv) => conv.Interlocutor?.id === senderId);
                    if (alreadyExists) return prev;
                    return [{ ...result.data, UnreadCount: 1 }, ...prev];
                });
                // Increment global unread count
                incrementUnreadCount();
            }
        }
    }, [user?.id, incrementUnreadCount]);

    // Register WebSocket handler
    useEffect(() => {
        addOnPrivateMessage(handlePrivateMessage);
        return () => removeOnPrivateMessage(handlePrivateMessage);
    }, [addOnPrivateMessage, removeOnPrivateMessage, handlePrivateMessage]);

    const value = {
        conversations,
        setConversations,
        addConversation,
        updateConversation,
        markAsRead,
        loadMore,
        isLoadingMore,
        hasMore,
    };

    return (
        <ConversationsContext.Provider value={value}>
            {children}
        </ConversationsContext.Provider>
    );
}

export function useConversations() {
    const context = useContext(ConversationsContext);
    if (!context) {
        throw new Error("useConversations must be used within a ConversationsProvider");
    }
    return context;
}
