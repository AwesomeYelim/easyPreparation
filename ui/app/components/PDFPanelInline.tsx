"use client";

import { useState, useEffect, useRef } from "react";

const BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "";

interface PDFPanelInlineProps {
  connected: boolean;
}

export default function PDFPanelInline({ connected }: PDFPanelInlineProps) {
  const [uploading, setUploading] = useState(false);
  const [slideCount, setSlideCount] = useState(0);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [toastMsg, setToastMsg] = useState<{ msg: string; type: "error" | "info" } | null>(null);
  const [isDragOver, setIsDragOver] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const showToast = (msg: string, type: "error" | "info" = "error") => {
    setToastMsg({ msg, type });
    setTimeout(() => setToastMsg(null), 3000);
  };

  const fetchStatus = async () => {
    try {
      const res = await fetch(`${BASE_URL}/api/pdf/slides`).then((r) => r.json());
      setSlideCount(res.count ?? 0);
      setCurrentIndex(res.currentIndex ?? 0);
    } catch {}
  };

  useEffect(() => {
    fetchStatus();
    pollRef.current = setInterval(fetchStatus, 3000);
    return () => {
      if (pollRef.current) clearInterval(pollRef.current);
    };
  }, []);

  const handleFileUpload = async (files: FileList | null) => {
    if (!files || files.length === 0) return;
    const file = files[0];
    if (!file.name.toLowerCase().endsWith(".pdf")) {
      showToast("PDF 파일만 업로드 가능합니다.");
      return;
    }
    setUploading(true);
    try {
      const fd = new FormData();
      fd.append("file", file);
      const res = await fetch(`${BASE_URL}/api/pdf/upload`, { method: "POST", body: fd }).then((r) =>
        r.json()
      );
      if (res.ok) {
        showToast("업로드 완료 — 슬라이드 로딩 중", "info");
        await fetchStatus();
      } else {
        showToast(res.error || "업로드 실패");
      }
    } catch {
      showToast("업로드 실패");
    }
    setUploading(false);
    if (fileInputRef.current) fileInputRef.current.value = "";
  };

  const navigate = async (action: "prev" | "next") => {
    try {
      const res = await fetch(`${BASE_URL}/api/pdf/navigate`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action }),
      }).then((r) => r.json());
      if (res.ok) {
        setCurrentIndex(res.currentIndex ?? currentIndex);
      }
    } catch {}
  };

  const handleAddOBSSource = async () => {
    try {
      const res = await fetch(`${BASE_URL}/api/obs/setup-display`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ url: "http://localhost:8080/display/pdf" }),
      }).then((r) => r.json());
      if (res.ok) {
        showToast("EP_PDF 소스가 씬에 추가되었습니다.", "info");
      } else {
        showToast(res.error || "OBS 소스 추가 실패");
      }
    } catch {
      showToast("OBS 소스 추가 실패");
    }
  };

  const handleReset = async () => {
    try {
      const res = await fetch(`${BASE_URL}/api/pdf/slides`, { method: "DELETE" }).then((r) =>
        r.json()
      );
      if (res.ok) {
        setSlideCount(0);
        setCurrentIndex(0);
        showToast("슬라이드 초기화 완료", "info");
      } else {
        showToast("초기화 실패");
      }
    } catch {
      showToast("초기화 실패");
    }
  };

  const btnBase =
    "px-4 py-2 rounded-md text-white text-xs font-semibold border-none cursor-pointer transition-opacity";
  const btnGray = `${btnBase} bg-white/20 hover:bg-white/30`;
  const btnDanger = `${btnBase} bg-[#c0392b] hover:bg-[#e74c3c]`;

  return (
    <div className="flex flex-col gap-4">
      {/* 업로드 */}
      <div>
        <div className="text-[11px] text-[#888] mb-2 font-semibold">PDF 업로드</div>
        <div
          className={`border-2 border-dashed rounded-lg p-5 text-center cursor-pointer transition-all ${
            isDragOver
              ? "border-[#4a9eff] bg-[rgba(74,158,255,0.1)]"
              : "border-white/20 hover:border-white/40"
          }`}
          onClick={() => !uploading && fileInputRef.current?.click()}
          onDragOver={(e) => { e.preventDefault(); setIsDragOver(true); }}
          onDragLeave={() => setIsDragOver(false)}
          onDrop={(e) => {
            e.preventDefault();
            setIsDragOver(false);
            handleFileUpload(e.dataTransfer.files);
          }}
        >
          <div className="text-[#aaa] text-xs">
            {uploading ? "업로드 중..." : "클릭하거나 PDF를 드래그하여 업로드"}
          </div>
          <div className="text-[#666] text-[10px] mt-1">최대 50 MB · 브라우저 렌더링</div>
          <input
            ref={fileInputRef}
            type="file"
            accept=".pdf"
            className="hidden"
            onChange={(e) => handleFileUpload(e.target.files)}
            disabled={uploading}
          />
        </div>
      </div>

      {/* 슬라이드 제어 */}
      <div className="bg-white/[0.06] rounded-lg p-4 border border-white/[0.08]">
        <div className="text-[11px] text-[#888] mb-3 font-semibold">슬라이드 제어</div>
        {slideCount === 0 ? (
          <div className="text-[#666] text-xs text-center py-2">슬라이드 없음</div>
        ) : (
          <div className="flex items-center justify-between gap-3">
            <button
              className={`${btnGray} ${currentIndex === 0 ? "opacity-40 cursor-default" : ""}`}
              onClick={() => navigate("prev")}
              disabled={currentIndex === 0}
            >
              ◀ 이전
            </button>
            <div className="text-white text-sm font-bold">
              {currentIndex + 1} / {slideCount}
            </div>
            <button
              className={`${btnGray} ${currentIndex >= slideCount - 1 ? "opacity-40 cursor-default" : ""}`}
              onClick={() => navigate("next")}
              disabled={currentIndex >= slideCount - 1}
            >
              다음 ▶
            </button>
          </div>
        )}
      </div>

      {/* 액션 버튼 */}
      <div className="flex flex-col gap-2">
        <button
          className={`px-4 py-2 rounded-md text-white text-xs font-semibold border-none transition-opacity ${
            connected
              ? "bg-[#204d87] hover:bg-[#2d5a8a] cursor-pointer"
              : "bg-white/20 opacity-50 cursor-not-allowed"
          }`}
          onClick={connected ? handleAddOBSSource : undefined}
          disabled={!connected}
          title={connected ? undefined : "OBS 연결 후 사용 가능"}
        >
          OBS 소스 추가 (EP_PDF)
        </button>
        <button
          className={`${btnDanger} ${slideCount === 0 ? "opacity-40 cursor-default" : ""}`}
          onClick={handleReset}
          disabled={slideCount === 0}
        >
          슬라이드 초기화
        </button>
      </div>

      {/* toast */}
      {toastMsg && (
        <div
          className={`fixed top-6 left-1/2 -translate-x-1/2 px-6 py-2.5 rounded-lg text-white text-xs z-[11100] shadow-lg ${
            toastMsg.type === "error" ? "bg-[#dc3545]" : "bg-[#204d87]"
          }`}
        >
          {toastMsg.msg}
        </div>
      )}
    </div>
  );
}
