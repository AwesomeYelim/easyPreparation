"use client";

import { useState, useRef } from "react";

interface Props {
  imageUrl?: string;
  loading?: boolean;
  onFile: (f: File) => void;
  onClear?: () => void;
  height?: number;
  placeholder?: string;
}

export default function DarkImageDropZone({
  imageUrl,
  loading,
  onFile,
  onClear,
  height = 100,
  placeholder,
}: Props) {
  const [dragging, setDragging] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  return (
    <div
      className={`relative border-2 rounded-lg cursor-pointer flex items-center justify-center overflow-hidden transition-all ${
        dragging
          ? "border-[#4a9eff] bg-[rgba(74,158,255,0.1)] border-solid"
          : imageUrl
          ? "border-white/20 border-solid"
          : "border-white/15 border-dashed bg-white/5 hover:border-white/30"
      }`}
      style={{ height }}
      onDragOver={(e) => {
        e.preventDefault();
        e.stopPropagation();
        if (!dragging) setDragging(true);
      }}
      onDragLeave={(e) => {
        e.stopPropagation();
        setDragging(false);
      }}
      onDrop={(e) => {
        e.preventDefault();
        e.stopPropagation();
        setDragging(false);
        const file = e.dataTransfer.files[0];
        if (file && file.type.startsWith("image/")) onFile(file);
      }}
      onClick={(e) => {
        e.stopPropagation();
        inputRef.current?.click();
      }}
    >
      {loading ? (
        <span className="text-[10px] text-[#888] pointer-events-none text-center px-3">
          업로드 중...
        </span>
      ) : imageUrl ? (
        <>
          <img src={imageUrl} alt="" className="w-full h-full object-cover" />
          {onClear && (
            <button
              className="absolute top-0 right-0 w-9 h-9 bg-black/50 text-white border-none text-sm cursor-pointer flex items-center justify-center hover:bg-[rgba(239,68,68,0.8)] transition-colors rounded-bl-lg"
              onClick={(e) => {
                e.stopPropagation();
                onClear();
              }}
              title="기본 배경으로"
            >
              &times;
            </button>
          )}
        </>
      ) : (
        <span className="text-[10px] text-[#666] pointer-events-none text-center px-3">
          {placeholder || "이미지를 드래그하거나 클릭"}
        </span>
      )}
      <input
        ref={inputRef}
        type="file"
        accept="image/png,image/jpeg"
        className="hidden"
        onChange={(e) => {
          const f = e.target.files?.[0];
          if (f) onFile(f);
          e.target.value = "";
        }}
      />
    </div>
  );
}
