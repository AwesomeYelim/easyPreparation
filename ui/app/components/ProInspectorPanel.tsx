"use client";

import { useState, useEffect } from "react";
import { useRecoilState, useRecoilValue } from "recoil";
import {
  inspectorOpenState,
  bulletinPreviewState,
  inspectorTabState,
  scheduleActiveState,
  InspectorCategory,
} from "@/recoilState";
import FeatureGate from "./FeatureGate";
import OBSSourcePanel from "./OBSSourcePanel";
import InspectorPreviewTab from "./inspector/InspectorPreviewTab";
import InspectorDisplayTab from "./inspector/InspectorDisplayTab";
import InspectorBackgroundsTab from "./inspector/InspectorBackgroundsTab";
import InspectorScheduleTab from "./inspector/InspectorScheduleTab";
import InspectorSpecialTab from "./inspector/InspectorSpecialTab";
import InspectorConfigTab from "./inspector/InspectorConfigTab";
import InspectorYoutubeTab from "./inspector/InspectorYoutubeTab";

const TABS: { key: InspectorCategory; label: string; desc: string }[] = [
  { key: "preview",     label: "미리보기", desc: "Display 화면 실시간 미리보기" },
  { key: "display",     label: "화면",    desc: "항목별 배경 이미지" },
  { key: "backgrounds", label: "배경",    desc: "" },
  { key: "special",     label: "썸네일",  desc: "기념주일 썸네일 배경" },
  { key: "youtube",     label: "유튜브",  desc: "방송 제목 · 설명 템플릿" },
  { key: "schedule",    label: "스케줄",  desc: "정기 스트리밍 스케줄" },
  { key: "obs",         label: "OBS",     desc: "OBS 소스 및 씬 관리" },
  { key: "config",      label: "스타일",  desc: "폰트 · 로고 · 오버레이" },
];

const BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ||
  (typeof window !== "undefined" ? window.location.origin : "http://localhost:8080");

export default function ProInspectorPanel() {
  const [inspOpen, setInspOpen] = useRecoilState(inspectorOpenState);
  const [tab, setTab] = useRecoilState(inspectorTabState);
  const [bulletinPreview, setBulletinPreview] = useRecoilState(bulletinPreviewState);
  const [toastMsg, setToastMsg] = useState<{ msg: string; type: "error" | "info" } | null>(null);
  const scheduleActive = useRecoilValue(scheduleActiveState);

  // 주보 미리보기 설정 시 자동으로 preview 탭으로 전환
  useEffect(() => {
    if (bulletinPreview.template !== null) {
      setTab("preview");
    }
  }, [bulletinPreview.template]); // eslint-disable-line react-hooks/exhaustive-deps

  const showToast = (msg: string, type: "error" | "info" = "error") => {
    setToastMsg({ msg, type });
    setTimeout(() => setToastMsg(null), 3000);
  };

  if (!inspOpen) return null;

  return (
    <div
      className="flex flex-col bg-pro-surface border-l border-pro-border overflow-hidden"
      style={{ gridColumn: "4", gridRow: "2" }}
    >
      {/* header */}
      <div className="flex justify-between items-center px-3 py-2.5 border-b border-pro-border flex-shrink-0">
        <span className="text-[12px] font-semibold text-pro-text">Studio</span>
        <button
          onClick={() => setInspOpen(false)}
          className="bg-transparent border-none text-pro-text-dim text-base cursor-pointer leading-none hover:text-pro-text transition-colors w-9 h-9 flex items-center justify-center"
        >
          ✕
        </button>
      </div>

      {/* tabs */}
      <div className="flex flex-wrap border-b border-pro-border px-1 flex-shrink-0 overflow-x-hidden">
        {TABS.map((t) => (
          <button
            key={t.key}
            onClick={() => setTab(t.key)}
            className={`px-2 py-1.5 bg-transparent border-none border-b-2 text-[10px] cursor-pointer transition-colors whitespace-nowrap ${
              tab === t.key
                ? "border-[#4a9eff] text-[#4a9eff] font-semibold"
                : "border-transparent text-pro-text-dim font-normal hover:text-pro-text"
            }`}
          >
            {t.label}
            {t.key === "schedule" && scheduleActive && (
              <span className="inline-block w-1.5 h-1.5 rounded-full bg-amber-400 ml-1 align-middle" />
            )}
          </button>
        ))}
      </div>

      {/* body */}
      <div className={`flex-1 overflow-y-auto overflow-x-hidden ${tab === "obs" ? "" : "px-3 py-3"}`}>
        {tab !== "obs" && !!TABS.find((t) => t.key === tab)?.desc && (
          <div className="text-[10px] text-[#888] mb-2">
            {TABS.find((t) => t.key === tab)?.desc}
          </div>
        )}

        {tab === "preview" && (
          <InspectorPreviewTab
            bulletinPreview={bulletinPreview}
            setBulletinPreview={setBulletinPreview}
          />
        )}
        {tab === "display" && <InspectorDisplayTab baseUrl={BASE_URL} />}
        {tab === "backgrounds" && <InspectorBackgroundsTab baseUrl={BASE_URL} />}
        {tab === "special" && (
          <FeatureGate feature="thumbnail">
            <InspectorSpecialTab />
          </FeatureGate>
        )}
        {tab === "youtube" && <InspectorYoutubeTab />}
        {tab === "schedule" && <InspectorScheduleTab />}
        {tab === "obs" && (
          <OBSSourcePanel inline open onClose={() => setTab("display")} />
        )}
        {tab === "config" && <InspectorConfigTab showToast={showToast} />}
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
