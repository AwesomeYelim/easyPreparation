import { atom } from "recoil";
import worshipData from "@/data/data.json";
import { WorshipOrderItem } from "./bulletin/components/WorshipOrder";

// 예배 순서 상태
export const worshipOrderState = atom({
  key: "worshipOrderState",
  default: worshipData,
});

export const selectedDetailState = atom<WorshipOrderItem | null>({
  key: "selectedDetailState",
  default: null,
});

export const churchNewsState = atom({
  key: "churchNewsState",
  default: worshipData.find((el) => el.title === "교회소식"),
});
