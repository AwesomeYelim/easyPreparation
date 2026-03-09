import bibleData from "@/data/bible_info.json";

type BibleKey = keyof typeof bibleData;

type Selection = {
  book: string;
  chapter: number;
  verse: number;
};

export const formatBibleRanges = (multiSelection: Selection[][]): string =>
  multiSelection
    .map((ranges) => {
      const first = ranges[0];
      const last = ranges[1] || first;
      return (
        `${first.book}_${bibleData[first.book as BibleKey]?.index}/${first.chapter}:${first.verse}` +
        (ranges.length > 1 ? `-${last.chapter}:${last.verse}` : "")
      );
    })
    .join(", ");

export const formatBibleReference = (obj: string): string => {
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
