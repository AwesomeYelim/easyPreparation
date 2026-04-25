import { atom } from "recoil";
import { WorshipOrderItem, UserChurchInfo, SongBlock, UserSettings, LicenseStatus } from "@/types";

// 예배 타입 키만 모아두기
export type WorshipType = "main_worship" | "after_worship" | "wed_worship" | "fri_worship";

// localStorage 영속화 effect (atoms 선언 전에 정의)
// setTimeout(0)으로 React 하이드레이션 완료 후에 setSelf를 실행해 서버/클라이언트 HTML 불일치(hydration error) 방지
const localStorageEffect = <T>(key: string) => ({ setSelf, onSet }: any) => {
  if (typeof window === "undefined") return;
  const timer = setTimeout(() => {
    const savedValue = localStorage.getItem(key);
    if (savedValue != null) {
      try { setSelf(JSON.parse(savedValue) as T); } catch {}
    }
  }, 0);
  onSet((newValue: T, _: unknown, isReset: boolean) => {
    if (isReset) localStorage.removeItem(key);
    else localStorage.setItem(key, JSON.stringify(newValue));
  });
  return () => clearTimeout(timer);
};

// 예배 순서 상태 (API에서 로드)
export const worshipOrderState = atom<Record<WorshipType, WorshipOrderItem[]>>({
  key: "worshipOrderState",
  default: {
    main_worship: [],
    after_worship: [],
    wed_worship: [],
    fri_worship: [],
  },
});

export const selectedDetailState = atom<WorshipOrderItem>({
  key: "selectedDetailState",
  default: { key: "", title: "", obj: "", info: "-", lead: "" },
});

// 가사 곡 목록 (페이지 이동해도 유지)
export const lyricsSongsState = atom<SongBlock[]>({
  key: "lyricsSongsState",
  default: [],
});

// Display 제어판 열림 상태
export const displayPanelOpenState = atom<boolean>({
  key: "displayPanelOpenState",
  default: false,
});

// Display에 전송된 항목 목록 (localStorage 영속화 — 페이지 새로고침 후 복원)
export const displayItemsState = atom<WorshipOrderItem[]>({
  key: "displayItemsState",
  default: [],
  effects: [localStorageEffect<WorshipOrderItem[]>("ep_display_items")],
});

export const userInfoState = atom<UserChurchInfo>({
  key: "userInfoState",
  default: {
    id: 0,
    name: "",
    english_name: "",
    title: "",
    content: "",
    email: "",
  },
});

export const userSettingsState = atom<UserSettings>({
  key: "userSettingsState",
  default: {
    preferred_bible_version: 1,
    theme: "light",
    font_size: 16,
    default_bpm: 100,
    display_layout: "default",
  },
});

export const licenseState = atom<LicenseStatus>({
  key: "licenseState",
  default: {
    plan: "free",
    features: [],
    expires_at: null,
    days_remaining: 0,
    device_id: "",
    grace_period: false,
    is_active: false,
  },
});

export const sidebarCollapsedState = atom<boolean>({
  key: "sidebarCollapsedState",
  default: false,
});

// ── ProShell 전용 UI 상태 ──

// Sequence 패널(좌측 예배순서 패널) 열림/닫힘
export const sequencePanelOpenState = atom<boolean>({
  key: "sequencePanelOpenState",
  default: true,
});

// Inspector 패널(우측 정보 패널) 열림/닫힘
export const inspectorOpenState = atom<boolean>({
  key: "inspectorOpenState",
  default: true,
});

// Inspector 현재 탭
export type InspectorTab = "cue" | "background" | "obs" | "pdf";
export const inspectorTabState = atom<InspectorTab>({
  key: "inspectorTabState",
  default: "cue",
});

// 서비스 시작 시각 (하단 타임라인 경과 시간 표시용, Phase 4에서 사용)
export const serviceStartTimeState = atom<number | null>({
  key: "serviceStartTimeState",
  default: null,
});

// 아이템별 자동 타이머 (초 단위, key → seconds)
export const itemTimersState = atom<Record<string, number>>({
  key: "itemTimersState",
  default: {},
  effects: [localStorageEffect<Record<string, number>>("ep_item_timers")],
});

// 자동 진행 ON/OFF
export const autoAdvanceState = atom<boolean>({
  key: "autoAdvanceState",
  default: false,
});

// 현재 display 위치 — ProTimeline + ProSequencePanel 공유
// 낙관적 업데이트(자동진행) + WS position 메시지 모두 이 atom을 통해 동기화
export const displayPositionState = atom<number>({
  key: "displayPositionState",
  default: 0,
});
