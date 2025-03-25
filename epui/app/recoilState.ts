import { atom } from "recoil";
import worshipData from "@/data/data.json";
import { WorshipOrderItem } from "./bulletin/page";

// 예배 순서 상태
export const worshipOrderState = atom({
  key: "worshipOrderState",
  default: worshipData,
});

export const selectedDetailState = atom<WorshipOrderItem>({
  key: "selectedDetailState",
  default: worshipData[0],
});
