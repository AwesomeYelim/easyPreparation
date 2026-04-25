"use client";

import { useState, useEffect } from "react";
import { Toaster } from "react-hot-toast";

export default function ToastProvider() {
  const [mounted, setMounted] = useState(false);
  useEffect(() => setMounted(true), []);
  if (!mounted) return null;

  return (
    <Toaster
      position="top-right"
      toastOptions={{
        duration: 4000,
        style: {
          borderRadius: "8px",
          background: "var(--surface-elevated, #fff)",
          color: "var(--text-primary, #1f2937)",
        },
      }}
    />
  );
}
