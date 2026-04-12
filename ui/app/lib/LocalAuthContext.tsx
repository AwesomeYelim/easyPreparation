"use client";
import { createContext, useContext, useEffect, useState, ReactNode } from "react";

interface Church {
  id: number;
  name: string;
  englishName: string;
  email: string;
}

interface AuthContextType {
  church: Church | null;
  needsSetup: boolean;
  isLoading: boolean;
  setupError: string | null;
  completeSetup: (name: string, englishName: string) => Promise<void>;
  updateChurch: (patch: Partial<Church>) => void;
}

const AuthContext = createContext<AuthContextType>({
  church: null,
  needsSetup: false,
  isLoading: true,
  setupError: null,
  completeSetup: async () => {},
  updateChurch: () => {},
});

export function useAuth() {
  return useContext(AuthContext);
}

const BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";

export function LocalAuthProvider({ children }: { children: ReactNode }) {
  const [church, setChurch] = useState<Church | null>(null);
  const [needsSetup, setNeedsSetup] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [setupError, setSetupError] = useState<string | null>(null);

  const fetchStatus = async () => {
    try {
      const res = await fetch(`${BASE_URL}/api/setup/status`);
      const data = await res.json();
      if (data.needsSetup) {
        setNeedsSetup(true);
      } else {
        // Go 핸들러가 englishName(camelCase)으로 반환하므로 그대로 사용
        setChurch(data.church);
        setNeedsSetup(false);
      }
    } catch (err) {
      console.error("Setup status fetch failed:", err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => { fetchStatus(); }, []);

  const completeSetup = async (name: string, englishName: string) => {
    setSetupError(null);
    const res = await fetch(`${BASE_URL}/api/setup`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name, englishName }),
    });
    if (!res.ok) {
      let msg = `서버 오류 (${res.status})`;
      try {
        const errData = await res.json();
        if (errData.error) msg = errData.error;
      } catch {
        msg = await res.text().catch(() => msg);
      }
      setSetupError(msg);
      throw new Error(msg);
    }
    await fetchStatus();
  };

  const updateChurch = (patch: Partial<Church>) => {
    setChurch((prev) => prev ? { ...prev, ...patch } : prev);
  };

  return (
    <AuthContext.Provider value={{ church, needsSetup, isLoading, setupError, completeSetup, updateChurch }}>
      {children}
    </AuthContext.Provider>
  );
}
