import { atom } from "recoil";

// 예배 순서 상태
export const worshipOrderState = atom({
  key: "worshipOrderState",
  default: [
    { title: "1_전주", obj: "전주", info: "-", lead: "반주자" },
    {
      title: "2_예배의 부름",
      obj: "렘 33:2-3",
      info: "b_edit",
      lead: "인도자",
    },
    { title: "3_개회기도", obj: "-", info: "-", lead: "인도자" },
    { title: "4_찬송", obj: "29장", info: "c_edit", lead: "인도자" },
    {
      title: "5_성시교독",
      obj: "125. 사순절(2)",
      info: "c_edit",
      lead: "일어서서",
    },
    { title: "6_신앙고백", obj: "사도신경", info: "-", lead: "일어서서" },
  ],
});

// 교회 소식 상태
export const churchNewsState = atom({
  key: "churchNewsState",
  default: [
    { title: "예배 참여 안내", obj: "매주 금요일 기도 모임", info: "c_edit" },
    {
      title: "교회 절기 및 행사",
      obj: "-",
      info: "c_edit",
      children: [
        { title: "사순절 첫째 주일", obj: "3/9(주일)", info: "c_edit" },
      ],
    },
  ],
});
