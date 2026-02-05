/**
 * Responsive Container Component
 * Provides consistent max-width and padding across different screen sizes
 * Following modern design best practices for readability and UX
 */

export default function Container({ children, size = "default", className = "" }) {
    // Size variants optimized for different content types
    const sizeClasses = {
        // Narrow: Best for forms and focused content (400-600px)
        narrow: "max-w-md",

        // Default: Optimal for feed/social content (600-800px)
        // Perfect for reading - follows 65-75 characters per line rule
        default: "max-w-3xl",

        // Wide: For dashboards and multi-column layouts (1200-1400px)
        wide: "max-w-7xl",

        // Full: Takes full width with padding (useful for special cases)
        full: "max-w-full",
    };

    return (
        <div className={`
            mx-auto
            w-full
            px-4
            sm:px-6
            lg:px-0
            ${sizeClasses[size]}
            ${className}
        `}>
            {children}
        </div>
    );
}
