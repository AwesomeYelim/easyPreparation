import { atom } from "recoil";
import worshipData from "@/data/data.json";
import { WorshipOrderItem } from "./bulletin/page";
import { UserChurchInfo } from "./components/SideBar";

// 예배 순서 상태
export const worshipOrderState = atom({
  key: "worshipOrderState",
  default: worshipData,
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
