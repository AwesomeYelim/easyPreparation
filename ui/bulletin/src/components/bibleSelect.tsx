import React, { useState } from "react";
import bibleData from "../bible_info.json"; // JSON 파일 import
import "./bibleSelect.css"; // CSS 파일 import

type Selection = {
  book: string;
  chapter: number;
  verse: number;
};

interface BibleSelectProps {
  handleValueChange: (key: string, newObj: string) => void;
  parentKey: string;
}
const BibleSelect: React.FC<BibleSelectProps> = ({ handleValueChange, parentKey }) => {
  const [selectedBook, setSelectedBook] = useState<Selection>({
    book: "",
    chapter: 0,
    verse: 0,
  });
  const [selectedRanges, setSelectedRanges] = useState<[Selection, Selection]>([]);

  const handler = {
    bookChange: (event: React.ChangeEvent<HTMLSelectElement>) => {
      const bookKey = event.target.value;
      setSelectedBook({ book: bookKey, chapter: 0, verse: 0 });
    },
    chapterChange: (event: React.ChangeEvent<HTMLSelectElement>) => {
      setSelectedBook({
        ...selectedBook,
        chapter: Number(event.target.value),
        verse: 0,
      });
    },
    verseChange: (event: React.ChangeEvent<HTMLSelectElement>) => {
      setSelectedBook({ ...selectedBook, verse: Number(event.target.value) });
    },
    addSelection: () => {
      if (selectedBook.book && selectedBook.chapter > 0 && selectedBook.verse > 0) {
        setSelectedRanges((prevRanges) => {
          const updatedRanges = [...prevRanges, selectedBook];
          const first = updatedRanges[0];
          const last = updatedRanges[1];

          if (updatedRanges.length > 1) {
            handleValueChange(
              parentKey,
              `${bibleData[selectedBook.book].kor}_${bibleData[selectedBook.book].eng}/${first.chapter}:${
                first.verse
              }-${last.chapter}:${last.verse}`
            );
          } else {
            handleValueChange(
              parentKey,
              `${bibleData[selectedBook.book].kor}_${bibleData[selectedBook.book].eng}/${first.chapter}:${first.verse}`
            );
          }

          return updatedRanges;
        });
        setSelectedBook({ book: selectedBook.book, chapter: 0, verse: 0 }); // 초기화
      }
    },
  };

  const currentBook = selectedBook.book ? bibleData[selectedBook.book] : null;
  const currentChapterVerses = currentBook && selectedBook.chapter ? currentBook.chapters[selectedBook.chapter - 1] : 0;

  const formatRange = (ranges: Selection[]) => {
    return ranges
      .map((range, i) => {
        const bookName = bibleData[range.book]?.kor;
        if (!i) {
          return `${bookName} ${range.chapter}장 ${range.verse}절`;
        } else {
          return `${range.chapter}장 ${range.verse}절`;
        }
      })
      .join(" ~ ");
  };

  return (
    <div className="bible-select-container">
      <h3 className="title">성경 구절 선택</h3>
      {selectedRanges.length !== 2 && (
        <>
          <div className="select-group">
            <label className="select-label">
              책 선택:
              <select
                className="select-box"
                onChange={handler.bookChange}
                value={selectedRanges.length > 0 ? selectedRanges[0].book : selectedBook.book || ""}>
                <option value="" disabled>
                  책을 선택하세요
                </option>
                {selectedRanges.length > 0 ? (
                  <option value={bibleData[selectedRanges[0]?.book]}>{bibleData[selectedRanges[0]?.book].kor}</option>
                ) : (
                  Object.entries(bibleData).map(([key, value]) => (
                    <option key={key} value={key}>
                      {value.kor}
                    </option>
                  ))
                )}
              </select>
            </label>

            {currentBook && (
              <label className="select-label">
                장 선택:
                <select className="select-box" onChange={handler.chapterChange} value={selectedBook.chapter || ""}>
                  <option value="" disabled>
                    장을 선택하세요
                  </option>
                  {currentBook.chapters.map((_, index) => (
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
                  <option value="" disabled>
                    절을 선택하세요
                  </option>
                  {Array.from({ length: currentChapterVerses }, (_, i) => i + 1).map((verse, index) => (
                    <option
                      key={verse}
                      value={verse}
                      disabled={
                        selectedRanges[0]?.chapter == selectedBook.chapter ? selectedRanges[0]?.verse > index : false
                      }>
                      {verse}절
                    </option>
                  ))}
                </select>
              </label>
            )}
          </div>
          <button
            className="add-selection-button"
            onClick={handler.addSelection}
            disabled={!(selectedBook.book && selectedBook.chapter > 0 && selectedBook.verse > 0)}>
            추가
          </button>
        </>
      )}
      {selectedRanges.length > 0 && (
        <button
          className="add-selection-button"
          onClick={() => {
            setSelectedRanges([]);
            setSelectedBook({ book: "", chapter: 0, verse: 0 });
          }}>
          다시 선택
        </button>
      )}
      {selectedRanges.length > 0 && (
        <div className="result-container">
          <p className="result-text">{formatRange(selectedRanges)}</p>
        </div>
      )}
    </div>
  );
};

export default BibleSelect;
