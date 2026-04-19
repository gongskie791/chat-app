import { getAccessToken } from "@/api/client";
import type { ChatMessage, OnlineUser, OutgoingEvent } from "@/types";
import { useCallback, useEffect, useRef, useState } from "react";

interface UseChatOptions {
  roomId: string;
}

export function useChat({ roomId }: UseChatOptions) {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [onlineUsers, setOnlineUsers] = useState<OnlineUser[]>([]);
  const [typingUsers, setTypingUsers] = useState<string[]>([]);
  const [connected, setConnected] = useState(false);

  const wsRef = useRef<WebSocket | null>(null);
  const typingTimeouts = useRef<Map<string, ReturnType<typeof setTimeout>>>(
    new Map(),
  );

  useEffect(() => {
    const token = getAccessToken();
    const proto = window.location.protocol === "https:" ? "wss" : "ws";
    const url = `${proto}://${window.location.host}/api/ws/${roomId}?token=${token}`;

    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => setConnected(true);
    ws.onclose = () => setConnected(false);

    ws.onmessage = (e: MessageEvent) => {
      const evt: OutgoingEvent = JSON.parse(e.data as string);

      switch (evt.type) {
        case "history": {
          const history = evt.content as ChatMessage[];
          setMessages(history ?? []);
          break;
        }
        case "message": {
          const msg = evt.content as ChatMessage;
          setMessages((prev) => [...prev, msg]);
          break;
        }
        case "online_users": {
          setOnlineUsers((evt.content as OnlineUser[]) ?? []);
          break;
        }
        case "typing": {
          const username = evt.username!;
          setTypingUsers((prev) =>
            prev.includes(username) ? prev : [...prev, username],
          );
          // auto-clear after 3 s of silence
          const existing = typingTimeouts.current.get(username);
          if (existing) clearTimeout(existing);
          typingTimeouts.current.set(
            username,
            setTimeout(() => {
              setTypingUsers((prev) => prev.filter((u) => u !== username));
              typingTimeouts.current.delete(username);
            }, 3000),
          );
          break;
        }
      }
    };

    return () => {
      ws.close();
      typingTimeouts.current.forEach(clearTimeout);
    };
  }, [roomId]);

  const sendMessage = useCallback((content: string) => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) return;
    wsRef.current.send(JSON.stringify({ type: "message", content }));
  }, []);

  const sendTyping = useCallback(() => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) return;
    wsRef.current.send(JSON.stringify({ type: "typing", content: "" }));
  }, []);

  return { messages, onlineUsers, typingUsers, connected, sendMessage, sendTyping };
}
