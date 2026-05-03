"use client";

import { useState, useEffect, useRef } from "react";

// Display iframe은 Next.js(:3000)를 거치면 WebSocket 프록시가 안 되므로
// 개발 환경에서는 Go 서버(:8080)로 직접 연결
const DISPLAY_ORIGIN =
  typeof window !== "undefined" && window.location.port === "3000"
    ? "http://localhost:8080"
    : "";

interface Props {
  bulletinPreview: { template: number | null; worshipType: string };
  setBulletinPreview: (v: { template: number | null; worshipType: string }) => void;
}

export default function InspectorPreviewTab({ bulletinPreview, setBulletinPreview }: Props) {
  const [iframeKey, setIframeKey] = useState(0);
  const [previewScale, setPreviewScale] = useState(0.133);
  const previewContainerRef = useRef<HTMLDivElement>(null);

  // 컨테이너 너비 감지 → iframe scale 계산
  useEffect(() => {
    const el = previewContainerRef.current;
    if (!el) return;
    const obs = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const w = entry.contentRect.width;
        if (w > 0) setPreviewScale(w / 1920);
      }
    });
    obs.observe(el);
    // 초기값
    const w = el.clientWidth;
    if (w > 0) setPreviewScale(w / 1920);
    return () => obs.disconnect();
  }, []);

  return (
    <div className="flex flex-col gap-3">
      {/* 주보 시안 미리보기 헤더 */}
      {bulletinPreview.template !== null && (
        <div className="flex items-center justify-between -mb-1">
          <span className="text-[10px] font-bold text-pro-text">
            시안 {bulletinPreview.template} 주보 미리보기
          </span>
          <button
            onClick={() => setBulletinPreview({ template: null, worshipType: "main_worship" })}
            className="text-[9px] text-pro-text-muted hover:text-electric-blue transition-colors px-1.5 py-0.5 rounded hover:bg-pro-hover"
          >
            ← Display 미리보기
          </button>
        </div>
      )}

      {/* iframe 미리보기 */}
      <div
        ref={previewContainerRef}
        className="relative bg-black w-full rounded-lg overflow-hidden"
        style={
          bulletinPreview.template !== null
            ? { paddingBottom: `${(1700 / 1200) * 100}%` }
            : { paddingBottom: "56.25%" }
        }
      >
        {bulletinPreview.template !== null ? (
          <iframe
            key={`bulletin-${bulletinPreview.worshipType}-${bulletinPreview.template}-${iframeKey}`}
            src={`/display/bulletin-print?type=${bulletinPreview.worshipType}&template=${bulletinPreview.template}`}
            scrolling="no"
            title="주보 시안 미리보기"
            style={{
              position: "absolute",
              top: 0,
              left: 0,
              width: "1200px",
              height: "1700px",
              transform: `scale(${previewScale * (1920 / 1200)})`,
              transformOrigin: "top left",
              border: "none",
              pointerEvents: "none",
            }}
          />
        ) : (
          <iframe
            key={`display-live-${iframeKey}`}
            src={`${DISPLAY_ORIGIN}/display`}
            scrolling="no"
            title="씬 미리보기"
            style={{
              position: "absolute",
              top: 0,
              left: 0,
              width: "1920px",
              height: "1080px",
              transform: `scale(${previewScale})`,
              transformOrigin: "top left",
              border: "none",
              pointerEvents: "none",
            }}
          />
        )}
      </div>

      {/* 새로고침 + 새 탭 열기 */}
      <div className="flex justify-end gap-2">
        {bulletinPreview.template !== null && (
          <button
            onClick={async () => {
              const url = `/display/bulletin-print?type=${bulletinPreview.worshipType}&template=${bulletinPreview.template}`;
              // Desktop 모드: 서버가 시스템 브라우저로 열어줌
              const res = await fetch(
                `/api/bulletin-preview?type=${bulletinPreview.worshipType}&template=${bulletinPreview.template}`
              ).catch(() => null);
              if (res && res.ok) return;
              // 웹 브라우저 모드 fallback
              window.open(url, "_blank");
            }}
            className="bg-pro-surface hover:bg-pro-hover text-electric-blue text-xs px-3 py-1 rounded border border-pro-border transition-colors cursor-pointer flex items-center gap-1"
            title="전체 크기로 보기"
          >
            <svg width="10" height="10" viewBox="0 0 16 16" fill="none">
              <path
                d="M7 3H3V13H13V9M9 3H13V7M13 3L8 8"
                stroke="currentColor"
                strokeWidth="1.5"
                strokeLinecap="round"
                strokeLinejoin="round"
              />
            </svg>
            확대 보기
          </button>
        )}
        <button
          onClick={() => setIframeKey((k) => k + 1)}
          className="bg-pro-surface hover:bg-pro-hover text-pro-text text-xs px-3 py-1 rounded border border-pro-border transition-colors cursor-pointer"
        >
          새로고침
        </button>
      </div>

      {bulletinPreview.template === null && (
        <p className="text-[9px] text-[#555] leading-relaxed">
          씬 클릭 시 해당 씬이 Display에 어떻게 보일지 미리봅니다. 순서를 변경해도 미리보기가 따라갑니다.
        </p>
      )}
    </div>
  );
}
