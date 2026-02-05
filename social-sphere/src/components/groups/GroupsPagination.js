"use client";

export default function GroupsPagination({ currentPage, totalItems, itemsPerPage, onPageChange }) {
    const totalPages = Math.ceil(totalItems / itemsPerPage);

    if (totalPages <= 1) return null;

    const getPageNumbers = () => {
        const pages = [];
        const maxVisible = 5;

        if (totalPages <= maxVisible) {
            for (let i = 1; i <= totalPages; i++) {
                pages.push(i);
            }
        } else {
            if (currentPage <= 3) {
                for (let i = 1; i <= 4; i++) pages.push(i);
                pages.push('...');
                pages.push(totalPages);
            } else if (currentPage >= totalPages - 2) {
                pages.push(1);
                pages.push('...');
                for (let i = totalPages - 3; i <= totalPages; i++) pages.push(i);
            } else {
                pages.push(1);
                pages.push('...');
                for (let i = currentPage - 1; i <= currentPage + 1; i++) pages.push(i);
                pages.push('...');
                pages.push(totalPages);
            }
        }

        return pages;
    };

    const pageNumbers = getPageNumbers();

    return (
        <div className="flex items-center justify-center gap-2 mt-8">
            <button
                onClick={() => onPageChange(currentPage - 1)}
                disabled={currentPage === 1}
                className="px-3 py-2 rounded-lg border border-(--border) text-sm font-medium text-(--muted) hover:text-foreground hover:bg-(--muted)/10 transition-colors disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
            >
                Previous
            </button>

            {pageNumbers.map((page, index) => (
                page === '...' ? (
                    <span key={`ellipsis-${index}`} className="px-2 text-(--muted)">...</span>
                ) : (
                    <button
                        key={page}
                        onClick={() => onPageChange(page)}
                        className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors cursor-pointer ${
                            currentPage === page
                                ? 'bg-(--accent) text-white'
                                : 'border border-(--border) text-(--muted) hover:text-foreground hover:bg-(--muted)/10'
                        }`}
                    >
                        {page}
                    </button>
                )
            ))}

            <button
                onClick={() => onPageChange(currentPage + 1)}
                disabled={currentPage === totalPages}
                className="px-3 py-2 rounded-lg border border-(--border) text-sm font-medium text-(--muted) hover:text-foreground hover:bg-(--muted)/10 transition-colors disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
            >
                Next
            </button>
        </div>
    );
}
