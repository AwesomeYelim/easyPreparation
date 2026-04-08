"use client";

import { useEffect } from "react";

interface ConfirmModalProps {
  open: boolean;
  message: string;
  onConfirm: () => void;
  onCancel: () => void;
  confirmLabel?: string;
  cancelLabel?: string;
  danger?: boolean;
}

export default function ConfirmModal({
  open,
  message,
  onConfirm,
  onCancel,
  confirmLabel = "확인",
  cancelLabel = "취소",
  danger = false,
}: ConfirmModalProps) {
  // ESC 키로 닫기
  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === "Escape") onCancel();
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [open, onCancel]);

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-[11000] flex items-center justify-center">
      {/* backdrop */}
      <div
        className="absolute inset-0 bg-black/50 backdrop-blur-sm"
        onClick={onCancel}
      />

      {/* dialog */}
      <div className="relative bg-[#2c2c2c] rounded-xl px-7 py-6 min-w-[320px] max-w-[400px] shadow-2xl animate-[confirmFadeIn_0.15s_ease-out]">
        <p className="text-[#e8e8e8] text-sm leading-relaxed mb-5 break-keep">
          {message}
        </p>

        <div className="flex gap-2 justify-end">
          <button
            onClick={onCancel}
            className="px-4 py-2 text-sm bg-white/10 border border-white/15 rounded-lg text-[#ccc] cursor-pointer hover:bg-white/[0.18] transition-colors"
          >
            {cancelLabel}
          </button>
          <button
            onClick={onConfirm}
            className={`px-4 py-2 text-sm font-semibold rounded-lg text-white cursor-pointer transition-colors ${
              danger
                ? "bg-[#dc3545] hover:bg-[#c82333]"
                : "bg-[#204d87] hover:bg-[#1a3d6e]"
            }`}
          >
            {confirmLabel}
          </button>
        </div>
      </div>

      {/* animation keyframes */}
      <style>{`
        @keyframes confirmFadeIn {
          from { opacity: 0; transform: scale(0.95); }
          to { opacity: 1; transform: scale(1); }
        }
      `}</style>
    </div>
  );
}
