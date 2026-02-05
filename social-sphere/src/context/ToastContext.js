"use client";

import { createContext, useContext, useState, useCallback, useRef, useEffect } from "react";
import { useLiveSocket } from "./LiveSocketContext";
import ToastContainer from "@/components/ui/ToastContainer";

const ToastContext = createContext(null);

const MAX_VISIBLE_TOASTS = 3;
const AUTO_DISMISS_MS = 4000;

export function ToastProvider({ children }) {
    const [toasts, setToasts] = useState([]);
    const [queue, setQueue] = useState([]);
    const { addOnNotification, removeOnNotification } = useLiveSocket();
    const toastIdRef = useRef(0);
    const timersRef = useRef(new Map());
    const remainingTimeRef = useRef(new Map());
    const pausedAtRef = useRef(new Map());
    const dismissToastRef = useRef(null);

    const startTimer = useCallback((id, duration = AUTO_DISMISS_MS) => {
        const timer = setTimeout(() => {
            dismissToastRef.current?.(id);
        }, duration);
        timersRef.current.set(id, timer);
        remainingTimeRef.current.set(id, duration);
        pausedAtRef.current.set(id, Date.now());
    }, []);

    const dismissToast = useCallback((id) => {
        // Clear the auto-dismiss timer
        if (timersRef.current.has(id)) {
            clearTimeout(timersRef.current.get(id));
            timersRef.current.delete(id);
        }

        setToasts((prev) => prev.filter((t) => t.id !== id));

        // Move next queued toast to visible
        setQueue((prevQueue) => {
            if (prevQueue.length === 0) return prevQueue;

            const [next, ...rest] = prevQueue;
            setToasts((prevToasts) => {
                if (prevToasts.length < MAX_VISIBLE_TOASTS) {
                    startTimer(next.id);
                    return [...prevToasts, next];
                }
                return prevToasts;
            });
            return rest;
        });
    }, [startTimer]);

    // Keep the ref updated
    dismissToastRef.current = dismissToast;

    const pauseToast = useCallback((id) => {
        if (timersRef.current.has(id)) {
            clearTimeout(timersRef.current.get(id));
            timersRef.current.delete(id);

            const pausedAt = pausedAtRef.current.get(id) || Date.now();
            const originalRemaining = remainingTimeRef.current.get(id) || AUTO_DISMISS_MS;
            const elapsed = Date.now() - pausedAt;
            const remaining = Math.max(originalRemaining - elapsed, 1000);
            remainingTimeRef.current.set(id, remaining);
        }
    }, []);

    const resumeToast = useCallback((id) => {
        const remaining = remainingTimeRef.current.get(id);
        if (remaining && !timersRef.current.has(id)) {
            pausedAtRef.current.set(id, Date.now());
            const timer = setTimeout(() => {
                dismissToastRef.current?.(id);
            }, remaining);
            timersRef.current.set(id, timer);
        }
    }, []);

    const showToast = useCallback((notification) => {
        const id = ++toastIdRef.current;
        const toast = { id, notification, createdAt: Date.now() };

        setToasts((prev) => {
            if (prev.length < MAX_VISIBLE_TOASTS) {
                startTimer(id);
                return [...prev, toast];
            } else {
                setQueue((prevQueue) => [...prevQueue, toast]);
                return prev;
            }
        });
    }, [startTimer]);

    // Listen for notifications from LiveSocket
    useEffect(() => {
        const handleNotification = (notification) => {
            showToast(notification);
        };

        addOnNotification(handleNotification);
        return () => {
            removeOnNotification(handleNotification);
        };
    }, [addOnNotification, removeOnNotification, showToast]);

    // Cleanup timers on unmount
    useEffect(() => {
        return () => {
            timersRef.current.forEach((timer) => clearTimeout(timer));
            timersRef.current.clear();
        };
    }, []);

    const value = {
        toasts,
        showToast,
        dismissToast,
        pauseToast,
        resumeToast,
    };

    return (
        <ToastContext.Provider value={value}>
            {children}
            <ToastContainer
                toasts={toasts}
                onDismiss={dismissToast}
                onPause={pauseToast}
                onResume={resumeToast}
            />
        </ToastContext.Provider>
    );
}

export function useToast() {
    const context = useContext(ToastContext);
    if (!context) {
        throw new Error("useToast must be used within a ToastProvider");
    }
    return context;
}
