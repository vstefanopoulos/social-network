/**
 * Converts a timestamp into a relative time format (e.g., "2 mins ago", "1 hour ago")
 * @param {string} timestamp - ISO 8601 timestamp string
 * @returns {string} Relative time string
 */
export function getRelativeTime(timestamp) {
    const now = new Date();
    const past = new Date(timestamp);
    const diffInSeconds = Math.floor((now - past) / 1000);

    // Just now (less than 1 minute)
    if (diffInSeconds < 60) {
        return "just now";
    }

    // Minutes (less than 1 hour)
    const diffInMinutes = Math.floor(diffInSeconds / 60);
    if (diffInMinutes < 60) {
        return diffInMinutes === 1 ? "1 min ago" : `${diffInMinutes} mins ago`;
    }

    // Hours (less than 24 hours)
    const diffInHours = Math.floor(diffInMinutes / 60);
    if (diffInHours < 24) {
        return diffInHours === 1 ? "1 hour ago" : `${diffInHours} hours ago`;
    }

    // Days (less than 7 days)
    const diffInDays = Math.floor(diffInHours / 24);
    if (diffInDays < 7) {
        return diffInDays === 1 ? "1 day ago" : `${diffInDays} days ago`;
    }

    // Weeks (less than 4 weeks)
    const diffInWeeks = Math.floor(diffInDays / 7);
    if (diffInWeeks < 4) {
        return diffInWeeks === 1 ? "1 week ago" : `${diffInWeeks} weeks ago`;
    }

    // Months (less than 12 months)
    const diffInMonths = Math.floor(diffInDays / 30);
    if (diffInMonths < 12) {
        return diffInMonths === 1 ? "1 month ago" : `${diffInMonths} months ago`;
    }

    // Years
    const diffInYears = Math.floor(diffInDays / 365);
    return diffInYears === 1 ? "1 year ago" : `${diffInYears} years ago`;
}
