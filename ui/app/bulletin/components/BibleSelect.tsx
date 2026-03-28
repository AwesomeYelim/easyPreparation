import React, { useState, useEffect } from "react";
import { fetchBibleData, formatBibleRanges, BibleData } from "@/lib/bibleUtils";
import { useRecoilValue } from "recoil";
import { selectedDetailState } from "@/recoilState";
import "./BibleSelect.css";

type Selection = {
  book: string;
  chapter: number;
  verse: number;
};

interface BibleSelectProps {
  handleValueChange: (key: string, { newObj, newLead }: { newObj: string; newLead?: string }) => void;
  parentKey: string;
}

const BibleSelect: React.FC<BibleSelectProps> = ({ handleValueChange, parentKey }) => {
  const selectedDetail = useRecoilValue(selectedDetailState);
  const [bibleData, setBibleData] = useState<BibleData>({});

  useEffect(() => {
    fetchBibleData().then(setBibleData);
  }, []);

  const selectedInitInfo = (() => {
    let sermonsSelection: Selection[][] = [];
    if (selectedDetail.obj) {
      const sermons = selectedDetail.obj.split(/,\s*/);
      sermonsSelection = sermons.map((sEl) => {
        const splitStandardUnder = sEl.split("_");
        const book = splitStandardUnder[0];
        const chapter = splitStandardUnder[1]?.split("/")[1];
        if (!chapter) return [];
        if (chapter.includes("-")) {
          return chapter.split("-").map((el) => {
            const [c, v] = el.split(":");
            return { book, chapter: +c, verse: +v };
          });
        } else {
          const [c, v] = chapter.split(":");
          return [{ book, chapter: +c, verse: +v }];
        }
      });
    }
    return sermonsSelection;
  })();

  const [selectedBook, setSelectedBook] = useState<Selection>({ book: "", chapter: 0, verse: 0 });
  const [selectedRanges, setSelectedRanges] = useState<Selection[]>([]);
  const [multiSelection, setMultiSelection] = useState<Selection[][]>(selectedInitInfo);

  const books = Object.keys(bibleData);
  const currentBook = selectedBook.book ? bibleData[selectedBook.book] : null;
  const currentChapterVerses = currentBook && selectedBook.chapter ? currentBook.chapters[selectedBook.chapter - 1] : 0;

  const handler = {
    bookChange: (e: React.ChangeEvent<HTMLSelectElement>) =>
      setSelectedBook({ book: e.target.value, chapter: 0, verse: 0 }),
    chapterChange: (e: React.ChangeEvent<HTMLSelectElement>) =>
      setSelectedBook((prev) => ({ ...prev, chapter: Number(e.target.value), verse: 0 })),
    verseChange: (e: React.ChangeEvent<HTMLSelectElement>) =>
      setSelectedBook((prev) => ({ ...prev, verse: Number(e.target.value) })),
    addSelection: () => {
      if (selectedBook.book && selectedBook.chapter > 0 && selectedBook.verse > 0) {
        setSelectedRanges((prev) => [...prev, selectedBook]);
        setSelectedBook({ book: selectedBook.book, chapter: 0, verse: 0 });
      }
    },
    finalizeSelection: () => {
      if (selectedRanges.length === 0) return;
      setMultiSelection((prev) => {
        const updated = [...prev, selectedRanges];
        handleValueChange(parentKey, { newObj: formatBibleRanges(updated, bibleData) });
        return updated;
      });
      setSelectedRanges([]);
      setSelectedBook({ book: "", chapter: 0, verse: 0 });
    },
    deleteSelection: (i: number) => {
      setMultiSelection((prev) => {
        const updated = prev.filter((_, idx) => idx !== i);
        handleValueChange(parentKey, { newObj: formatBibleRanges(updated, bibleData) });
        return updated;
      });
    },
  };

  const formatRange = (ranges: Selection[]) =>
    ranges.map((r, i) => (i === 0 ? `${r.book} ${r.chapter}장 ${r.verse}절` : `${r.chapter}장 ${r.verse}절`)).join(" ~ ");

  if (books.length === 0) return <div>성경 데이터 로딩 중...</div>;

  return (
    <>
      <div className="bible-select-container">
        <h3 className="title">성경 구절 선택</h3>
        {selectedRanges.length !== 2 && (
          <>
            <div className="select-group">
              <label className="select-label">
                책 선택:
                <select className="select-box" onChange={handler.bookChange} value={selectedBook.book || ""}>
                  <option value="" disabled>책을 선택하세요</option>
                  {books.map((key) => (
                    <option key={key} value={key}>{key}</option>
                  ))}
                </select>
              </label>
              {currentBook && (
                <label className="select-label">
                  장 선택:
                  <select className="select-box" onChange={handler.chapterChange} value={selectedBook.chapter || ""}>
                    <option value="" disabled>장을 선택하세요</option>
                    {currentBook.chapters.map((_: number, index: number) => (
                      <option key={index} value={index + 1} disabled={selectedRanges[0]?.chapter > index + 1}>
                        {index + 1}장
                      </option>
                    ))}
                  </select>
                </label>
              )}
              {currentBook && selectedBook.chapter > 0 && (
                <label className="select-label">
                  절 선택:
                  <select className="select-box" onChange={handler.verseChange} value={selectedBook.verse || ""}>
                    <option value="" disabled>절을 선택하세요</option>
                    {Array.from({ length: currentChapterVerses }, (_, i) => i + 1).map((verse, index) => (
                      <option key={verse} value={verse}
                        disabled={selectedRanges[0]?.chapter === selectedBook.chapter ? selectedRanges[0]?.verse > index : false}>
                        {verse}절
                      </option>
                    ))}
                  </select>
                </label>
              )}
            </div>
            <button className="add-selection-button" onClick={handler.addSelection}
              disabled={!(selectedBook.book && selectedBook.chapter > 0 && selectedBook.verse > 0)}>
              추가
            </button>
          </>
        )}
        {selectedRanges.length > 0 && (
          <button className="add-selection-button" onClick={() => {
            setSelectedRanges([]);
            setSelectedBook({ book: "", chapter: 0, verse: 0 });
          }}>다시 선택</button>
        )}
        {selectedRanges.length > 0 && (
          <div className="result-container">
            <p className="result-text">{formatRange(selectedRanges)}</p>
          </div>
        )}
      </div>
      <button className="add-selection-button" onClick={handler.finalizeSelection}>구절 추가</button>
      {multiSelection.length > 0 && (
        <div className="multi-selection-list">
          {multiSelection.map((ranges, index) => {
            const first = ranges[0];
            const last = ranges[1] || first;
            const displayText =
              `${first.book} ${first.chapter}:${first.verse}` +
              (ranges.length > 1 ? `-${first.chapter === last.chapter ? "" : `${last.chapter}:`}${last.verse}` : "");
            return (
              <span key={index} className="verse-chip">
                📖 {displayText}
                <button className="delete-button" onClick={() => handler.deleteSelection(index)}>x</button>
              </span>
            );
          })}
        </div>
      )}
    </>
  );
};

export default BibleSelect;
