import { useEffect, useRef, useState } from "react";

interface RecevedData {
  code: number;
  message: string;
  target: string;
  type: string;
}

export function useGlobalWebSocket(url: string) {
  const ws = useRef<WebSocket | null>(null);
  const [message, setMessage] = useState<RecevedData>();
  const [isOpen, setIsOpen] = useState(false);

  useEffect(() => {
    let reconnectTimer: ReturnType<typeof setTimeout>;

    const connect = () => {
      const socket = new WebSocket(url);
      ws.current = socket;

      const pingInterval = setInterval(() => {
        if (ws.current?.readyState === WebSocket.OPEN) {
          ws.current.send(JSON.stringify({ type: "ping" }));
        }
      }, 30000);

      socket.onopen = () => {
        setIsOpen(true);
        console.log("WebSocket connected");
      };

      socket.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          setMessage(data);
        } catch (e) {
          console.warn("Invalid WS message:", event.data);
          setMessage(event.data);
        }
      };

      socket.onclose = () => {
        setIsOpen(false);
        clearInterval(pingInterval);
        console.log("WebSocket closed");

        // 5초 후 재연결
        reconnectTimer = setTimeout(() => {
          connect();
        }, 5000);
      };
    };

    connect();

    return () => {
      clearTimeout(reconnectTimer);
      ws.current?.close();
    };
  }, [url]);

  return { ws: ws.current, message, isOpen };
}
