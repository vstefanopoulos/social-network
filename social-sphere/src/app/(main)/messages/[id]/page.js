import { getMessages } from "@/actions/chat/get-messages";
import { markAsRead } from "@/actions/chat/mark-read";
import MessagesContent from "@/components/messages/MessagesContent";

export default async function ConversationPage({ params }) {
    const { id } = await params;
    let firstMessage = false;
    let initialMessages = [];

    // Get messages by interlocutor ID
    const messagesResult = await getMessages({
        interlocutorId: id,
        limit: 20,
    });

    if (messagesResult.success && messagesResult.data?.Messages?.length > 0) {
        // Messages exist - reverse for display (newest first from API)
        initialMessages = messagesResult.data.Messages.reverse();

        // Get conversation_id from the last message and mark as read on server
        const lastMessage = initialMessages[initialMessages.length - 1];
        if (lastMessage?.conversation_id && lastMessage?.id) {
            await markAsRead({ convID: lastMessage.conversation_id, lastMsgID: lastMessage.id });
        }
    } else {
        // No messages found - this is a new conversation
        firstMessage = true;
    }

    return (
        <MessagesContent
            interlocutorId={id}
            initialMessages={initialMessages}
            firstMessage={firstMessage}
        />
    );
}
