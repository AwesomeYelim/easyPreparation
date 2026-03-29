"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import { useSetRecoilState } from "recoil";
import { displayPanelOpenState } from "@/recoilState";
import { apiClient, openDisplayWindow } from "@/lib/apiClient";
import "./bible.scss";

const BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL;

type Book = {
  name_kor: string;
  abbr_kor: string;
  book_order: number;
  chapterCount: number;
};

type Version = { id: number; name: string; code: string };
type Verse = { verse: number; text: string };
type SearchResult = {
  book_name: string;
  book_order: number;
  chapter: number;
  verse: number;
  text: string;
};

export default function BiblePage() {
  const [versions, setVersions] = useState<Version[]>([]);
  const [versionId, setVersionId] = useState(1);
  const [books, setBooks] = useState<Book[]>([]);
  const [selectedBook, setSelectedBook] = useState<Book | null>(null);
  const [chapterCount, setChapterCount] = useState(0);
  const [selectedChapter, setSelectedChapter] = useState(0);
  const [verses, setVerses] = useState<Verse[]>([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<SearchResult[]>([]);
  const [searching, setSearching] = useState(false);
  const [tab, setTab] = useState<"ot" | "nt">("ot");
  const [selectedVerses, setSelectedVerses] = useState<Set<number>>(new Set());
  const setDisplayPanelOpen = useSetRecoilState(displayPanelOpenState);
  const versesRef = useRef<HTMLDivElement>(null);

  // 비교 모드
  const [compareMode, setCompareMode] = useState(false);
  const [compareVersionId, setCompareVersionId] = useState(0);
  const [compareVerses, setCompareVerses] = useState<Verse[]>([]);

  useEffect(() => {
    fetch(`${BASE_URL}/api/bible/versions`)
      .then((r) => r.json())
      .then((data) => {
        if (Array.isArray(data) && data.length > 0) {
          setVersions(data);
          setVersionId(data[0].id);
          if (data.length > 1) setCompareVersionId(data[1].id);
        }
      })
      .catch((e) => console.error("bible fetch 에러:", e));

    fetch(`${BASE_URL}/api/bible/books`)
      .then((r) => r.json())
      .then((data) => {
        if (data && !Array.isArray(data) && typeof data === "object") {
          const parsed: Book[] = Object.entries(data).map(([name, info]: [string, any]) => ({
            name_kor: name,
            abbr_kor: info.kor || name,
            book_order: info.index,
            chapterCount: info.chapters?.length || 0,
          }));
          parsed.sort((a, b) => a.book_order - b.book_order);
          setBooks(parsed);
        }
      })
      .catch((e) => console.error("bible fetch 에러:", e));
  }, []);

  const handleBookSelect = useCallback((book: Book) => {
    setSelectedBook(book);
    setSelectedChapter(0);
    setVerses([]);
    setCompareVerses([]);
    setSearchResults([]);
    setChapterCount(book.chapterCount);
    setSelectedVerses(new Set());
  }, []);

  const fetchChapter = useCallback(
    (bookOrder: number, chapter: number, highlightVerse?: number) => {
      setSelectedChapter(chapter);
      setSearchResults([]);
      setSelectedVerses(highlightVerse ? new Set([highlightVerse]) : new Set());
      versesRef.current?.scrollTo({ top: 0 });

      fetch(
        `${BASE_URL}/api/bible/verses?book=${bookOrder}&chapter=${chapter}&version=${versionId}`
      )
        .then((r) => r.json())
        .then((data) => {
          if (Array.isArray(data)) {
            setVerses(data);
            if (highlightVerse) {
              requestAnimationFrame(() => {
                const el = versesRef.current?.querySelector(`[data-verse="${highlightVerse}"]`);
                el?.scrollIntoView({ block: "center", behavior: "smooth" });
              });
            }
          }
        })
        .catch((e) => console.error("bible fetch 에러:", e));

      // 비교 모드가 켜져 있으면 비교 버전도 fetch
      if (compareMode && compareVersionId) {
        fetch(
          `${BASE_URL}/api/bible/verses?book=${bookOrder}&chapter=${chapter}&version=${compareVersionId}`
        )
          .then((r) => r.json())
          .then((data) => {
            if (Array.isArray(data)) setCompareVerses(data);
            else setCompareVerses([]);
          })
          .catch(() => setCompareVerses([]));
      }
    },
    [versionId, compareMode, compareVersionId]
  );

  // 비교 모드 토글 시 / 비교 버전 변경 시 비교 데이터 fetch
  useEffect(() => {
    if (compareMode && compareVersionId && selectedBook && selectedChapter > 0) {
      fetch(
        `${BASE_URL}/api/bible/verses?book=${selectedBook.book_order}&chapter=${selectedChapter}&version=${compareVersionId}`
      )
        .then((r) => r.json())
        .then((data) => {
          if (Array.isArray(data)) setCompareVerses(data);
          else setCompareVerses([]);
        })
        .catch(() => setCompareVerses([]));
    } else {
      setCompareVerses([]);
    }
  }, [compareMode, compareVersionId, selectedBook, selectedChapter]);

  const handleChapterSelect = useCallback(
    (chapter: number) => {
      if (!selectedBook) return;
      fetchChapter(selectedBook.book_order, chapter);
    },
    [selectedBook, fetchChapter]
  );

  const handleSearch = useCallback(() => {
    if (!searchQuery.trim()) return;
    setSearching(true);
    setSelectedBook(null);
    setSelectedChapter(0);
    setVerses([]);
    setCompareVerses([]);

    fetch(`${BASE_URL}/api/bible/search?q=${encodeURIComponent(searchQuery)}&version=${versionId}`)
      .then((r) => r.json())
      .then((data) => {
        if (Array.isArray(data)) setSearchResults(data);
        else setSearchResults([]);
      })
      .catch(() => setSearchResults([]))
      .finally(() => setSearching(false));
  }, [searchQuery, versionId]);

  const handleSearchResultClick = useCallback(
    (result: SearchResult) => {
      const book = books.find((b) => b.book_order === result.book_order);
      if (book) {
        setSelectedBook(book);
        setChapterCount(book.chapterCount);
        setTab(book.book_order <= 39 ? "ot" : "nt");
        fetchChapter(book.book_order, result.chapter, result.verse);
      }
    },
    [books, fetchChapter]
  );

  const lastClickedRef = useRef<number | null>(null);

  const handleVerseClick = useCallback((e: React.MouseEvent, verse: number) => {
    setSelectedVerses((prev) => {
      const next = new Set(prev);
      if (e.shiftKey && lastClickedRef.current !== null) {
        const lo = Math.min(lastClickedRef.current, verse);
        const hi = Math.max(lastClickedRef.current, verse);
        for (let v = lo; v <= hi; v++) next.add(v);
      } else {
        if (next.has(verse)) next.delete(verse);
        else next.add(verse);
      }
      lastClickedRef.current = verse;
      return next;
    });
  }, []);

  const groupRanges = useCallback((verses: number[]): [number, number][] => {
    const sorted = [...verses].sort((a, b) => a - b);
    const ranges: [number, number][] = [];
    let start = sorted[0], end = sorted[0];
    for (let i = 1; i < sorted.length; i++) {
      if (sorted[i] === end + 1) {
        end = sorted[i];
      } else {
        ranges.push([start, end]);
        start = end = sorted[i];
      }
    }
    ranges.push([start, end]);
    return ranges;
  }, []);

  const selectionLabel = useCallback(() => {
    if (selectedVerses.size === 0 || !selectedBook) return "";
    const ranges = groupRanges([...selectedVerses]);
    const parts = ranges.map(([s, e]) => s === e ? `${s}` : `${s}-${e}`);
    return `${selectedBook.name_kor} ${selectedChapter}:${parts.join(", ")}`;
  }, [selectedVerses, selectedBook, selectedChapter, groupRanges]);

  const handleSendToDisplay = useCallback(async () => {
    if (!selectedBook || selectedChapter === 0 || selectedVerses.size === 0) return;
    const ranges = groupRanges([...selectedVerses]);
    const objParts = ranges.map(([s, e]) =>
      `${selectedBook.name_kor}_${selectedBook.book_order}/${selectedChapter}:${s}` +
      (e > s ? `-${selectedChapter}:${e}` : "")
    );
    const obj = objParts.join(", ");

    try {
      setDisplayPanelOpen(true);
      openDisplayWindow();
      await apiClient.appendToDisplay([{
        title: "성경",
        info: "b_edit",
        obj,
      }]);
    } catch (e) {
      console.error("Display 전송 에러:", e);
    }
  }, [selectedBook, selectedChapter, selectedVerses, groupRanges, setDisplayPanelOpen]);

  const otBooks = books.filter((b) => b.book_order <= 39);
  const ntBooks = books.filter((b) => b.book_order > 39);
  const displayBooks = tab === "ot" ? otBooks : ntBooks;

  const canPrev = selectedChapter > 1;
  const canNext = selectedChapter > 0 && selectedChapter < chapterCount;

  const getVersionName = (id: number) => versions.find((v) => v.id === id)?.name || "";

  return (
    <div className="bible_page">
      {/* 상단 바 */}
      <div className="bible_header">
        <div className="bible_header_left">
          <h2 className="bible_logo">Bible</h2>
          {versions.length > 1 && (
            <select
              className="bible_version_select"
              value={versionId}
              onChange={(e) => setVersionId(Number(e.target.value))}
            >
              {versions.map((v) => (
                <option key={v.id} value={v.id}>{v.name}</option>
              ))}
            </select>
          )}
          {versions.length > 1 && (
            <button
              className={`bible_compare_btn${compareMode ? " active" : ""}`}
              onClick={() => setCompareMode(!compareMode)}
              title="비교 모드"
            >
              비교
            </button>
          )}
          {compareMode && versions.length > 1 && (
            <select
              className="bible_version_select"
              value={compareVersionId}
              onChange={(e) => setCompareVersionId(Number(e.target.value))}
            >
              {versions
                .filter((v) => v.id !== versionId)
                .map((v) => (
                  <option key={v.id} value={v.id}>{v.name}</option>
                ))}
            </select>
          )}
        </div>
        <div className="bible_header_right">
          {selectedVerses.size > 0 && selectedBook && (
            <div className="bible_display_bar">
              <span>{selectionLabel()}</span>
              <button className="display_send_btn" onClick={handleSendToDisplay}>
                Display 전송
              </button>
            </div>
          )}
          <div className="bible_search">
            <input
              type="text"
              placeholder="구절 검색..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyUp={(e) => e.key === "Enter" && handleSearch()}
            />
            <button onClick={handleSearch} disabled={searching}>
              {searching ? "..." : "검색"}
            </button>
          </div>
        </div>
      </div>

      <div className="bible_body">
        {/* 좌측 사이드바 */}
        <aside className="bible_sidebar">
          <div className="bible_tabs">
            <button
              className={`bible_tab ${tab === "ot" ? "active" : ""}`}
              onClick={() => setTab("ot")}
            >
              구약
            </button>
            <button
              className={`bible_tab ${tab === "nt" ? "active" : ""}`}
              onClick={() => setTab("nt")}
            >
              신약
            </button>
          </div>
          <div className="bible_book_list">
            {displayBooks.map((b) => (
              <button
                key={b.book_order}
                className={`bible_book_btn ${selectedBook?.book_order === b.book_order ? "active" : ""}`}
                onClick={() => handleBookSelect(b)}
              >
                <span className="book_abbr">{b.abbr_kor}</span>
                <span className="book_full">{b.name_kor}</span>
              </button>
            ))}
          </div>
        </aside>

        {/* 우측 콘텐츠 */}
        <main className="bible_main" ref={versesRef}>
          {/* 장 선택 그리드 */}
          {selectedBook && chapterCount > 0 && selectedChapter === 0 && (
            <div className="bible_chapter_picker">
              <h3>{selectedBook.name_kor}</h3>
              <p className="chapter_subtitle">{chapterCount}장</p>
              <div className="bible_chapter_grid">
                {Array.from({ length: chapterCount }, (_, i) => i + 1).map((ch) => (
                  <button
                    key={ch}
                    className="bible_ch_btn"
                    onClick={() => handleChapterSelect(ch)}
                  >
                    {ch}
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* 절 본문 — 비교 모드 */}
          {selectedChapter > 0 && verses.length > 0 && compareMode && compareVerses.length > 0 && (
            <div className="bible_reading bible_compare_view">
              <div className="reading_header">
                <button
                  className="ch_nav_btn"
                  disabled={!canPrev}
                  onClick={() => handleChapterSelect(selectedChapter - 1)}
                >
                  ‹
                </button>
                <h3>
                  {selectedBook?.name_kor} {selectedChapter}장
                </h3>
                <button
                  className="ch_nav_btn"
                  disabled={!canNext}
                  onClick={() => handleChapterSelect(selectedChapter + 1)}
                >
                  ›
                </button>
              </div>
              <div className="compare_labels">
                <span className="compare_label">{getVersionName(versionId)}</span>
                <span className="compare_label">{getVersionName(compareVersionId)}</span>
              </div>
              <div className="bible_verses_compare">
                {verses.map((v) => {
                  const cv = compareVerses.find((c) => c.verse === v.verse);
                  return (
                    <div
                      key={v.verse}
                      className={`verse_compare_row${selectedVerses.has(v.verse) ? " selected" : ""}`}
                      onClick={(e) => handleVerseClick(e, v.verse)}
                    >
                      <sup className="verse_num">{v.verse}</sup>
                      <div className="verse_col">{v.text}</div>
                      <div className="verse_col compare">{cv?.text || ""}</div>
                    </div>
                  );
                })}
              </div>
              <div className="reading_footer">
                <button
                  className="ch_footer_btn"
                  disabled={!canPrev}
                  onClick={() => handleChapterSelect(selectedChapter - 1)}
                >
                  ‹ 이전 장
                </button>
                <button
                  className="ch_footer_btn"
                  disabled={!canNext}
                  onClick={() => handleChapterSelect(selectedChapter + 1)}
                >
                  다음 장 ›
                </button>
              </div>
            </div>
          )}

          {/* 절 본문 — 일반 모드 */}
          {selectedChapter > 0 && verses.length > 0 && !(compareMode && compareVerses.length > 0) && (
            <div className="bible_reading">
              <div className="reading_header">
                <button
                  className="ch_nav_btn"
                  disabled={!canPrev}
                  onClick={() => handleChapterSelect(selectedChapter - 1)}
                >
                  ‹
                </button>
                <h3>
                  {selectedBook?.name_kor} {selectedChapter}장
                </h3>
                <button
                  className="ch_nav_btn"
                  disabled={!canNext}
                  onClick={() => handleChapterSelect(selectedChapter + 1)}
                >
                  ›
                </button>
              </div>
              <div className="bible_verses">
                {verses.map((v) => (
                  <p
                    key={v.verse}
                    data-verse={v.verse}
                    className={`verse_line${selectedVerses.has(v.verse) ? " selected" : ""}`}
                    onClick={(e) => handleVerseClick(e, v.verse)}
                  >
                    <sup>{v.verse}</sup>
                    {v.text}
                  </p>
                ))}
              </div>
              <div className="reading_footer">
                <button
                  className="ch_footer_btn"
                  disabled={!canPrev}
                  onClick={() => handleChapterSelect(selectedChapter - 1)}
                >
                  ‹ 이전 장
                </button>
                <button
                  className="ch_footer_btn"
                  disabled={!canNext}
                  onClick={() => handleChapterSelect(selectedChapter + 1)}
                >
                  다음 장 ›
                </button>
              </div>
            </div>
          )}

          {/* 검색 결과 */}
          {searchResults.length > 0 && (
            <div className="bible_search_results">
              <h3>검색 결과 ({searchResults.length}건)</h3>
              {searchResults.map((r, i) => (
                <div
                  key={i}
                  className="search_result_card"
                  onClick={() => handleSearchResultClick(r)}
                >
                  <div className="sr_ref">
                    {r.book_name} {r.chapter}:{r.verse}
                  </div>
                  <div className="sr_text">{r.text}</div>
                </div>
              ))}
            </div>
          )}

          {/* 빈 상태 */}
          {!selectedBook && searchResults.length === 0 && (
            <div className="bible_empty">
              <div className="empty_icon">+</div>
              <p>성경 책을 선택하거나</p>
              <p>검색어를 입력하세요</p>
            </div>
          )}
        </main>
      </div>
    </div>
  );
}
