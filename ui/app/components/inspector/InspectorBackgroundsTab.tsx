"use client";

import { useState, useEffect, useRef } from "react";
import { apiClient } from "@/lib/apiClient";

import type { FileItem } from "@/types";

interface Props {
  baseUrl: string;
}

export default function InspectorBackgroundsTab({ baseUrl }: Props) {
  const [defaultFiles, setDefaultFiles] = useState<FileItem[]>([]);
  const [lyricsFiles, setLyricsFiles] = useState<FileItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [imgRevision, setImgRevision] = useState(0);
  const [toastMsg, setToastMsg] = useState<{ msg: string; type: "error" | "info" } | null>(null);
  const fileInputDefaultRef = useRef<HTMLInputElement>(null);
  const fileInputLyricsRef = useRef<HTMLInputElement>(null);

  const [videoBgList, setVideoBgList] = useState<{ filename: string; url: string }[]>([]);
  const [globalVideoBg, setGlobalVideoBg] = useState("");
  const [globalImageBgDisabled, setGlobalImageBgDisabled] = useState(false);
  const videoBgRef = useRef<HTMLInputElement>(null);

  const showToast = (msg: string, type: "error" | "info" = "error") => {
    setToastMsg({ msg, type });
    setTimeout(() => setToastMsg(null), 3000);
  };

  const fetchFiles = async () => {
    setLoading(true);
    try {
      const [defRes, lyrRes] = await Promise.all([
        apiClient.getTemplates("display-default"),
        apiClient.getTemplates("lyrics"),
      ]);
      setDefaultFiles(defRes.files || []);
      setLyricsFiles(lyrRes.files || []);
      setImgRevision((r) => r + 1);
    } catch {
      setDefaultFiles([]);
      setLyricsFiles([]);
    }
    setLoading(false);
  };

  useEffect(() => {
    fetchFiles();
    apiClient.listVideoBg().then(setVideoBgList).catch(() => {});
    apiClient.getDisplayConfig().then((c) => {
      setGlobalVideoBg(c.globalVideoBg || "");
      setGlobalImageBgDisabled(c.globalImageBgDisabled ?? false);
    }).catch(() => {});
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const saveDisplayBg = async (patch: { globalVideoBg?: string; globalImageBgDisabled?: boolean }) => {
    try {
      const cfg = await apiClient.getDisplayConfig();
      await apiClient.saveDisplayConfig({ ...cfg, ...patch });
      if (patch.globalVideoBg !== undefined) setGlobalVideoBg(patch.globalVideoBg);
      if (patch.globalImageBgDisabled !== undefined) setGlobalImageBgDisabled(patch.globalImageBgDisabled);
    } catch {
      showToast("저장 실패");
    }
  };

  const handleUpload = async (fileList: FileList | null, category: "display-default" | "lyrics") => {
    if (!fileList || fileList.length === 0) return;
    const file = fileList[0];
    const ext = file.name.toLowerCase().split(".").pop();
    if (!["png", "jpg", "jpeg"].includes(ext || "")) {
      showToast("PNG/JPG 파일만 업로드 가능합니다.");
      return;
    }
    try {
      await apiClient.uploadTemplate(file, category);
      fetchFiles();
    } catch {
      showToast("업로드 실패");
    }
  };

  const handleDeleteLyrics = async (name: string) => {
    try {
      await apiClient.deleteTemplate("lyrics", name);
      fetchFiles();
    } catch {
      showToast("삭제 실패");
    }
  };

  const handleVideoBgUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    try {
      const res = await apiClient.uploadVideoBg(file);
      const newList = await apiClient.listVideoBg();
      setVideoBgList(newList);
      await saveDisplayBg({ globalVideoBg: res.filename, globalImageBgDisabled: false });
      showToast("비디오 배경이 업로드되었습니다.", "info");
    } catch {
      showToast("업로드 실패");
    }
    e.target.value = "";
  };

  const handleVideoBgDelete = async (filename: string) => {
    try {
      await apiClient.deleteVideoBg(filename);
      const newList = await apiClient.listVideoBg();
      setVideoBgList(newList);
      if (globalVideoBg === filename) {
        await saveDisplayBg({ globalVideoBg: "", globalImageBgDisabled: false });
      }
      showToast("삭제되었습니다.", "info");
    } catch {
      showToast("삭제 실패");
    }
  };

  // 현재 선택된 배경 판별
  const isNoneActive = globalVideoBg === "" && globalImageBgDisabled;
  const isImageActive = globalVideoBg === "" && !globalImageBgDisabled;

  return (
    <div className="flex flex-col gap-4">
      {/* ── Display 배경 (이미지 + 비디오 통합) ── */}
      <div>
        <div className="flex justify-between items-center mb-1">
          <span className="text-[10px] font-semibold text-[#ccc]">Display 배경</span>
          <div className="flex items-center gap-1.5">
            <button
              className="px-2 py-1 text-[10px] bg-pro-elevated text-pro-text-dim border border-pro-border rounded cursor-pointer hover:opacity-90"
              onClick={() => fileInputDefaultRef.current?.click()}
              title="이미지 배경 업로드 (Frame 2)"
            >
              이미지
            </button>
            <button
              className="px-2 py-1 text-[10px] bg-[#4a9eff] text-white border-none rounded cursor-pointer hover:opacity-90"
              onClick={() => videoBgRef.current?.click()}
              title="비디오 배경 업로드 (MP4/WebM)"
            >
              비디오
            </button>
          </div>
          <input
            ref={fileInputDefaultRef}
            type="file"
            accept="image/png,image/jpeg"
            className="hidden"
            onChange={(e) => handleUpload(e.target.files, "display-default")}
          />
          <input
            ref={videoBgRef}
            type="file"
            accept="video/mp4,video/webm,video/quicktime,.mov"
            className="hidden"
            onChange={handleVideoBgUpload}
          />
        </div>
        <p className="text-[10px] text-[#555] mb-2">
          클릭으로 적용 · Display 화면 및 예배 PDF에 함께 사용 · 비디오 MP4/WebM 최대 300MB
        </p>

        <div className="grid grid-cols-2 gap-2">
          {/* 없음 (검정) 타일 */}
          <div
            onClick={() => saveDisplayBg({ globalVideoBg: "", globalImageBgDisabled: true })}
            className={`relative rounded-lg overflow-hidden cursor-pointer border-2 transition-all flex flex-col items-center justify-center bg-[#0a0a0a] aspect-video ${
              isNoneActive ? "border-[#4a9eff]" : "border-transparent hover:border-white/30"
            }`}
          >
            <span className="text-[16px] mb-0.5">⬛</span>
            <span className="text-[10px] text-[#666]">없음</span>
            {isNoneActive && (
              <div className="absolute top-1 right-1 bg-[#4a9eff] text-white text-[8px] font-bold px-1.5 py-0.5 rounded">
                적용 중
              </div>
            )}
          </div>

          {/* Frame 2 이미지 타일 */}
          {loading ? (
            <div className="rounded-lg bg-[#1a1a1a] aspect-video flex items-center justify-center">
              <span className="text-[10px] text-[#555]">로딩 중...</span>
            </div>
          ) : defaultFiles.length > 0 ? (
            defaultFiles.map((f) => (
              <div
                key={f.name}
                onClick={() => saveDisplayBg({ globalVideoBg: "", globalImageBgDisabled: false })}
                className={`relative rounded-lg overflow-hidden cursor-pointer border-2 transition-all ${
                  isImageActive ? "border-[#4a9eff]" : "border-transparent hover:border-white/30"
                }`}
              >
                <img
                  src={`${baseUrl}${f.url}?v=${imgRevision}`}
                  alt={f.name}
                  className="w-full aspect-video object-cover block bg-[#1a1a1a]"
                />
                <div className="flex items-center px-1.5 py-1 bg-black/60 absolute bottom-0 left-0 right-0">
                  <span className="text-[9px] text-white/80 truncate flex-1">{f.name}</span>
                </div>
                {isImageActive && (
                  <div className="absolute top-1 right-1 bg-[#4a9eff] text-white text-[8px] font-bold px-1.5 py-0.5 rounded">
                    적용 중
                  </div>
                )}
              </div>
            ))
          ) : (
            <div
              onClick={() => fileInputDefaultRef.current?.click()}
              className="rounded-lg border-2 border-dashed border-white/15 bg-white/5 aspect-video flex flex-col items-center justify-center cursor-pointer hover:border-white/30 transition-colors"
            >
              <span className="text-[20px] mb-1">🖼</span>
              <span className="text-[10px] text-[#555]">이미지 없음</span>
            </div>
          )}

          {/* 비디오 타일 */}
          {videoBgList.map((v) => {
            const isActive = globalVideoBg === v.filename;
            return (
              <div
                key={v.filename}
                onClick={() =>
                  saveDisplayBg({
                    globalVideoBg: isActive ? "" : v.filename,
                    globalImageBgDisabled: false,
                  })
                }
                className={`relative rounded-lg overflow-hidden cursor-pointer border-2 transition-all ${
                  isActive ? "border-[#4a9eff]" : "border-transparent hover:border-white/30"
                }`}
              >
                <video
                  src={`${baseUrl}${v.url}`}
                  className="w-full aspect-video object-cover block bg-[#1a1a1a]"
                  muted
                  preload="metadata"
                  onLoadedMetadata={(e) => {
                    (e.target as HTMLVideoElement).currentTime = 1;
                  }}
                />
                <div className="flex items-center justify-between px-1.5 py-1 bg-black/60 absolute bottom-0 left-0 right-0">
                  <span className="text-[9px] text-white/80 truncate flex-1">{v.filename}</span>
                  <button
                    className="text-[9px] text-[#ff6b6b] hover:opacity-80 cursor-pointer ml-1 shrink-0 bg-transparent border-none"
                    onClick={(e) => {
                      e.stopPropagation();
                      handleVideoBgDelete(v.filename);
                    }}
                  >
                    삭제
                  </button>
                </div>
                {isActive && (
                  <div className="absolute top-1 right-1 bg-[#4a9eff] text-white text-[8px] font-bold px-1.5 py-0.5 rounded">
                    적용 중
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>

      <div className="h-px bg-pro-border" />

      {/* ── 가사 배경 (Frame 1) — PDF 전용 ── */}
      <div>
        <div className="flex justify-between items-center mb-1">
          <div className="flex items-center gap-1.5">
            <span className="text-[10px] font-semibold text-[#ccc]">가사 배경</span>
            <span className="text-[9px] font-semibold bg-[#2a1a00] text-[#f59e0b] px-1.5 py-0.5 rounded">
              PDF 전용
            </span>
          </div>
          <button
            className="px-2 py-1 text-[10px] bg-[#4a9eff] text-white border-none rounded cursor-pointer hover:opacity-90"
            onClick={() => fileInputLyricsRef.current?.click()}
          >
            업로드
          </button>
          <input
            ref={fileInputLyricsRef}
            type="file"
            accept="image/png,image/jpeg"
            className="hidden"
            onChange={(e) => handleUpload(e.target.files, "lyrics")}
          />
        </div>
        <p className="text-[10px] text-[#555] mb-2">
          찬양 가사 PDF 생성에만 사용 · Display 화면에는 영향 없음
        </p>
        {loading ? (
          <div className="text-[#888] text-center py-3 text-xs">로딩 중...</div>
        ) : lyricsFiles.length === 0 ? (
          <div className="text-[#555] text-center py-3 text-xs">가사 배경 없음</div>
        ) : (
          <div className="flex flex-col gap-1.5">
            {lyricsFiles.map((f) => (
              <div key={f.name} className="relative rounded overflow-hidden bg-[#1a1a1a]">
                <img
                  src={`${baseUrl}${f.url}?v=${imgRevision}`}
                  alt={f.name}
                  className="w-full object-cover max-h-[120px]"
                />
                <div className="px-2 py-1 flex items-center justify-between">
                  <span className="text-[#aaa] text-[10px] truncate flex-1">{f.name}</span>
                  <button
                    onClick={() => handleDeleteLyrics(f.name)}
                    className="bg-[rgba(255,60,60,0.2)] border-none text-[#ff6b6b] text-[10px] px-1.5 py-0.5 rounded cursor-pointer ml-1 flex-shrink-0 hover:bg-[rgba(255,60,60,0.35)] transition-colors"
                  >
                    삭제
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

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
    </div>
  );
}
