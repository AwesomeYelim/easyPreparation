import { atom } from "recoil";
import main_worship from "@/data/main_worship.json";
import after_worship from "@/data/after_worship.json";
import wed_worship from "@/data/wed_worship.json";
import { WorshipOrderItem } from "./bulletin/page";
import { UserChurchInfo } from "./components/SideBar";

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
