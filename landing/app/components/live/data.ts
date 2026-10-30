// Anonymized worship sequence shared by every live mock on the landing page.
// Mirrors the order captured in /public/shots (names kept as OO).

export type Kind = "hymn" | "scripture" | "prayer" | "text";

export type Item = {
  title: string;
  detail: string;
  who: string;
  kind: Kind;
  /** Slide body for scripture items (개역한글). */
  verse?: string;
  /** Real hymn score capture, when we have one. */
  img?: string;
};

export const sequence: Item[] = [
  { title: "전주", detail: "전주", who: "반주자", kind: "text" },
  {
    title: "예배의 부름",
    detail: "이사야 55:7",
    who: "인도자",
    kind: "scripture",
    verse: "악인은 그 길을, 불의한 자는 그 생각을 버리고 여호와께로 돌아오라",
  },
  { title: "개회기도", detail: "-", who: "인도자", kind: "prayer" },
  { title: "찬송", detail: "28장", who: "인도자", kind: "hymn" },
  { title: "성시교독", detail: "14. 시편 24편", who: "일어서서", kind: "text" },
  { title: "신앙고백", detail: "사도신경", who: "일어서서", kind: "text" },
  { title: "찬송", detail: "96장", who: "함께", kind: "hymn" },
  { title: "대표기도", detail: "-", who: "김OO 집사", kind: "prayer" },
  {
    title: "성경봉독",
    detail: "요한복음 1:9-14",
    who: "일어서서",
    kind: "scripture",
    verse: "참 빛 곧 세상에 와서 각 사람에게 비취는 빛이 있었나니",
  },
  { title: "찬양", detail: "-", who: "OO성가대", kind: "text" },
  { title: "말씀", detail: "예수님을 믿으면", who: "박OO 목사", kind: "text" },
  { title: "헌금봉헌", detail: "287장", who: "함께", kind: "hymn" },
  { title: "교회소식", detail: "-", who: "인도자", kind: "text" },
  {
    title: "찬송",
    detail: "524장",
    who: "일어서서",
    kind: "hymn",
    img: "/shots/display.jpg",
  },
  { title: "축도", detail: "-", who: "박OO 목사", kind: "prayer" },
];

/** Chips offered in the bulletin editor's "예배 순서 선택하기" box. */
export const palette = [
  "행사", "전주", "예배의 부름", "개회기도", "성시교독", "신앙고백", "찬송",
  "대표기도", "성경봉독", "찬양", "참회의 기도", "말씀", "기도", "헌금봉헌",
  "봉헌기도", "교회소식", "축도", "주기도문",
];
