import ChatShell from "@/components/chat-shell";

type ChatPageProps = {
  params: {
    chatId: string;
  };
};

export default function ChatPage({ params }: ChatPageProps) {
  const parsed = Number(params.chatId);
  const chatId = Number.isFinite(parsed) ? parsed : undefined;

  return <ChatShell activeChatId={chatId} />;
}
