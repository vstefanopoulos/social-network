"use client";

import { useEffect, useState, useRef, useCallback } from "react";
import { useRouter } from "next/navigation";
import { getMessages } from "@/actions/chat/get-messages";
import { useStore } from "@/store/store";
import { User, Send, MessageCircle, Loader2, ChevronLeft, Smile } from "lucide-react";
import { motion } from "motion/react";
import { useLiveSocket } from "@/context/LiveSocketContext";
import { useConversations } from "@/context/ConversationsContext";
import EmojiPicker from "emoji-picker-react";
import { useMsgReceiver } from "@/store/store";
import Link from "next/link";

export default function MessagesContent({
    interlocutorId,
    initialMessages = [],
    firstMessage = false,
}) {
    const router = useRouter();
    const user = useStore((state) => state.user);
    const receiver = useMsgReceiver((state) => state.msgReceiver);
    const clearMsgReceiver = useMsgReceiver((state) => state.clearMsgReceiver);

    const { conversations, addConversation, updateConversation } = useConversations();

    const [messages, setMessages] = useState(initialMessages);
    const [isLoadingMoreMessages, setIsLoadingMoreMessages] = useState(false);
    const [hasMoreMessages, setHasMoreMessages] = useState(() => initialMessages.length >= 20);
    const [messageText, setMessageText] = useState("");
    const [showEmojiPicker, setShowEmojiPicker] = useState(false);

    const messagesEndRef = useRef(null);
    const messagesContainerRef = useRef(null);
    const isLoadingMoreRef = useRef(false);
    const emojiPickerRef = useRef(null);
    const interlocutorIdRef = useRef(interlocutorId);

    // Keep interlocutorId ref in sync for WebSocket callback
    useEffect(() => {
        interlocutorIdRef.current = interlocutorId;
    }, [interlocutorId]);

    // Find selected conversation from conversations list
    const selectedConv = conversations.find((conv) => conv.Interlocutor?.id === interlocutorId) || null;

    // Handle new message flow - add conversation to list if receiver exists
    useEffect(() => {
        if (firstMessage && receiver) {
            const newConv = {
                Interlocutor: {
                    id: receiver.id,
                    username: receiver.username,
                    avatar_url: receiver.avatar_url,
                },
                UnreadCount: 0,
            };
            addConversation(newConv);
            clearMsgReceiver();
        }
    }, [firstMessage, receiver, addConversation, clearMsgReceiver]);

    const hasMessageText = messageText.length > 0;

    // WebSocket connection from context
    const { isConnected, addOnPrivateMessage, removeOnPrivateMessage, sendPrivateMessage } = useLiveSocket();

    // Handle incoming private messages from WebSocket
    const handlePrivateMessage = useCallback((msg) => {
        const senderId = msg.sender?.id;
        const isOwnMessage = senderId === user?.id;
        const currentInterlocutorId = interlocutorIdRef.current;

        // Only add to messages if it's for the current conversation
        if (isOwnMessage) {
            // This is a confirmation of our sent message - replace pending with confirmed
            setMessages((prev) => {
                const pendingIndex = prev.findIndex(
                    (m) => m._pending && m.message_text === msg.message_text
                );
                scrollToBottom(true);

                if (pendingIndex !== -1) {
                    const updated = [...prev];
                    updated[pendingIndex] = { ...msg, _pending: false };
                    return updated;
                }

                if (prev.some((m) => m.id === msg.id)) return prev;
                return [...prev, msg];
            });
        } else if (senderId === currentInterlocutorId) {
            // Message from the other person in this conversation
            setMessages((prev) => {
                if (prev.some((m) => m.id === msg.id)) return prev;
                return [...prev, msg];
            });
            scrollToBottom(true);
        }
    }, [user?.id]);

    // Register message handler
    useEffect(() => {
        addOnPrivateMessage(handlePrivateMessage);
        return () => removeOnPrivateMessage(handlePrivateMessage);
    }, [addOnPrivateMessage, removeOnPrivateMessage, handlePrivateMessage]);

    // Scroll to bottom of messages
    const scrollToBottom = (instant = false) => {
        if (instant) {
            messagesContainerRef.current?.scrollTo(0, messagesContainerRef.current.scrollHeight);
        } else {
            messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
        }
    };

    // Scroll to bottom on first load
    useEffect(() => {
        if (messages.length > 0) {
            scrollToBottom(true);
        }
    }, [interlocutorId]);

    // Format message time
    const formatMessageTime = (dateString) => {
        const date = new Date(dateString);
        return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
    };

    // Load more messages (pagination)
    const loadMoreMessages = useCallback(async () => {
        if (isLoadingMoreMessages || !hasMoreMessages || messages.length === 0 || !interlocutorId) return;

        isLoadingMoreRef.current = true;
        setIsLoadingMoreMessages(true);

        const container = messagesContainerRef.current;
        const prevScrollHeight = container?.scrollHeight || 0;

        try {
            const oldestMsg = messages[0];
            const boundary = oldestMsg.id;

            const result = await getMessages({
                interlocutorId,
                boundary,
                limit: 10,
            });

            if (result.success && result.data) {
                const olderMsgs = result.data.Messages?.reverse() || [];
                setMessages((prev) => [...olderMsgs, ...prev]);
                setHasMoreMessages(olderMsgs.length >= 10);

                requestAnimationFrame(() => {
                    if (container) {
                        const newScrollHeight = container.scrollHeight;
                        container.scrollTop = newScrollHeight - prevScrollHeight;
                    }
                });
            }
        } catch (error) {
            console.error("Error loading more messages:", error);
        } finally {
            setIsLoadingMoreMessages(false);
            isLoadingMoreRef.current = false;
        }
    }, [isLoadingMoreMessages, hasMoreMessages, messages, interlocutorId]);

    // Handle scroll for infinite loading
    const handleMessagesScroll = (e) => {
        if (isLoadingMoreRef.current || !hasMoreMessages) return;

        const container = e.target;
        if (container.scrollTop < 100) {
            loadMoreMessages();
        }
    };

    // Handle send message - uses WebSocket
    const handleSendMessage = async (e) => {
        e.preventDefault();
        if (!messageText.trim() || !interlocutorId || !isConnected) return;

        const msgToSend = messageText.trim();
        setMessageText("");

        const tempId = `temp-${Date.now()}`;

        // Optimistically add message
        const optimisticMessage = {
            id: tempId,
            message_text: msgToSend,
            sender: { id: user?.id },
            created_at: new Date().toISOString(),
            _pending: true,
        };
        setMessages((prev) => [...prev, optimisticMessage]);

        // Update conversation in context
        updateConversation(interlocutorId, {
            LastMessage: {
                message_text: msgToSend,
                sender: { id: user?.id },
            },
            UpdatedAt: new Date().toISOString(),
        });

        try {
            await sendPrivateMessage(interlocutorId, msgToSend);
        } catch (error) {
            console.error("Error sending message:", error);
            setMessages((prev) => prev.filter((m) => m.id !== tempId));
            setMessageText(msgToSend);
        }
    };

    // Handle back button on mobile
    const handleBackToList = () => {
        router.push("/messages");
    };

    // Handle emoji selection
    const onEmojiClick = (emojiData) => {
        setMessageText((prev) => prev + emojiData.emoji);
        setShowEmojiPicker(false);
    };

    // Close emoji picker when clicking outside
    useEffect(() => {
        const handleClickOutside = (event) => {
            if (emojiPickerRef.current && !emojiPickerRef.current.contains(event.target)) {
                setShowEmojiPicker(false);
            }
        };

        if (showEmojiPicker) {
            document.addEventListener("mousedown", handleClickOutside);
        }

        return () => {
            document.removeEventListener("mousedown", handleClickOutside);
        };
    }, [showEmojiPicker]);

    // Get interlocutor info from selectedConv or receiver
    const interlocutor = selectedConv?.Interlocutor || receiver;

    return (
        <div className="flex-1 flex flex-col">
            {interlocutor ? (
                <>
                    {/* Chat Header */}
                    <div className="p-4 border-b border-(--border) flex items-center gap-3">
                        {/* Back button for mobile */}
                        <button
                            onClick={handleBackToList}
                            className="md:hidden p-2 -ml-2 rounded-full hover:bg-(--muted)/10 transition-colors"
                        >
                            <ChevronLeft className="w-5 h-5 text-(--muted)" />
                        </button>

                        <div className="w-10 h-10 rounded-full bg-(--muted)/10 flex items-center justify-center overflow-hidden border border-(--border)">
                            {interlocutor.avatar_url ? (
                                <img
                                    src={interlocutor.avatar_url}
                                    alt={interlocutor.username || "User"}
                                    className="w-full h-full object-cover"
                                />
                            ) : (
                                <User className="w-5 h-5 text-(--muted)" />
                            )}
                        </div>
                        <div>
                            <Link
                                href={`/profile/${interlocutor.id}`}
                                prefetch={false}
                                onClick={(e) => e.stopPropagation()}
                                className="font-semibold text-foreground hover:underline hover:text-(--accent)"
                            >
                                {interlocutor.username}
                            </Link>
                            {/* <p className="font-semibold text-foreground">
                                {interlocutor.username || "Unknown User"}
                            </p> */}
                        </div>
                    </div>

                    {/* Messages */}
                    <div
                        ref={messagesContainerRef}
                        onScroll={handleMessagesScroll}
                        className="flex-1 overflow-y-auto p-4 space-y-3"
                    >
                        {messages.length > 0 ? (
                            <>
                                {isLoadingMoreMessages && (
                                    <div className="flex justify-center py-2 mb-3">
                                        <Loader2 className="w-4 h-4 text-(--muted) animate-spin" />
                                    </div>
                                )}
                                {messages.map((msg, index) => {
                                    const isMe = msg.sender?.id === user?.id;
                                    const isPending = msg._pending;
                                    return (
                                        <motion.div
                                            key={msg.id || index}
                                            initial={{ opacity: 0, y: 10 }}
                                            animate={{ opacity: isPending ? 0.5 : 1, y: 0 }}
                                            transition={{ duration: 0.2 }}
                                            className={`flex ${isMe ? "justify-end" : "justify-start"}`}
                                        >
                                            <div
                                                className={`max-w-[75%] px-4 py-2.5 rounded-2xl ${
                                                    isMe
                                                        ? "bg-(--accent) text-white rounded-br-md"
                                                        : "bg-(--muted)/10 text-foreground rounded-bl-md"
                                                }`}
                                            >
                                                <p className="text-sm whitespace-pre-wrap wrap-break-word">
                                                    {msg.message_text}
                                                </p>
                                                <p
                                                    className={`text-[10px] mt-1 ${
                                                        isMe ? "text-white/70" : "text-(--muted)"
                                                    }`}
                                                >
                                                    {formatMessageTime(msg.created_at)}
                                                </p>
                                            </div>
                                        </motion.div>
                                    );
                                })}
                                <div ref={messagesEndRef} />
                            </>
                        ) : (
                            <div className="flex flex-col items-center justify-center py-12">
                                <MessageCircle className="w-12 h-12 text-(--muted) mb-3 opacity-30" />
                                <p className="text-(--muted)">No messages yet</p>
                                <p className="text-(--muted) text-sm">Say hello!</p>
                            </div>
                        )}
                    </div>

                    {/* Message Input */}
                    <form onSubmit={handleSendMessage} className="p-4 border-t border-(--border)">
                        <div className="flex items-center gap-3">
                            {/* Emoji Picker */}
                            <div className="relative" ref={emojiPickerRef}>
                                <button
                                    type="button"
                                    onClick={() => setShowEmojiPicker(!showEmojiPicker)}
                                    className="p-3 text-(--muted) hover:text-foreground hover:bg-(--muted)/10 rounded-full transition-all cursor-pointer"
                                >
                                    <Smile className="w-5 h-5" />
                                </button>
                                {showEmojiPicker && (
                                    <div className="absolute bottom-14 left-0 z-50">
                                        <EmojiPicker
                                            onEmojiClick={onEmojiClick}
                                            width={320}
                                            height={400}
                                            previewConfig={{ showPreview: false }}
                                        />
                                    </div>
                                )}
                            </div>
                            <input
                                type="text"
                                value={messageText}
                                onChange={(e) => setMessageText(e.target.value)}
                                placeholder="Type a message..."
                                className="flex-1 px-4 py-3 border border-(--border) rounded-full text-sm bg-(--muted)/5 text-foreground placeholder-(--muted) hover:border-foreground focus:outline-none focus:border-(--accent) focus:ring-2 focus:ring-(--accent)/10 transition-all"
                            />
                            {hasMessageText && (
                                <button
                                    type="submit"
                                    disabled={!messageText.trim() || !isConnected}
                                    className="p-3 bg-(--accent) text-white rounded-full hover:bg-(--accent-hover) transition-all disabled:opacity-50"
                                >
                                    <Send className="w-5 h-5" />
                                </button>
                            )}
                        </div>
                    </form>
                </>
            ) : (
                /* No interlocutor found - shouldn't happen normally */
                <div className="flex-1 flex flex-col items-center justify-center px-4">
                    <div className="w-20 h-20 rounded-full bg-(--muted)/10 flex items-center justify-center mb-4">
                        <MessageCircle className="w-10 h-10 text-(--muted)" />
                    </div>
                    <h2 className="text-xl font-semibold text-foreground mb-2">Conversation not found</h2>
                    <p className="text-(--muted) text-center max-w-sm">
                        This conversation doesn't exist or couldn't be loaded.
                    </p>
                </div>
            )}
        </div>
    );
}
