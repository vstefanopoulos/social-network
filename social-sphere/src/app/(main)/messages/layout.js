import  ConversationsContent from "@/components/messages/ConversationsContent"
import { getConv } from "@/actions/chat/get-conv";

export default async function MessageLayout({ children }) {

    let conversations = [];
    const result = await getConv({ first: true, limit: 15 });
    if (result.success && result.data) {
        conversations = result.data;
    }

    return (
        <ConversationsContent conversations={conversations} >
            {children}
        </ConversationsContent>
    );
}