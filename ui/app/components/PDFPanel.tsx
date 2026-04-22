"use client";

import { useState, useEffect, useRef } from "react";

const BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "";

interface PDFPanelProps {
  open: boolean;
  onClose: () => void;
}

export default function PDFPanel({ open, onClose }: PDFPanelProps) {
  const [uploading, setUploading] = useState(false);
  const [slideCount, setSlideCount] = useState(0);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [toastMsg, setToastMsg] = useState<{ msg: string; type: "error" | "info" } | null>(null);
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
    if (open) {
      fetchStatus();
      pollRef.current = setInterval(fetchStatus, 3000);
    } else {
      if (pollRef.current) clearInterval(pollRef.current);
    }
    return () => {
      if (pollRef.current) clearInterval(pollRef.current);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open]);

  const handleFileChange = async (files: FileList | null) => {
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
        showToast("업로드 완료 — OBS 소스에서 페이지 수 로딩 중", "info");
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
        showToast(`EP_PDF 소스가 씬에 추가되었습니다.`, "info");
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

  if (!open) return null;

  const btnBase =
    "px-4 py-2 rounded-md text-white text-xs font-semibold border-none cursor-pointer transition-opacity";
  const btnPrimary = `${btnBase} bg-[#204d87] hover:bg-[#2d5a8a]`;
  const btnGray = `${btnBase} bg-white/20 hover:bg-white/30`;
  const btnDanger = `${btnBase} bg-[#c0392b] hover:bg-[#e74c3c]`;

  return (
    <div className="fixed inset-0 z-[10600] flex items-center justify-center">
      {/* overlay */}
      <div className="absolute inset-0 bg-black/60" onClick={onClose} />

      {/* modal */}
      <div className="relative w-[440px] bg-[#6b7280] rounded-xl flex flex-col overflow-hidden shadow-2xl">
        {/* header */}
        <div className="flex justify-between items-center px-5 py-4 border-b border-white/20">
          <h3 className="m-0 text-white text-base font-semibold">외부 PDF</h3>
          <button
            onClick={onClose}
            className="bg-transparent border-none text-white text-xl cursor-pointer leading-none"
          >
            ✕
          </button>
        </div>

        {/* body */}
        <div className="px-5 py-5 flex flex-col gap-4">
          {/* 업로드 */}
          <div>
            <div className="text-xs text-white/70 mb-2 font-semibold">PDF 업로드</div>
            <div
              className="border-2 border-dashed border-white/30 rounded-lg p-5 text-center cursor-pointer hover:border-white/60 transition-colors"
              onClick={() => !uploading && fileInputRef.current?.click()}
            >
              <div className="text-white/70 text-xs">
                {uploading ? "변환 중... (Ghostscript 필요)" : "클릭하여 PDF 파일 선택"}
              </div>
              <div className="text-white/40 text-[10px] mt-1">최대 50 MB · Ghostscript 설치 필요</div>
              <input
                ref={fileInputRef}
                type="file"
                accept=".pdf"
                className="hidden"
                onChange={(e) => handleFileChange(e.target.files)}
                disabled={uploading}
              />
            </div>
          </div>

          {/* 슬라이드 상태 + 제어 */}
          <div className="bg-white/10 rounded-lg p-4">
            <div className="text-xs text-white/70 mb-3 font-semibold">슬라이드 제어</div>
            {slideCount === 0 ? (
              <div className="text-white/40 text-xs text-center py-2">슬라이드 없음</div>
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
            <button className={btnPrimary} onClick={handleAddOBSSource}>
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
        </div>
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
