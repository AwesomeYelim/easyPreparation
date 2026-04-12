"use client";

import React, { createContext, useContext } from "react";
import { useGlobalWebSocket, WsMessage } from "@/lib/wsClient";

type WsContextValue = {
  ws: WebSocket | null;
  subscribe: (fn: (msg: WsMessage) => void) => () => void;
  isOpen: boolean;
};

const WebSocketContext = createContext<WsContextValue>(null!);

export const WebSocketProvider = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  const baseUrl = process.env.NEXT_PUBLIC_WS_URL
    || (typeof window !== 'undefined'
      ? `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}/ws`
      : 'ws://localhost:8080/ws');
  const wsData = useGlobalWebSocket(baseUrl);

  return (
    <WebSocketContext.Provider value={wsData}>
      {children}
    </WebSocketContext.Provider>
  );
};

export const useWS = () => useContext(WebSocketContext);
