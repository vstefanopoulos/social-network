import { MessageCircle } from "lucide-react";

export default function MsgPage() {
    return (
        <div className="flex-1 flex flex-col items-center justify-center px-4 hidden md:flex">
            <div className="w-20 h-20 rounded-full bg-(--muted)/10 flex items-center justify-center mb-4">
                <MessageCircle className="w-10 h-10 text-(--muted)" />
            </div>
            <h2 className="text-xl font-semibold text-foreground mb-2">Your Messages</h2>
            <p className="text-(--muted) text-center max-w-sm">
                Select a conversation from the list to start chatting
            </p>
        </div>
    );
}
