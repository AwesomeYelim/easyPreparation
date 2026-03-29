import { atom } from "recoil";
import main_worship from "@/data/main_worship.json";
import after_worship from "@/data/after_worship.json";
import wed_worship from "@/data/wed_worship.json";
import { WorshipOrderItem, UserChurchInfo, SongBlock, UserSettings } from "@/types";

// 예배 타입 키만 모아두기
export type WorshipType = "main_worship" | "after_worship" | "wed_worship";

// 예배 순서 상태
export const worshipOrderState = atom<Record<WorshipType, WorshipOrderItem[]>>({
  key: "worshipOrderState",
  default: {
    main_worship,
    after_worship,
    wed_worship,
  },
});

export const selectedDetailState = atom<WorshipOrderItem>({
  key: "selectedDetailState",
  default: main_worship[0],
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

// Display에 전송된 항목 목록
export const displayItemsState = atom<WorshipOrderItem[]>({
  key: "displayItemsState",
  default: [],
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
    figmaInfo: {
      key: "",
      token: "",
    },
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
