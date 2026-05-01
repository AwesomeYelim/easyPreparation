"use client";

import { useState, useRef, useEffect } from "react";
import { useSetRecoilState } from "recoil";
import { bulletinPreviewState, inspectorOpenState } from "@/recoilState";
import toast from "react-hot-toast";

interface TemplateSelectorProps {
  worshipType: string;
}

// coverAdapt: true인 시안만 표지 이미지 색상 팔레트 적용
const TEMPLATES = [
  { n: 1,  name: "클래식",     desc: "와인 + 아이보리",  bg: "#F5EFE4", accent: "#6B1F2A", coverAdapt: true  },
  { n: 2,  name: "미니멀",     desc: "네이비 + 화이트",  bg: "#FAFAF7", accent: "#0F1B2D", coverAdapt: false },
  { n: 3,  name: "에디토리얼", desc: "크림 + 와인",      bg: "#EFE9DD", accent: "#7A2C2C", coverAdapt: true  },
  { n: 4,  name: "자연/은혜",  desc: "올리브 + 크림",   bg: "#F6F1E3", accent: "#5D6A3A", coverAdapt: true  },
  { n: 5,  name: "청년부",     desc: "오렌지 · 에너지", bg: "#EDE6D7", accent: "#D85B2A", coverAdapt: false },
  { n: 6,  name: "한국 클래식",desc: "명조 · 네이비",   bg: "#F2EBD9", accent: "#1D2939", coverAdapt: false },
  { n: 9,  name: "영수증/노트",desc: "빈티지 · 모노",   bg: "#F5F1E8", accent: "#A83232", coverAdapt: true  },
  { n: 10, name: "사진집",     desc: "무인양품 · 청자", bg: "#F3F1EC", accent: "#3A5A78", coverAdapt: true  },
];

