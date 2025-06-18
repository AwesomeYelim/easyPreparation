import { atom } from "recoil";
import worshipData from "@/data/data.json";
import { WorshipOrderItem } from "./bulletin/page";
import { UserChurchInfo } from "./components/SideBar";

// 예배 타입 키만 모아두기
export type WorshipType = "main_worship" | "after_worship" | "wed_worship";

// 예배 순서 상태
export const worshipOrderState = atom<Record<WorshipType, WorshipOrderItem[]>>({
  key: "worshipOrderState",
  default: {
    main_worship: worshipData,
    after_worship: worshipData,
    wed_worship: worshipData,
  },
});

export const selectedDetailState = atom<WorshipOrderItem>({
  key: "selectedDetailState",
  default: worshipData[0],
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
