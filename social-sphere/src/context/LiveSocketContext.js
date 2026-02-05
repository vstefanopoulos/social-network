"use client";

import { createContext, useContext, useEffect, useRef, useState, useCallback } from "react";
import { useStore } from "@/store/store";

const LiveSocketContext = createContext(null);

export const ConnectionState = {
    CONNECTING: "connecting",
    CONNECTED: "connected",
    DISCONNECTED: "disconnected",
    RECONNECTING: "reconnecting",
};

export function LiveSocketProvider({ children, wsUrl }) {
    const user = useStore((state) => state.user);
    const wsRef = useRef(null);
    const reconnectTimeoutRef = useRef(null);
    const reconnectAttemptsRef = useRef(0);
    const subscribedGroupsRef = useRef(new Set());
    const hadUserRef = useRef(false); // Track if we ever had a user (to detect logout vs initial state)

    const [connectionState, setConnectionState] = useState(ConnectionState.DISCONNECTED);
    const [unreadMessages, setUnreadMessages] = useState([]);
    const [unreadNotifications, setUnreadNotifications] = useState([]);

    // Sets of callback listeners (supports multiple consumers)
    const privateMessageListenersRef = useRef(new Set());
    const groupMessageListenersRef = useRef(new Set());
    const notificationListenersRef = useRef(new Set());

    const connect = useCallback(() => {
        // Don't connect if no user or already connected
        if (!user) {
            return;
        }

        if (wsRef.current?.readyState === WebSocket.OPEN) {
            console.log("[LiveSocket] Already connected");
            return;
        }

        // Clean up existing connection
        if (wsRef.current) {
            wsRef.current.close();
        }

        setConnectionState(
            reconnectAttemptsRef.current > 0
                ? ConnectionState.RECONNECTING
                : ConnectionState.CONNECTING
        );

        try {
            const ws = new WebSocket(`${wsUrl}/live`);
            wsRef.current = ws;

            ws.onopen = () => {
                console.log("[LiveSocket] Connected");
                setConnectionState(ConnectionState.CONNECTED);
                reconnectAttemptsRef.current = 0;

                // Re-subscribe to any groups we were subscribed to (for reconnections)
                if (subscribedGroupsRef.current.size > 0) {
                    console.log("[LiveSocket] Re-subscribing to groups:", [...subscribedGroupsRef.current]);
                    subscribedGroupsRef.current.forEach((groupId) => {
                        ws.send(`sub:${groupId}`);
                    });
                }
            };

            ws.onmessage = (event) => {
                try {
                    // Skip empty messages
                    if (!event.data || event.data.trim() === "") {
                        return;
                    }

                    const data = JSON.parse(event.data);

                    // Handle both array (batched from NATS) and single object (direct response)
                    const messages = Array.isArray(data) ? data : [data];

                    for (const msg of messages) {
                        if (msg.group_id || msg.GroupId) {
                            // Group message (handle both snake_case and PascalCase)
                            console.log("[LiveSocket] Group message received:", msg);
                            groupMessageListenersRef.current.forEach((listener) => listener(msg));
                        } else if (msg.conversation_id || msg.ConversationId) {
                            // Private message (handle both snake_case and PascalCase)
                            console.log("[LiveSocket] Private message received:", msg);

                            // Add to unread if not from current user
                            const senderId = msg.sender?.id || msg.Sender?.id;
                            if (senderId !== user?.id) {
                                setUnreadMessages((prev) => {
                                    if (prev.some((m) => m.id === msg.id)) return prev;
                                    return [...prev, msg];
                                });
                            }

                            privateMessageListenersRef.current.forEach((listener) => listener(msg));
                        } else if (msg.notification_type || msg.type) {
                            // Notification
                            console.log("[LiveSocket] Notification received:", msg);
                            setUnreadNotifications((prev) => [...prev, msg]);
                            notificationListenersRef.current.forEach((listener) => listener(msg));
                        } else {
                            console.log("[LiveSocket] Unknown message type:", msg);
                        }
                    }
                } catch (err) {
                    console.error("[LiveSocket] Failed to parse message:", err);
                }
            };

            ws.onerror = (error) => {
                console.error("[LiveSocket] Error:", error);
            };

            ws.onclose = (event) => {
                console.log("[LiveSocket] Disconnected:", event.code, event.reason);
                setConnectionState(ConnectionState.DISCONNECTED);

                // Reconnect if not a clean close and user is still logged in
                if (event.code !== 1000 && user) {
                    const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current), 30000);
                    reconnectAttemptsRef.current++;

                    console.log(`[LiveSocket] Reconnecting in ${delay}ms (attempt ${reconnectAttemptsRef.current})`);

                    reconnectTimeoutRef.current = setTimeout(() => {
                        connect();
                    }, delay);
                }
            };
        } catch (err) {
            console.error("[LiveSocket] Failed to create connection:", err);
            setConnectionState(ConnectionState.DISCONNECTED);
        }
    }, [user]);

    const disconnect = useCallback((clearSubscriptions = true) => {
        console.log("[LiveSocket] Disconnecting...");

        if (reconnectTimeoutRef.current) {
            clearTimeout(reconnectTimeoutRef.current);
            reconnectTimeoutRef.current = null;
        }

        reconnectAttemptsRef.current = 0;

        // Only clear subscriptions on actual logout, not on React Strict Mode remounts
        if (clearSubscriptions) {
            subscribedGroupsRef.current.clear();
        }

        if (wsRef.current) {
            wsRef.current.close(1000, "User logout");
            wsRef.current = null;
        }

        setConnectionState(ConnectionState.DISCONNECTED);
        setUnreadMessages([]);
        setUnreadNotifications([]);
    }, []);

    const subscribeToGroup = useCallback((groupId, isMember) => {
        // Only subscribe if user is a group member
        if (!isMember) {
            return;
        }

        // Don't re-subscribe if already subscribed
        if (subscribedGroupsRef.current.has(groupId)) {
            return;
        }
        subscribedGroupsRef.current.add(groupId);

        // Send immediately (caller should ensure WS is connected)
        if (wsRef.current?.readyState === WebSocket.OPEN) {
            wsRef.current.send(`sub:${groupId}`);
            console.log("[LiveSocket] Subscribed to group:", groupId);
        }
    }, []);

    const unsubscribeFromGroup = useCallback((groupId) => {
        // Don't unsubscribe if not subscribed
        if (!subscribedGroupsRef.current.has(groupId)) {
            return;
        }
        subscribedGroupsRef.current.delete(groupId);
        if (wsRef.current?.readyState === WebSocket.OPEN) {
            wsRef.current.send(`unsub:${groupId}`);
            console.log("[LiveSocket] Unsubscribed from group:", groupId);
        }
    }, []);

    const clearUnreadNotification = useCallback((notificationId) => {
        setUnreadNotifications((prev) => prev.filter((n) => n.id !== notificationId));
    }, []);

    const clearAllUnreadNotifications = useCallback(() => {
        setUnreadNotifications([]);
    }, []);

    // Add/remove callback listeners (supports multiple consumers)
    const addOnPrivateMessage = useCallback((callback) => {
        privateMessageListenersRef.current.add(callback);
    }, []);

    const removeOnPrivateMessage = useCallback((callback) => {
        privateMessageListenersRef.current.delete(callback);
    }, []);

    const addOnGroupMessage = useCallback((callback) => {
        groupMessageListenersRef.current.add(callback);
    }, []);

    const removeOnGroupMessage = useCallback((callback) => {
        groupMessageListenersRef.current.delete(callback);
    }, []);

    const addOnNotification = useCallback((callback) => {
        notificationListenersRef.current.add(callback);
    }, []);

    const removeOnNotification = useCallback((callback) => {
        notificationListenersRef.current.delete(callback);
    }, []);

    // Send a private message through WebSocket
    // Returns a Promise that resolves with the server response or rejects on error/timeout
    const sendPrivateMessage = useCallback((interlocutorId, messageText) => {
        return new Promise((resolve, reject) => {
            if (wsRef.current?.readyState !== WebSocket.OPEN) {
                reject(new Error("WebSocket not connected"));
                return;
            }

            const payload = JSON.stringify({
                interlocutor_id: interlocutorId,
                message_text: messageText,
            });

            wsRef.current.send(`private_chat:${payload}`);
            console.log("[LiveSocket] Sent private message");

            // The server will send the created message back directly to the sender
            // We'll resolve immediately since the message will come through onmessage
            resolve({ sent: true });
        });
    }, []);

    // Send a group message through WebSocket
    const sendGroupMessage = useCallback((groupId, messageText) => {
        return new Promise((resolve, reject) => {
            if (wsRef.current?.readyState !== WebSocket.OPEN) {
                reject(new Error("WebSocket not connected"));
                return;
            }

            const payload = JSON.stringify({
                group_id: groupId,
                message_text: messageText,
            });

            wsRef.current.send(`group_chat:${payload}`);
            console.log("[LiveSocket] Sent group message");

            // The server will send the created message back directly to the sender
            resolve({ sent: true });
        });
    }, []);

    // Connect when user logs in, disconnect when user logs out
    useEffect(() => {
        if (user) {
            hadUserRef.current = true;
            connect();
        } else if (hadUserRef.current) {
            // Only clear subscriptions on actual logout (user was truthy before)
            // Not on initial mount when user is still loading
            disconnect(true);
        }

        return () => {
            // Don't clear subscriptions on cleanup (React Strict Mode remounts)
            // Just close the connection
            disconnect(false);
        };
    }, [user, connect, disconnect]);

    const value = {
        connectionState,
        isConnected: connectionState === ConnectionState.CONNECTED,
        unreadMessages,
        unreadNotifications,
        unreadMessageCount: unreadMessages.length,
        unreadNotificationCount: unreadNotifications.length,
        subscribeToGroup,
        unsubscribeFromGroup,
        clearUnreadNotification,
        clearAllUnreadNotifications,
        addOnPrivateMessage,
        removeOnPrivateMessage,
        addOnGroupMessage,
        removeOnGroupMessage,
        addOnNotification,
        removeOnNotification,
        sendPrivateMessage,
        sendGroupMessage,
        disconnect,
    };

    return (
        <LiveSocketContext.Provider value={value}>
            {children}
        </LiveSocketContext.Provider>
    );
}

export function useLiveSocket() {
    const context = useContext(LiveSocketContext);
    if (!context) {
        throw new Error("useLiveSocket must be used within a LiveSocketProvider");
    }
    return context;
}
