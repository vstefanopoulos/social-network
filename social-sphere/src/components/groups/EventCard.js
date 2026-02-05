"use client";

import { useState } from "react";
import { Calendar, Users, Check, X, Pencil, Trash2 } from "lucide-react";
import PostImage from "@/components/ui/PostImage";
import Modal from "@/components/ui/Modal";
import Tooltip from "@/components/ui/Tooltip";
import { useStore } from "@/store/store";
import { respondToEvent } from "@/actions/events/respond-to-event";
import { removeEventResponse } from "@/actions/events/remove-event-response";
import { deleteEvent } from "@/actions/events/delete-event";

export default function EventCard({ event, onDelete, onEdit }) {
    const user = useStore((state) => state.user);
    const [goingCount, setGoingCount] = useState(event.going_count || 0);
    const [notGoingCount, setNotGoingCount] = useState(event.not_going_count || 0);
    const [userResponse, setUserResponse] = useState(event.user_response); // true = going, false = not going, null = no response
    const [isResponding, setIsResponding] = useState(false);
    const [showDeleteModal, setShowDeleteModal] = useState(false);
    const [isDeleting, setIsDeleting] = useState(false);
    const [error, setError] = useState("");

    const isCreator = user && event.user && event.user.id === user.id;

    // Format event date nicely
    const formatEventDate = (dateString) => {
        const date = new Date(dateString);
        const options = {
            weekday: 'short',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        };
        return date.toLocaleDateString('en-US', options);
    };

    // Check if event is in the past
    const isPastEvent = new Date(event.event_date) < new Date();

    const handleResponse = async (going) => {
        if (isResponding) return;

        // If clicking the same response, remove it
        if (userResponse === going) {
            await handleRemoveResponse();
            return;
        }

        setIsResponding(true);
        setError("");

        // Optimistic update
        const previousResponse = userResponse;
        const previousGoingCount = goingCount;
        const previousNotGoingCount = notGoingCount;

        // Update counts based on previous and new response
        if (previousResponse === true) {
            setGoingCount(prev => prev - 1);
        } else if (previousResponse === false) {
            setNotGoingCount(prev => prev - 1);
        }

        if (going) {
            setGoingCount(prev => prev + 1);
        } else {
            setNotGoingCount(prev => prev + 1);
        }
        setUserResponse(going);

        try {
            const resp = await respondToEvent({
                id: event.event_id,
                going: going
            });

            if (!resp.success) {
                // Revert on error
                setUserResponse(previousResponse);
                setGoingCount(previousGoingCount);
                setNotGoingCount(previousNotGoingCount);
                setError(resp.error || "Failed to respond");
            }
        } catch (err) {
            // Revert on error
            setUserResponse(previousResponse);
            setGoingCount(previousGoingCount);
            setNotGoingCount(previousNotGoingCount);
            setError("Failed to respond. Please try again.");
        } finally {
            setIsResponding(false);
        }
    };

    const handleRemoveResponse = async () => {
        if (isResponding || userResponse === null) return;

        setIsResponding(true);
        setError("");

        const previousResponse = userResponse;
        const previousGoingCount = goingCount;
        const previousNotGoingCount = notGoingCount;

        // Optimistic update
        if (previousResponse === true) {
            setGoingCount(prev => prev - 1);
        } else if (previousResponse === false) {
            setNotGoingCount(prev => prev - 1);
        }
        setUserResponse(null);

        try {
            const resp = await removeEventResponse({
                id: event.event_id
            });

            if (!resp.success) {
                // Revert on error
                setUserResponse(previousResponse);
                setGoingCount(previousGoingCount);
                setNotGoingCount(previousNotGoingCount);
                setError(resp.error || "Failed to remove response");
            }
        } catch (err) {
            // Revert on error
            setUserResponse(previousResponse);
            setGoingCount(previousGoingCount);
            setNotGoingCount(previousNotGoingCount);
            setError("Failed to remove response. Please try again.");
        } finally {
            setIsResponding(false);
        }
    };

    const handleDeleteClick = (e) => {
        e.preventDefault();
        e.stopPropagation();
        setShowDeleteModal(true);
    };

    const handleDeleteConfirm = async () => {
        setIsDeleting(true);
        setError("");

        try {
            const resp = await deleteEvent(event.event_id);

            if (!resp.success) {
                setError(resp.error || "Failed to delete event");
                setIsDeleting(false);
                setShowDeleteModal(false);
                return;
            }

            setShowDeleteModal(false);
            setIsDeleting(false);

            if (onDelete) {
                onDelete(event.event_id);
            }
        } catch (err) {
            console.error("Failed to delete event:", err);
            setError("Failed to delete event. Please try again.");
            setIsDeleting(false);
            setShowDeleteModal(false);
        }
    };

    return (
        <div className="group bg-background border border-(--border) rounded-2xl transition-all hover:border-(--muted)/40 hover:shadow-sm overflow-hidden">
            {/* Event Image */}
            {event.image_url && (
                <PostImage src={event.image_url} alt={"image"} />
            )}

            {/* Content */}
            <div className="p-5">
                {/* Date Badge + Creator Actions */}
                <div className="flex items-center justify-between">
                    {/* Date Badge */}
                    <div
                        className={`inline-flex items-center gap-2 px-3 py-1.5 rounded-full text-xs font-medium mb-3 ${isPastEvent
                            ? "bg-(--muted)/10 text-(--muted)"
                            : "bg-(--accent)/10 text-(--accent)"
                            }`}
                    >
                        <Calendar className="w-3.5 h-3.5" />
                        <span>{formatEventDate(event.event_date)}</span>
                        {isPastEvent && <span className="ml-1">(Past)</span>}
                    </div>

                    {/* Creator Actions */}
                    {isCreator && (
                        <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 pr-3 transition-opacity">
                            <Tooltip content="Edit Event">
                                <button
                                    onClick={() => onEdit && onEdit(event)}
                                    className="p-2 text-(--muted) hover:text-(--accent) hover:bg-(--accent)/5 rounded-full transition-colors cursor-pointer"
                                >
                                    <Pencil className="w-4 h-4" />
                                </button>
                            </Tooltip>

                            <Tooltip content="Delete Event">
                                <button
                                    onClick={handleDeleteClick}
                                    className="p-2 text-(--muted) hover:text-red-500 hover:bg-red-500/5 rounded-full transition-colors cursor-pointer"
                                >
                                    <Trash2 className="w-4 h-4" />
                                </button>
                            </Tooltip>
                        </div>
                    )}
                </div>



                {/* Title and RSVP */}
                <div className="flex items-center justify-between">
                    {/* Title */}
                    <h3 className="text-lg font-semibold text-foreground">
                        {event.event_title}
                    </h3>
                </div>


                {/* Body */}
                <p className="text-sm text-(--foreground)/80 leading-relaxed whitespace-pre-wrap mb-4">
                    {event.event_body}
                </p>



                {/* Error Message */}
                {error && (
                    <div className="text-red-500 text-sm bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg px-4 py-2.5 mb-4">
                        {error}
                    </div>
                )}

                {/* Response Buttons */}
                {!isPastEvent && (
                    <div className="flex items-center justify-between gap-4">
                        {/* Buttons (left) */}
                        <div className="flex items-center gap-3">
                            <button
                                onClick={() => handleResponse(true)}
                                disabled={isResponding}
                                className={`flex items-center gap-2 px-4 py-2 rounded-full text-sm font-medium transition-all cursor-pointer disabled:opacity-50 ${userResponse === true
                                        ? "bg-green-600 text-white"
                                        : "bg-(--muted)/5 border border-(--border) text-(--muted) hover:text-green-600 hover:border-green-600"
                                    }`}
                            >
                                <Check className="w-4 h-4" />
                                <span>{userResponse === true ? "You said YES" : "Going"}</span>
                            </button>

                            <button
                                onClick={() => handleResponse(false)}
                                disabled={isResponding}
                                className={`flex items-center gap-2 px-4 py-2 rounded-full text-sm font-medium transition-all cursor-pointer disabled:opacity-50 ${userResponse === false
                                        ? "bg-red-500 text-white"
                                        : "bg-(--muted)/5 border border-(--border) text-(--muted) hover:text-red-500 hover:border-red-500"
                                    }`}
                            >
                                <X className="w-4 h-4" />
                                <span>{userResponse === false ? "You said NO" : "Not Going"}</span>
                            </button>
                        </div>

                        {/* RSVP Stats (right) */}
                        <div className="flex items-center gap-2">
                            <div className="flex items-center gap-2 text-sm">
                                <div className="flex items-center gap-1 text-green-600">
                                    <span className="font-medium">{goingCount}</span>
                                </div>
                                <span className="text-(--muted)">Accepted</span>
                            </div>

                            <span>|</span>

                            <div className="flex items-center gap-2 text-sm">
                                <div className="flex items-center gap-1 text-red-500">
                                    <span className="font-medium">{notGoingCount}</span>
                                </div>
                                <span className="text-(--muted)">Declined</span>
                            </div>
                        </div>
                    </div>
)}

                {/* Past Event Response Display */}
                {isPastEvent && userResponse !== null && (
                    <div className={`text-center py-2 px-4 rounded-xl text-sm ${userResponse ? "bg-green-600/10 text-green-600" : "bg-red-500/10 text-red-500"
                        }`}>
                        You responded: {userResponse ? "Going" : "Not Going"}
                    </div>
                )}


            </div>

            {/* Delete Event Confirmation Modal */}
            <Modal
                isOpen={showDeleteModal}
                onClose={() => setShowDeleteModal(false)}
                title="Delete Event"
                description="Are you sure you want to delete this event? This action cannot be undone."
                onConfirm={handleDeleteConfirm}
                confirmText="Delete"
                cancelText="Cancel"
                isLoading={isDeleting}
                loadingText="Deleting..."
            />
        </div>
    );
}
