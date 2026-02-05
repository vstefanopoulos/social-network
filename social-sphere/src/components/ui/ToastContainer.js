"use client";

import { AnimatePresence } from "motion/react";
import Toast from "./Toast";

export default function ToastContainer({ toasts, onDismiss, onPause, onResume, onClick }) {
    return (
        <div className="fixed bottom-6 right-6 z-50 pointer-events-none flex flex-col gap-3">
            <AnimatePresence mode="popLayout">
                {toasts.map((toast) => (
                    <Toast
                        key={toast.id}
                        notification={toast.notification}
                        onDismiss={() => onDismiss(toast.id)}
                        onMouseEnter={() => onPause(toast.id)}
                        onMouseLeave={() => onResume(toast.id)}
                    />
                ))}
            </AnimatePresence>
        </div>
    );
}
