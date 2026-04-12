const BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

export type BibleBookInfo = {
  index: number;
  eng: string;
  kor: string;
  chapters: number[];
};

export type BibleData = Record<string, BibleBookInfo>;

let _cache: BibleData | null = null;

export async function fetchBibleData(): Promise<BibleData> {
  if (_cache) return _cache;
  const res = await fetch(`${BASE_URL}/api/bible/books`, { cache: "force-cache" });
  _cache = await res.json();
  return _cache!;
}

type Selection = {
  book: string;
  chapter: number;
  verse: number;
};

export const formatBibleRanges = (multiSelection: Selection[][], bibleData: BibleData): string =>
  multiSelection
    .map((ranges) => {
      const first = ranges[0];
      const last = ranges[1] || first;
      return (
        `${first.book}_${bibleData[first.book]?.index}/${first.chapter}:${first.verse}` +
        (ranges.length > 1 ? `-${last.chapter}:${last.verse}` : "")
      );
    })
    .join(", ");

export const formatBibleReference = (obj: string): string => {
  if (!obj) return "";
  const bibleRegex = /^(.+?)_\d+\/(\d+):(\d+)(?:-(\d+):)?(\d+)?$/;
  return obj
    .split(",")
    .map((item) => {
      const trimmed = item.trim();
      const match = trimmed.match(bibleRegex);
      if (!match) return trimmed;
      const [, bookName, chapterStart, verseStart, chapterEnd, verseEnd] = match;
      if (!chapterEnd && !verseEnd) return `${bookName} ${chapterStart}:${verseStart}`;
      if (!chapterEnd || chapterStart === chapterEnd) return `${bookName} ${chapterStart}:${verseStart}-${verseEnd}`;
      return `${bookName} ${chapterStart}:${verseStart}-${chapterEnd}:${verseEnd}`;
    })
    .join(", ");
};
