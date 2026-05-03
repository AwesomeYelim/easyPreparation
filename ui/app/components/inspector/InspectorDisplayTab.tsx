"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import { apiClient } from "@/lib/apiClient";
import ConfirmModal from "@/components/ConfirmModal";

type FileItem = { name: string; url: string; size: number };

interface Props {
  baseUrl: string;
}

export default function InspectorDisplayTab({ baseUrl }: Props) {
  const [files, setFiles] = useState<FileItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [dragOver, setDragOver] = useState(false);
  const [uploadName, setUploadName] = useState("");
  const [confirmTarget, setConfirmTarget] = useState<string | null>(null);
  const [toastMsg, setToastMsg] = useState<{ msg: string; type: "error" | "info" } | null>(null);
  const [imgRevision, setImgRevision] = useState(0);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const showToast = (msg: string, type: "error" | "info" = "error") => {
    setToastMsg({ msg, type });
    setTimeout(() => setToastMsg(null), 3000);
  };

  const fetchFiles = useCallback(async () => {
    setLoading(true);
    try {
      const res = await apiClient.getTemplates("display");
      setFiles(res.files || []);
      setImgRevision((r) => r + 1);
    } catch {
      setFiles([]);
    }
    setLoading(false);
  }, []);

  // 마운트 시 자동 로드
  useEffect(() => {
    fetchFiles();
  }, [fetchFiles]);

  const handleUpload = async (fileList: FileList | null) => {
    if (!fileList || fileList.length === 0) return;
    const file = fileList[0];
    const ext = file.name.toLowerCase().split(".").pop();
    if (!["png", "jpg", "jpeg"].includes(ext || "")) {
      showToast("PNG/JPG 파일만 업로드 가능합니다.");
      return;
    }
    try {
      const name = uploadName.trim() || undefined;
      await apiClient.uploadTemplate(file, "display", name);
      setUploadName("");
      fetchFiles();
    } catch {
      showToast("업로드 실패");
    }
  };

  const handleDelete = async (name: string) => {
    try {
      await apiClient.deleteTemplate("display", name);
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

  return (
    <>
      {/* upload zone */}
      <div
        onDragOver={onDragOver}
        onDragLeave={onDragLeave}
        onDrop={onDrop}
        onClick={() => fileInputRef.current?.click()}
        className={`border-2 border-dashed rounded-lg p-4 text-center cursor-pointer mb-3 transition-all ${
          dragOver
            ? "border-[#4a9eff] bg-[rgba(74,158,255,0.1)]"
            : "border-white/20 bg-transparent hover:border-white/40"
        }`}
      >
        <div className="text-[#aaa] text-[11px]">이미지를 드래그하거나 클릭하여 업로드</div>
        <div className="text-[#666] text-[10px] mt-1">PNG, JPG (최대 10MB)</div>
        <input
          ref={fileInputRef}
          type="file"
          accept="image/png,image/jpeg"
          className="hidden"
          onChange={(e) => handleUpload(e.target.files)}
        />
      </div>

      {/* 항목명 입력 */}
      <div className="mb-2 flex gap-2 items-center">
        <input
          value={uploadName}
          onChange={(e) => setUploadName(e.target.value)}
          placeholder="항목명 (예: 전주, 찬양) — 비워두면 파일명 사용"
          className="flex-1 px-2 py-1.5 bg-white/10 border border-white/20 rounded-md text-white text-[10px] outline-none placeholder:text-[#666]"
        />
      </div>

      {/* file grid */}
      {loading ? (
        <div className="text-[#888] text-center py-5 text-xs">로딩 중...</div>
      ) : files.length === 0 ? (
        <div className="text-[#666] text-center py-5 text-xs">배경 이미지가 없습니다.</div>
      ) : (
        <div className="grid gap-2 grid-cols-2">
          {files.map((f) => (
            <div
              key={f.name}
              className="relative rounded-lg overflow-hidden bg-[#1a1a1a]"
            >
              <img
                src={`${baseUrl}${f.url}?v=${imgRevision}`}
                alt={f.name}
                className="w-full object-cover block aspect-video"
              />
              <div className="px-2 py-1 flex justify-between items-center">
                <span className="text-[#ccc] text-[10px] overflow-hidden text-ellipsis whitespace-nowrap flex-1">
                  {f.name}
                </span>
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    setConfirmTarget(f.name);
                  }}
                  className="bg-[rgba(255,60,60,0.2)] border-none text-[#ff6b6b] text-[10px] px-1.5 py-0.5 rounded cursor-pointer ml-1 flex-shrink-0 hover:bg-[rgba(255,60,60,0.35)] transition-colors"
                >
                  삭제
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* toast */}
      {toastMsg && (
        <div
          className={`absolute bottom-4 left-1/2 -translate-x-1/2 px-4 py-2 rounded-lg text-white text-[10px] shadow-lg z-10 ${
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
    </>
  );
}
