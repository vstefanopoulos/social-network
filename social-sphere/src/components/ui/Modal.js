"use client";

import { useEffect } from "react";
import { X } from "lucide-react";

/**
 * Universal Modal Component
 *
 * @param {boolean} isOpen - Controls modal visibility
 * @param {function} onClose - Callback when modal is closed
 * @param {string} title - Modal header title
 * @param {string} description - Optional description text
 * @param {ReactNode} children - Custom content
 * @param {ReactNode} footer - Custom footer content (buttons, etc.)
 * @param {function} onConfirm - Optional: Primary action callback
 * @param {string} confirmText - Optional: Primary button text (default: "Confirm")
 * @param {string} cancelText - Optional: Cancel button text (default: "Cancel")
 * @param {boolean} isLoading - Optional: Loading state for primary button
 * @param {string} loadingText - Optional: Text to show when loading
 * @param {boolean} showCloseButton - Optional: Show X button in header (default: true)
 */
export default function Modal({
    isOpen,
    onClose,
    title,
    description,
    children,
    footer,
    onConfirm,
    confirmText = "Confirm",
    cancelText = "Cancel",
    isLoading = false,
    loadingText = "Loading...",
    showCloseButton = true
}) {
    useEffect(() => {
        if (isOpen) {
            document.body.style.overflow = "hidden";
        } else {
            document.body.style.overflow = "unset";
        }
        return () => {
            document.body.style.overflow = "unset";
        };
    }, [isOpen]);

    if (!isOpen) return null;

    // Default footer if onConfirm is provided but no custom footer
    const defaultFooter = onConfirm && !footer ? (
        <>
            <button
                type="button"
                onClick={onClose}
                disabled={isLoading}
                className="rounded-full bg-(--muted)/5 hover:bg-(--muted)/12 px-3 py-2 text-sm font-medium transition-colors cursor-pointer"
            >
                {cancelText}
            </button>
            <button
                type="button"
                onClick={onConfirm}
                disabled={isLoading}
                className="px-3 py-1.5 font-medium text-sm transition-colors cursor-pointer btn-primary flex items-center gap-2"
            >
                {isLoading ? (
                    <>
                        <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                        {loadingText}
                    </>
                ) : (
                    confirmText
                )}
            </button>
        </>
    ) : footer;

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center animate-in fade-in duration-200">
            {/* Backdrop */}
            <div
                className="absolute inset-0 bg-black/40 backdrop-blur-sm"
                onClick={isLoading ? undefined : onClose}
            />

            {/* Modal */}
            <div className="relative bg-background border border-(--border) rounded-2xl shadow-2xl max-w-md w-full mx-4 animate-in zoom-in-95 duration-200">
                {/* Header */}
                <div className="flex items-center justify-between p-2 pl-7 border-b border-(--border)">
                    <h3 className="text-lg font-semibold text-foreground">{title}</h3>
                    {showCloseButton && (
                        <button
                            onClick={onClose}
                            disabled={isLoading}
                            className="p-1 text-(--muted) hover:text-foreground hover:bg-(--muted)/10 rounded-full transition-colors disabled:opacity-50 cursor-pointer"
                        >
                            <X className="w-5 h-5" />
                        </button>
                    )}
                </div>

                {/* Content */}
                <div className="p-5">
                    {description && (
                        <p className="text-sm text-(--foreground)/80 mb-4">
                            {description}
                        </p>
                    )}
                    {children}
                </div>

                {/* Footer */}
                {defaultFooter && (
                    <div className="flex items-center justify-end gap-3 p-2 border-t border-(--border)">
                        {defaultFooter}
                    </div>
                )}
            </div>
        </div>
    );
}
