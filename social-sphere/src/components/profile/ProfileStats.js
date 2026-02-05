export default function ProfileStats({ stats }) {
    return (
        <div className="flex items-center gap-6 md:gap-8">
            <button className="flex flex-col items-center md:items-start group cursor-pointer">
                <span className="text-xl font-bold text-foreground group-hover:text-(--accent) transition-colors">
                    {stats.followers.toLocaleString()}
                </span>
                <span className="text-xs font-medium text-(--muted) uppercase tracking-wide group-hover:text-foreground transition-colors">
                    Followers
                </span>
            </button>

            <button className="flex flex-col items-center md:items-start group cursor-pointer">
                <span className="text-xl font-bold text-foreground group-hover:text-(--accent) transition-colors">
                    {stats.following.toLocaleString()}
                </span>
                <span className="text-xs font-medium text-(--muted) uppercase tracking-wide group-hover:text-foreground transition-colors">
                    Following
                </span>
            </button>

            <button className="flex flex-col items-center md:items-start group cursor-pointer">
                <span className="text-xl font-bold text-foreground group-hover:text-(--accent) transition-colors">
                    {stats.groups.toLocaleString()}
                </span>
                <span className="text-xs font-medium text-(--muted) uppercase tracking-wide group-hover:text-foreground transition-colors">
                    Groups
                </span>
            </button>
        </div>
    );
}
