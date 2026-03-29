import { useEffect, useRef, useState, useCallback } from "react";

export interface WsMessage {
  code: number;
  message: string;
  target: string;
  type: string;
  [key: string]: any;
}

export function useGlobalWebSocket(url: string) {
  const ws = useRef<WebSocket | null>(null);
  const [isOpen, setIsOpen] = useState(false);
  const listenersRef = useRef<Set<(msg: WsMessage) => void>>(new Set());

  const subscribe = useCallback((fn: (msg: WsMessage) => void) => {
    listenersRef.current.add(fn);
    return () => { listenersRef.current.delete(fn); };
  }, []);

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
      };

      socket.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          listenersRef.current.forEach((fn) => fn(data));
        } catch (e) {
          console.warn("Invalid WS message:", event.data);
        }
      };

      socket.onclose = () => {
        setIsOpen(false);
        clearInterval(pingInterval);
        reconnectTimer = setTimeout(() => connect(), 5000);
      };
    };

    connect();

    return () => {
      clearTimeout(reconnectTimer);
      ws.current?.close();
    };
  }, [url]);

  return { ws: ws.current, subscribe, isOpen };
}
