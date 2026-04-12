"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import { apiClient } from "@/lib/apiClient";
import ConfirmModal from "./ConfirmModal";

interface TemplatePanelProps {
  open: boolean;
  onClose: () => void;
}

type Category = "display" | "display-default" | "lyrics";
type FileItem = { name: string; url: string; size: number };

const TABS: { key: Category; label: string; desc: string }[] = [
  { key: "display", label: "예배 화면", desc: "항목별 배경 이미지" },
  { key: "display-default", label: "기본 배경", desc: "Display 기본 배경 (Frame 2)" },
  { key: "lyrics", label: "가사 PDF", desc: "가사 PDF 배경 (Frame 1)" },
];

const BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL
  || (typeof window !== "undefined" ? window.location.origin : "http://localhost:8080");

export default function TemplatePanel({ open, onClose }: TemplatePanelProps) {
  const [tab, setTab] = useState<Category>("display");
  const [files, setFiles] = useState<FileItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [dragOver, setDragOver] = useState(false);
  const [uploadName, setUploadName] = useState("");
  const [confirmTarget, setConfirmTarget] = useState<string | null>(null);
  const [toastMsg, setToastMsg] = useState<{ msg: string; type: "error" | "info" } | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const showToast = (msg: string, type: "error" | "info" = "error") => {
    setToastMsg({ msg, type });
    setTimeout(() => setToastMsg(null), 3000);
  };

  const fetchFiles = useCallback(async () => {
    setLoading(true);
    try {
      const res = await apiClient.getTemplates(tab);
      setFiles(res.files || []);
    } catch {
      setFiles([]);
    }
    setLoading(false);
  }, [tab]);

  useEffect(() => {
    if (open) fetchFiles();
  }, [open, fetchFiles]);

  const handleUpload = async (fileList: FileList | null) => {
    if (!fileList || fileList.length === 0) return;
    const file = fileList[0];
    const ext = file.name.toLowerCase().split(".").pop();
    if (!["png", "jpg", "jpeg"].includes(ext || "")) {
      showToast("PNG/JPG 파일만 업로드 가능합니다.");
      return;
    }
    try {
      const name = tab === "display" ? uploadName.trim() || undefined : undefined;
      await apiClient.uploadTemplate(file, tab, name);
      setUploadName("");
      fetchFiles();
    } catch {
      showToast("업로드 실패");
    }
  };

  const handleDelete = async (name: string) => {
    try {
      await apiClient.deleteTemplate(tab, name);
      fetchFiles();
    } catch {
      showToast("삭제 실패");
    }
  };

  const onDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(true);
  };
  const onDragLeave = () => setDragOver(false);
  const onDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
    handleUpload(e.dataTransfer.files);
  };

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-[10600] flex items-center justify-center">
      {/* overlay */}
      <div
        className="absolute inset-0 bg-black/60"
        onClick={onClose}
      />

      {/* modal */}
      <div className="relative w-[680px] max-h-[80vh] bg-[#2c2c2c] rounded-xl flex flex-col overflow-hidden">
        {/* header */}
        <div className="flex justify-between items-center px-5 py-4 border-b border-white/10">
          <h3 className="m-0 text-white text-base font-semibold">배경 템플릿 관리</h3>
          <button
            onClick={onClose}
            className="bg-transparent border-none text-[#aaa] text-xl cursor-pointer leading-none hover:text-white transition-colors"
          >
            ✕
          </button>
        </div>

        {/* tabs */}
        <div className="flex border-b border-white/10 px-5">
          {TABS.map((t) => (
            <button
              key={t.key}
              onClick={() => setTab(t.key)}
              className={`px-4 py-2.5 bg-transparent border-none border-b-2 text-xs cursor-pointer transition-colors ${
                tab === t.key
                  ? "border-[#4a9eff] text-[#4a9eff] font-semibold"
                  : "border-transparent text-[#aaa] font-normal hover:text-white"
              }`}
            >
              {t.label}
            </button>
          ))}
        </div>

        {/* body */}
        <div className="flex-1 overflow-y-auto px-5 py-4">
          <div className="text-xs text-[#888] mb-3">
            {TABS.find((t) => t.key === tab)?.desc}
          </div>

          {/* upload zone */}
          <div
            onDragOver={onDragOver}
            onDragLeave={onDragLeave}
            onDrop={onDrop}
            onClick={() => fileInputRef.current?.click()}
            className={`border-2 border-dashed rounded-lg p-5 text-center cursor-pointer mb-4 transition-all ${
              dragOver
                ? "border-[#4a9eff] bg-[rgba(74,158,255,0.1)]"
                : "border-white/20 bg-transparent hover:border-white/40"
            }`}
          >
            <div className="text-[#aaa] text-xs">이미지를 드래그하거나 클릭하여 업로드</div>
            <div className="text-[#666] text-[11px] mt-1">PNG, JPG (최대 10MB)</div>
            <input
              ref={fileInputRef}
              type="file"
              accept="image/png,image/jpeg"
              className="hidden"
              onChange={(e) => handleUpload(e.target.files)}
            />
          </div>

          {/* display 카테고리: 항목명 입력 */}
          {tab === "display" && (
            <div className="mb-3 flex gap-2 items-center">
              <input
                value={uploadName}
                onChange={(e) => setUploadName(e.target.value)}
                placeholder="항목명 (예: 전주, 찬양) — 비워두면 파일명 사용"
                className="flex-1 px-2.5 py-1.5 bg-white/10 border border-white/20 rounded-md text-white text-xs outline-none placeholder:text-[#666]"
              />
            </div>
          )}

          {/* file grid */}
          {loading ? (
            <div className="text-[#888] text-center py-5">로딩 중...</div>
          ) : files.length === 0 ? (
            <div className="text-[#666] text-center py-5">배경 이미지가 없습니다.</div>
          ) : (
            <div
              className={`grid gap-3 ${
                tab === "display" ? "grid-cols-3" : "grid-cols-1"
              }`}
            >
              {files.map((f) => (
                <div
                  key={f.name}
                  className="relative rounded-lg overflow-hidden bg-[#1a1a1a]"
                >
                  <img
                    src={`${BASE_URL}${f.url}?t=${Date.now()}`}
                    alt={f.name}
                    className={`w-full object-cover block ${
                      tab === "display" ? "aspect-video" : "max-h-[200px] h-auto"
                    }`}
                  />
                  <div className="px-2 py-1.5 flex justify-between items-center">
                    <span className="text-[#ccc] text-[11px] overflow-hidden text-ellipsis whitespace-nowrap flex-1">
                      {f.name}
                    </span>
                    {tab === "display" && (
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          setConfirmTarget(f.name);
                        }}
                        className="bg-[rgba(255,60,60,0.2)] border-none text-[#ff6b6b] text-[11px] px-2 py-0.5 rounded cursor-pointer ml-1 flex-shrink-0 hover:bg-[rgba(255,60,60,0.35)] transition-colors"
                      >
                        삭제
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
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

      <ConfirmModal
        open={confirmTarget !== null}
        message={`"${confirmTarget}" 배경을 삭제하시겠습니까?`}
        confirmLabel="삭제"
        danger
        onConfirm={() => {
          if (confirmTarget) handleDelete(confirmTarget);
          setConfirmTarget(null);
        }}
        onCancel={() => setConfirmTarget(null)}
      />
    </div>
  );
}
