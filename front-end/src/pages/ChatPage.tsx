import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useAuth } from "@/context/AuthContext";
import { useChat } from "@/hooks/useChat";
import type { ChatMessage } from "@/types";
import { ArrowLeft, Circle, Hash, Send } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { cn } from "@/lib/utils";

export default function ChatPage() {
  const { roomId } = useParams<{ roomId: string }>();
  const { user } = useAuth();
  const navigate = useNavigate();

  const { messages, onlineUsers, typingUsers, connected, sendMessage, sendTyping } =
    useChat({ roomId: roomId! });

  const [input, setInput] = useState("");
  const bottomRef = useRef<HTMLDivElement>(null);
  const typingTimeout = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Auto-scroll to bottom on new messages
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  function handleSend(e: React.FormEvent) {
    e.preventDefault();
    const text = input.trim();
    if (!text) return;
    sendMessage(text);
    setInput("");
  }

  function handleInputChange(e: React.ChangeEvent<HTMLInputElement>) {
    setInput(e.target.value);
    sendTyping();
    // throttle — don't send another typing event for 2 s
    if (typingTimeout.current) clearTimeout(typingTimeout.current);
    typingTimeout.current = setTimeout(() => {
      typingTimeout.current = null;
    }, 2000);
  }

  const visibleTypers = typingUsers.filter((u) => u !== user?.username);

  return (
    <div className="flex h-screen bg-background">
      {/* Sidebar — online users */}
      <aside className="hidden w-56 shrink-0 flex-col border-r border-border bg-card lg:flex">
        <div className="flex items-center gap-2 border-b border-border px-4 py-3">
          <Circle
            className={cn(
              "h-2 w-2 fill-current",
              connected ? "text-green-500" : "text-muted-foreground",
            )}
          />
          <span className="text-xs font-medium text-muted-foreground">
            {connected ? "Connected" : "Connecting…"}
          </span>
        </div>

        <div className="p-3">
          <p className="mb-2 px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Online — {onlineUsers.length}
          </p>
          <ul className="space-y-0.5">
            {onlineUsers.map((u) => (
              <li
                key={u.user_id}
                className="flex items-center gap-2 rounded-md px-2 py-1.5"
              >
                <Circle className="h-1.5 w-1.5 fill-green-500 text-green-500" />
                <span
                  className={cn(
                    "truncate text-sm",
                    u.username === user?.username
                      ? "font-medium text-foreground"
                      : "text-muted-foreground",
                  )}
                >
                  {u.username}
                  {u.username === user?.username && (
                    <span className="ml-1 text-xs text-muted-foreground">
                      (you)
                    </span>
                  )}
                </span>
              </li>
            ))}
          </ul>
        </div>
      </aside>

      {/* Main chat area */}
      <div className="flex flex-1 flex-col">
        {/* Header */}
        <header className="flex items-center gap-3 border-b border-border px-4 py-3">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => navigate("/rooms")}
            className="shrink-0"
          >
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <Hash className="h-4 w-4 shrink-0 text-muted-foreground" />
          <span className="font-semibold text-foreground">
            {roomId?.slice(0, 8)}…
          </span>
          <span className="ml-auto text-xs text-muted-foreground lg:hidden">
            {onlineUsers.length} online
          </span>
        </header>

        {/* Messages */}
        <ScrollArea className="flex-1 px-4 py-4">
          <div className="space-y-1">
            {messages.map((msg) => (
              <MessageBubble
                key={msg.id}
                msg={msg}
                isOwn={msg.user_id === user?.id}
              />
            ))}
          </div>

          {/* Typing indicator */}
          {visibleTypers.length > 0 && (
            <div className="mt-2 flex items-center gap-1.5 px-1">
              <TypingDots />
              <span className="text-xs text-muted-foreground">
                {visibleTypers.length === 1
                  ? `${visibleTypers[0]} is typing…`
                  : `${visibleTypers.join(", ")} are typing…`}
              </span>
            </div>
          )}

          <div ref={bottomRef} />
        </ScrollArea>

        {/* Input bar */}
        <form
          onSubmit={handleSend}
          className="flex items-center gap-2 border-t border-border px-4 py-3"
        >
          <Input
            placeholder="Message…"
            value={input}
            onChange={handleInputChange}
            autoComplete="off"
            disabled={!connected}
            className="flex-1"
          />
          <Button
            type="submit"
            size="icon"
            disabled={!input.trim() || !connected}
          >
            <Send className="h-4 w-4" />
          </Button>
        </form>
      </div>
    </div>
  );
}

// ─── Sub-components ───────────────────────────────────────────────────────────

function MessageBubble({
  msg,
  isOwn,
}: {
  msg: ChatMessage;
  isOwn: boolean;
}) {
  const time = new Date(msg.timestamp).toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
  });

  return (
    <div className={cn("flex flex-col", isOwn ? "items-end" : "items-start")}>
      {!isOwn && (
        <span className="mb-0.5 px-1 text-xs font-medium text-muted-foreground">
          {msg.username}
        </span>
      )}
      <div
        className={cn(
          "max-w-[70%] rounded-2xl px-3.5 py-2 text-sm leading-relaxed",
          isOwn
            ? "rounded-br-sm bg-primary text-primary-foreground"
            : "rounded-bl-sm bg-secondary text-secondary-foreground",
        )}
      >
        {msg.content}
      </div>
      <span className="mt-0.5 px-1 text-[10px] text-muted-foreground/60">
        {time}
      </span>
    </div>
  );
}

function TypingDots() {
  return (
    <div className="flex items-center gap-0.5">
      {[0, 1, 2].map((i) => (
        <span
          key={i}
          className="h-1 w-1 animate-bounce rounded-full bg-muted-foreground"
          style={{ animationDelay: `${i * 0.15}s` }}
        />
      ))}
    </div>
  );
}