export default function TemplateSelector({ worshipType }: TemplateSelectorProps) {
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState<number | null>(null);
  const [coverUrl, setCoverUrl] = useState<string | null>(null);
  const [coverTs, setCoverTs] = useState(0);
  const [coverUploading, setCoverUploading] = useState(false);
  const [accentColor, setAccentColor] = useState<string | null>(null);
  const [dropPos, setDropPos] = useState<{ top: number; right: number } | null>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const coverInputRef = useRef<HTMLInputElement>(null);
  const setBulletinPreview = useSetRecoilState(bulletinPreviewState);
  const setInspOpen = useSetRecoilState(inspectorOpenState);

  // 현재 표지 이미지 + 주조색 확인
  useEffect(() => {
    fetch("/api/bulletin-cover")
      .then((r) => { if (r.ok) setCoverTs(Date.now()); })
      .catch(() => {});
    fetch("/api/bulletin-theme")
      .then((r) => r.ok ? r.json() : null)
      .then((d) => { if (d?.accentColor) setAccentColor(d.accentColor); })
      .catch(() => {});
  }, []);

  // 외부 클릭 시 닫기
  useEffect(() => {
    if (!open) return;
    const handler = (e: MouseEvent) => {
      if (
        dropdownRef.current && !dropdownRef.current.contains(e.target as Node) &&
        triggerRef.current && !triggerRef.current.contains(e.target as Node)
      ) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [open]);

  const handleToggle = () => {
    if (!open && triggerRef.current) {
      const rect = triggerRef.current.getBoundingClientRect();
      setDropPos({
        top: rect.bottom + 6,
        right: window.innerWidth - rect.right,
      });
    }
    setOpen((v) => !v);
  };

  const handleDownloadPDF = async (n: number) => {
    setLoading(n);
    const toastId = toast.loading(`시안 ${n} PDF 생성 중…`);
    try {
      // Desktop 모드: 서버가 ~/Downloads에 직접 저장
      const saveRes = await fetch(`/api/bulletin-pdf-save?type=${worshipType}&template=${n}`);
      if (saveRes.ok) {
        toast.success("다운로드 폴더에 저장되었습니다.", { id: toastId });
        return;
      }
      if (saveRes.status !== 403) {
        const text = await saveRes.text();
        throw new Error(text || "저장 실패");
      }
      // 웹 브라우저 모드: fetch → blob → a.click()
      const res = await fetch(`/api/bulletin-pdf?type=${worshipType}&template=${n}`);
      if (!res.ok) throw new Error((await res.text()) || "PDF 생성 실패");
      const blob = await res.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `bulletin_${worshipType}_${new Date().toISOString().slice(0, 10)}.pdf`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
      toast.success("PDF 다운로드 완료", { id: toastId });
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "PDF 생성 중 오류", { id: toastId });
    } finally {
      setLoading(null);
    }
  };

  const handlePreview = (n: number) => {
    setBulletinPreview({ template: n, worshipType });
    setInspOpen(true);
    // 드롭다운 유지 — 시안을 보면서 다른 것도 골라볼 수 있도록
  };

  const handleCoverUpload = async (file: File) => {
    setCoverUploading(true);
    try {
      const form = new FormData();
      form.append("file", file);
      const res = await fetch("/api/bulletin-cover", { method: "POST", body: form });
      if (!res.ok) throw new Error("업로드 실패");
      setCoverTs(Date.now());
      toast.success("표지 이미지가 저장되었습니다.");
      // 주조색 갱신 (서버가 업로드 직후 추출함)
      fetch("/api/bulletin-theme")
        .then((r) => r.ok ? r.json() : null)
        .then((d) => { if (d?.accentColor) setAccentColor(d.accentColor); })
        .catch(() => {});
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "업로드 실패");
    } finally {
      setCoverUploading(false);
    }
  };

  return (
    <div className="relative flex-shrink-0">
      {/* 트리거 버튼 */}
      <button
        ref={triggerRef}
        onClick={handleToggle}
        className={`flex items-center gap-1.5 px-4 h-9 rounded-lg font-bold text-sm border transition-all whitespace-nowrap ${
          open
            ? "bg-pro-hover border-electric-blue text-pro-text"
            : "bg-pro-surface text-pro-text border-pro-border hover:bg-pro-hover"
        }`}
        title="주보 시안 선택 및 PDF 다운로드"
      >
        <svg width="13" height="13" viewBox="0 0 16 16" fill="none">
          <rect x="1" y="1" width="6" height="6" rx="1" stroke="currentColor" strokeWidth="1.5"/>
          <rect x="9" y="1" width="6" height="6" rx="1" stroke="currentColor" strokeWidth="1.5"/>
          <rect x="1" y="9" width="6" height="6" rx="1" stroke="currentColor" strokeWidth="1.5"/>
          <rect x="9" y="9" width="6" height="6" rx="1" stroke="currentColor" strokeWidth="1.5"/>
        </svg>
        주보 시안
        <svg
          width="9" height="9" viewBox="0 0 10 10" fill="none"
          className={`transition-transform duration-150 ${open ? "rotate-180" : ""}`}
        >
          <path d="M2 3.5L5 6.5L8 3.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"/>
        </svg>
      </button>

      {/* 드롭다운 패널 — fixed로 overflow 클리핑 탈출 */}
      {open && dropPos && (
        <div
          ref={dropdownRef}
          className="fixed bg-pro-surface border border-pro-border rounded-xl shadow-2xl z-[200] overflow-hidden"
          style={{ width: 400, top: dropPos.top, right: dropPos.right }}
        >

          {/* ── 표지 이미지 ── */}
          <div className="px-3 pt-3 pb-2.5 border-b border-pro-border">
            <div className="flex items-center justify-between mb-1.5">
              <span className="text-[11px] font-bold text-pro-text">표지 이미지</span>
              <span className="text-[9px] text-pro-text-muted">주보 커버에 사용 · 매주 교체 가능</span>
            </div>
            <div
              onClick={() => coverInputRef.current?.click()}
              className="flex items-center gap-2.5 p-2 bg-pro-bg rounded-lg border border-pro-border cursor-pointer hover:border-electric-blue/50 transition-colors group"
            >
              {coverUploading ? (
                <div className="w-9 h-9 flex items-center justify-center flex-shrink-0">
                  <div className="w-4 h-4 border-2 border-white/20 border-t-electric-blue rounded-full animate-spin" />
                </div>
              ) : coverTs > 0 ? (
                <img
                  src={`/api/bulletin-cover?t=${coverTs}`}
                  alt="표지"
                  className="w-9 h-9 object-cover rounded border border-pro-border flex-shrink-0"
                />
              ) : (
                <div className="w-9 h-9 rounded border border-dashed border-pro-border flex items-center justify-center flex-shrink-0">
                  <svg width="12" height="12" viewBox="0 0 16 16" fill="none">
                    <path d="M8 4V12M4 8H12" stroke="#666" strokeWidth="1.5" strokeLinecap="round"/>
                  </svg>
                </div>
              )}
              <div className="flex-1 min-w-0">
                <p className="text-[10px] text-pro-text font-medium">
                  {coverTs > 0 ? "표지 이미지 설정됨" : "표지 이미지 없음"}
                </p>
                <p className="text-[9px] text-pro-text-muted mt-0.5">
                  {coverTs > 0 ? "클릭하여 교체 (PNG/JPG)" : "클릭 또는 드래그 · PNG/JPG"}
                </p>
              </div>
              {coverTs > 0 && (
                <span className="text-[9px] text-electric-blue font-medium flex-shrink-0 opacity-0 group-hover:opacity-100 transition-opacity">교체</span>
              )}
            </div>
            <input
              ref={coverInputRef}
              type="file"
              accept="image/png,image/jpeg"
              className="hidden"
              onChange={(e) => { const f = e.target.files?.[0]; if (f) handleCoverUpload(f); e.target.value = ""; }}
            />
          </div>

          {/* ── 시안 그리드 ── */}
          <div className="p-3">
            <span className="text-[11px] font-bold text-pro-text block mb-2">시안 선택</span>
            <div className="grid grid-cols-4 gap-1.5">
              {TEMPLATES.map((t) => (
                <div
                  key={t.n}
                  className="relative rounded-lg border border-pro-border bg-pro-bg overflow-hidden hover:border-electric-blue transition-all group"
                >
                  {/* 색상 띠 */}
                  <div
                    className="h-1.5 w-full"
                    style={{
                      background: `linear-gradient(to right, ${t.bg} 55%, ${t.coverAdapt && accentColor ? accentColor : t.accent} 100%)`,
                    }}
                  />
                  {/* 텍스트 */}
                  <div className="px-1.5 py-1.5">
                    <div className="flex items-baseline justify-between gap-0.5">
                      <span className="text-[10px] font-bold text-pro-text leading-tight truncate">{t.name}</span>
                      <span className="text-[9px] text-pro-text-muted font-mono flex-shrink-0">{t.n}</span>
                    </div>
                    <p className="text-[8px] text-pro-text-muted leading-tight mt-0.5 truncate">
                      {t.coverAdapt && accentColor ? "이미지 색상 적용" : t.desc}
                    </p>
                  </div>

                  {/* 로딩 오버레이 */}
                  {loading === t.n && (
                    <div className="absolute inset-0 bg-black/70 flex items-center justify-center">
                      <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                    </div>
                  )}

                  {/* 호버 오버레이: 미리보기 + 다운로드 */}
                  {loading !== t.n && (
                    <div className="absolute inset-0 bg-black/80 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
                      {/* 미리보기 */}
                      <button
                        onClick={() => handlePreview(t.n)}
                        className="p-2 rounded-lg bg-white/15 hover:bg-white/30 transition-colors"
                        title="Studio 미리보기"
                      >
                        <svg width="17" height="17" viewBox="0 0 16 16" fill="none">
                          <path d="M2 8C2 8 4.5 3 8 3C11.5 3 14 8 14 8C14 8 11.5 13 8 13C4.5 13 2 8 2 8Z"
                            stroke="white" strokeWidth="1.5" strokeLinejoin="round"/>
                          <circle cx="8" cy="8" r="2" stroke="white" strokeWidth="1.5"/>
                        </svg>
                      </button>
                      {/* 다운로드 */}
                      <button
                        onClick={() => handleDownloadPDF(t.n)}
                        className="p-2 rounded-lg bg-white/15 hover:bg-white/30 transition-colors"
                        title="PDF 다운로드"
                      >
                        <svg width="17" height="17" viewBox="0 0 16 16" fill="none">
                          <path d="M8 2V10M8 10L5 7M8 10L11 7M3 13H13" stroke="white" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
                        </svg>
                      </button>
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
