import React, { useState } from "react";
import bibleData from "@/data/bible_info.json";
import { useRecoilValue } from "recoil";
import { selectedDetailState } from "@/recoilState";
import "./BibleSelect.css";

type Selection = {
  book: string;
  chapter: number;
  verse: number;
};

interface BibleSelectProps {
  handleValueChange: (
    key: string,
    { newObj, newLead }: { newObj: string; newLead?: string }
  ) => void;
  parentKey: string;
}

type BibleKey = keyof typeof bibleData;

const BibleSelect: React.FC<BibleSelectProps> = ({
  handleValueChange,
  parentKey,
}) => {
  const selectedDetail = useRecoilValue(selectedDetailState);
  // "신_5/4:5-6, 수_6/5:6"
  // "신명기_5/4:5-4:6, 여호수아_6/5:6"

  //  Selection = {
  //   book: '신';
  //   chapter: 4;
  //   verse: 5;
  // };
  const selectedInitInfo = (() => {
    let sermonsSelection: Selection[][] = [];
    if (selectedDetail.obj) {
      const sermons = selectedDetail.obj.split(/,\s*/);

      sermonsSelection = sermons.map((sEl, _) => {
        const splitStandardUnder = sEl.split("_");
        const book = splitStandardUnder[0];
        const chapter = splitStandardUnder[1].split("/")[1];

        if (chapter.includes("-")) {
          const splitStandardhyphen = chapter.split("-");
          const selections = splitStandardhyphen.map((el) => {
            const splitStandardColon = el.split(":");
            return {
              book,
              chapter: +splitStandardColon[0],
              verse: +splitStandardColon[1],
            };
          });

          return selections;
        } else {
          const splitStandardColon = chapter.split(":");
          return [
            {
              book,
              chapter: +splitStandardColon[0],
              verse: +splitStandardColon[1],
            },
          ];
        }
      });
    }

    return sermonsSelection;
  })();

  const [selectedBook, setSelectedBook] = useState<Selection>({
    book: "",
    chapter: 0,
    verse: 0,
  });

  const [selectedRanges, setSelectedRanges] = useState<Selection[]>([]);
  const [multiSelection, setMultiSelection] =
    useState<Selection[][]>(selectedInitInfo);

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
      if (
        selectedBook.book &&
        selectedBook.chapter > 0 &&
        selectedBook.verse > 0
      ) {
        setSelectedRanges((prev) => [...prev, selectedBook]);
        setSelectedBook({ book: selectedBook.book, chapter: 0, verse: 0 });
      }
    },
    finalizeSelection: () => {
      if (selectedRanges.length === 0) return;

      setMultiSelection((prev) => {
        const updated = [...prev, selectedRanges];

        const finalObj = updated
          .map((ranges) => {
            const first = ranges[0];
            const last = ranges[1] || first;

            return (
              `${first.book}_${bibleData[first.book as BibleKey]?.index}/${
                first.chapter
              }:${first.verse}` +
              (ranges.length > 1 ? `-${last.chapter}:${last.verse}` : "")
            );
          })
          .join(", ");

        handleValueChange(parentKey, {
          newObj: finalObj,
        });

        return updated;
      });

      setSelectedRanges([]);
      setSelectedBook({ book: "", chapter: 0, verse: 0 });
    },
    deleteSelection: (deleteIndex: number) => {
      setMultiSelection((prev) => {
        const updated = prev.filter((_, i) => i !== deleteIndex);
        const finalObj = updated
          .map((ranges) => {
            const first = ranges[0];
            const last = ranges[1] || first;

            return (
              `${first.book}_${bibleData[first.book as BibleKey]?.index}/${
                first.chapter
              }:${first.verse}` +
              (ranges.length > 1 ? `-${last.chapter}:${last.verse}` : "")
            );
          })
          .join(", ");
        handleValueChange(parentKey, {
          newObj: finalObj,
        });

        return updated;
      });
    },
  };

  const currentBook = selectedBook.book
    ? bibleData[selectedBook.book as BibleKey]
    : null;
  const currentChapterVerses =
    currentBook && selectedBook.chapter
      ? currentBook.chapters[selectedBook.chapter - 1]
      : 0;

  const formatRange = (ranges: Selection[]) => {
    return ranges
      .map((range, i) => {
        if (!i) {
          return `${selectedBook.book} ${range.chapter}장 ${range.verse}절`;
        } else {
          return `${range.chapter}장 ${range.verse}절`;
        }
      })
      .join(" ~ ");
  };

  return (
    <>
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
                  value={selectedBook.book || ""}
                >
                  <option value="" disabled>
                    책을 선택하세요
                  </option>
                  {Object.entries(bibleData).map(([key]) => (
                    <option key={key} value={key}>
                      {key}
                    </option>
                  ))}
                </select>
              </label>

              {currentBook && (
                <label className="select-label">
                  장 선택:
                  <select
                    className="select-box"
                    onChange={handler.chapterChange}
                    value={selectedBook.chapter || ""}
                  >
                    <option value="" disabled>
                      장을 선택하세요
                    </option>
                    {currentBook.chapters.map((_, index) => (
                      <option
                        key={index}
                        value={index + 1}
                        disabled={selectedRanges[0]?.chapter > index + 1}
                      >
                        {index + 1}장
                      </option>
                    ))}
                  </select>
                </label>
              )}

              {currentBook && selectedBook.chapter > 0 && (
                <label className="select-label">
                  절 선택:
                  <select
                    className="select-box"
                    onChange={handler.verseChange}
                    value={selectedBook.verse || ""}
                  >
                    <option value="" disabled>
                      절을 선택하세요
                    </option>
                    {Array.from(
                      { length: currentChapterVerses },
                      (_, i) => i + 1
                    ).map((verse, index) => (
                      <option
                        key={verse}
                        value={verse}
                        disabled={
                          selectedRanges[0]?.chapter == selectedBook.chapter
                            ? selectedRanges[0]?.verse > index
                            : false
                        }
                      >
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
              disabled={
                !(
                  selectedBook.book &&
                  selectedBook.chapter > 0 &&
                  selectedBook.verse > 0
                )
              }
            >
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
            }}
          >
            다시 선택
          </button>
        )}
        {selectedRanges.length > 0 && (
          <div className="result-container">
            <p className="result-text">{formatRange(selectedRanges)}</p>
          </div>
        )}
      </div>
      <button
        className="add-selection-button"
        onClick={handler.finalizeSelection}
      >
        구절 추가
      </button>
      {multiSelection.length > 0 && (
        <div className="multi-selection-list">
          {multiSelection.map((ranges, index) => {
            const first = ranges[0];
            const last = ranges[1] || first;

            const displayText =
              `${first.book} ${first.chapter}:${first.verse}` +
              (ranges.length > 1
                ? `-${
                    first.chapter === last.chapter ? "" : `${last.chapter}:`
                  }${last.verse}`
                : "");

            return (
              <span key={index} className="verse-chip">
                📖 {displayText}
                <button
                  className="delete-button"
                  onClick={() => handler.deleteSelection(index)}
                >
                  x
                </button>
              </span>
            );
          })}
        </div>
      )}
    </>
  );
};

export default BibleSelect;
