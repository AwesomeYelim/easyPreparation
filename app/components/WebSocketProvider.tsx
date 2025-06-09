"use client";

import React, { createContext, useContext } from "react";
import { useGlobalWebSocket } from "@/lib/wsClient";

const WebSocketContext = createContext<any>(null);

export const WebSocketProvider = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  const baseUrl = process.env.NEXT_PUBLIC_WS_URL!;
  const wsData = useGlobalWebSocket(baseUrl);

  return (
    <WebSocketContext.Provider value={wsData}>
      {children}
    </WebSocketContext.Provider>
  );
};

export const useWS = () => useContext(WebSocketContext);
