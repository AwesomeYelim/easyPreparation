"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import { useSetRecoilState, useRecoilValue } from "recoil";
import { displayPanelOpenState, userSettingsState } from "@/recoilState";
import { apiClient, openDisplayWindow } from "@/lib/apiClient";
// bible.scss 마이그레이션 완료 — Tailwind CSS로 전환됨

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
  const settings = useRecoilValue(userSettingsState);
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
          // 선호 버전이 유효하면 적용, 아니면 첫 번째 버전
          const preferred = settings.preferred_bible_version;
          const hasPreferred = data.some((v: Version) => v.id === preferred);
          setVersionId(hasPreferred ? preferred : data[0].id);
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

  // 메인 버전 변경 시 현재 장 다시 로드
  useEffect(() => {
    if (selectedBook && selectedChapter > 0) {
      fetch(
        `${BASE_URL}/api/bible/verses?book=${selectedBook.book_order}&chapter=${selectedChapter}&version=${versionId}`
      )
        .then((r) => r.json())
        .then((data) => {
          if (Array.isArray(data)) setVerses(data);
        })
        .catch((e) => console.error("bible fetch 에러:", e));
    }
  }, [versionId]); // eslint-disable-line react-hooks/exhaustive-deps

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
    if (e.shiftKey && lastClickedRef.current !== null) {
      e.preventDefault();
      const lo = Math.min(lastClickedRef.current, verse);
      const hi = Math.max(lastClickedRef.current, verse);
      setSelectedVerses((prev) => {
        const next = new Set(prev);
        // 범위가 모두 선택된 상태면 해제, 아니면 선택
        let allSelected = true;
        for (let v = lo; v <= hi; v++) {
          if (!next.has(v)) { allSelected = false; break; }
        }
        for (let v = lo; v <= hi; v++) {
          if (allSelected) next.delete(v);
          else next.add(v);
        }
        return next;
      });
    } else {
      lastClickedRef.current = verse;
      setSelectedVerses((prev) => {
        const next = new Set(prev);
        if (next.has(verse)) next.delete(verse);
        else next.add(verse);
        return next;
      });
    }
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
        versionId,
      }], "bible");
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
    <div className="flex flex-col w-full" style={{ height: "calc(100vh - 70px)", overflow: "hidden" }}>

      {/* ── 상단 컨트롤 바 ── */}
      <div className="flex-shrink-0 flex flex-wrap items-center justify-between gap-3 px-6 py-3 bg-white border-b border-outline/50">
        {/* 좌측: 로고 + 버전 선택 */}
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-1">
            <span className="text-[10px] font-black tracking-widest uppercase text-on-surface-variant mr-1">
              Scripture
            </span>
            <h2 className="text-xl font-black tracking-tight text-primary">Bible</h2>
          </div>

          {versions.length > 1 && (
            <select
              className="px-3 py-1.5 text-sm font-semibold rounded-lg border border-outline/60 bg-surface text-on-surface focus:outline-none focus:ring-2 focus:ring-secondary/40 cursor-pointer"
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
              className={`px-3 py-1.5 text-xs font-bold rounded-lg border transition-all ${
                compareMode
                  ? "bg-secondary text-white border-secondary"
                  : "bg-surface-low text-on-surface-variant border-outline/60 hover:bg-surface-high hover:text-secondary hover:border-secondary/40"
              }`}
              onClick={() => setCompareMode(!compareMode)}
              title="비교 모드"
            >
              비교
            </button>
          )}

          {compareMode && versions.length > 1 && (
            <select
              className="px-3 py-1.5 text-sm font-semibold rounded-lg border border-outline/60 bg-surface text-on-surface focus:outline-none focus:ring-2 focus:ring-secondary/40 cursor-pointer"
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

        {/* 우측: 선택 바 + 검색 */}
        <div className="flex items-center gap-3 flex-wrap">
          {selectedVerses.size > 0 && selectedBook && (
            <div className="flex items-center gap-2 px-3 py-1.5 bg-secondary/10 border border-secondary/30 rounded-lg">
              <span className="text-sm font-semibold text-secondary whitespace-nowrap">
                {selectionLabel()}
              </span>
              <button
                className="px-3 py-1 text-xs font-bold bg-secondary text-white rounded-md hover:bg-secondary/90 active:scale-95 transition-all"
                onClick={handleSendToDisplay}
              >
                Display 전송
              </button>
            </div>
          )}

          <div className="flex max-w-xs w-full">
            <input
              type="text"
              placeholder="구절 검색..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyUp={(e) => e.key === "Enter" && handleSearch()}
              className="flex-1 px-4 py-2 text-sm border border-r-0 border-outline/60 rounded-l-lg bg-surface text-on-surface placeholder:text-on-surface-variant/60 focus:outline-none focus:border-secondary focus:ring-2 focus:ring-secondary/20 transition-all"
            />
            <button
              onClick={handleSearch}
              disabled={searching}
              className="px-4 py-2 text-sm font-bold bg-secondary text-white rounded-r-lg hover:bg-secondary/90 disabled:bg-on-surface-variant/40 disabled:cursor-not-allowed transition-all"
            >
              {searching ? "..." : "검색"}
            </button>
          </div>
        </div>
      </div>

      {/* ── 본문 영역 ── */}
      <div className="flex flex-1 min-h-0">

        {/* ── 좌측 사이드바 ── */}
        <aside className="w-44 flex-shrink-0 flex flex-col bg-surface-low border-r border-outline/40 overflow-hidden">
          {/* 구약/신약 탭 */}
          <div className="flex flex-shrink-0 border-b border-outline/40">
            <button
              className={`flex-1 py-2.5 text-xs font-bold text-center transition-all border-b-2 ${
                tab === "ot"
                  ? "text-secondary border-secondary bg-white"
                  : "text-on-surface-variant border-transparent hover:text-on-surface hover:bg-surface-high"
              }`}
              onClick={() => setTab("ot")}
            >
              구약
            </button>
            <button
              className={`flex-1 py-2.5 text-xs font-bold text-center transition-all border-b-2 ${
                tab === "nt"
                  ? "text-secondary border-secondary bg-white"
                  : "text-on-surface-variant border-transparent hover:text-on-surface hover:bg-surface-high"
              }`}
              onClick={() => setTab("nt")}
            >
              신약
            </button>
          </div>

          {/* 책 목록 */}
          <div className="flex-1 overflow-y-auto py-1 [&::-webkit-scrollbar]:w-1 [&::-webkit-scrollbar-thumb]:bg-outline/40 [&::-webkit-scrollbar-thumb]:rounded-full">
            {displayBooks.map((b) => (
              <button
                key={b.book_order}
                className={`flex items-center gap-2 w-full px-3 py-2 text-left text-sm transition-all border-l-[3px] ${
                  selectedBook?.book_order === b.book_order
                    ? "border-secondary bg-secondary/10 text-secondary font-semibold"
                    : "border-transparent text-on-surface hover:bg-secondary/5 hover:text-secondary"
                }`}
                onClick={() => handleBookSelect(b)}
              >
                <span className={`text-xs font-black min-w-[1.5rem] ${
                  selectedBook?.book_order === b.book_order ? "text-secondary" : "text-on-surface-variant"
                }`}>
                  {b.abbr_kor}
                </span>
                <span className="truncate">{b.name_kor}</span>
              </button>
            ))}
          </div>
        </aside>

        {/* ── 우측 메인 ── */}
        <main
          ref={versesRef}
          className="flex-1 overflow-y-auto px-8 py-6 bg-white [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-thumb]:bg-outline/40 [&::-webkit-scrollbar-thumb]:rounded-full"
        >

          {/* ── 장 선택 그리드 ── */}
          {selectedBook && chapterCount > 0 && selectedChapter === 0 && (
            <div className="max-w-2xl mx-auto">
              <h3 className="text-2xl font-bold text-on-surface mb-1">{selectedBook.name_kor}</h3>
              <p className="text-sm text-on-surface-variant mb-5">{chapterCount}장</p>
              <div className="grid gap-1.5" style={{ gridTemplateColumns: "repeat(auto-fill, minmax(44px, 1fr))" }}>
                {Array.from({ length: chapterCount }, (_, i) => i + 1).map((ch) => (
                  <button
                    key={ch}
                    className="aspect-square flex items-center justify-center text-sm font-medium bg-surface-low border border-outline/40 rounded-xl text-on-surface hover:bg-secondary/10 hover:border-secondary/40 hover:text-secondary active:scale-95 transition-all"
                    onClick={() => handleChapterSelect(ch)}
                  >
                    {ch}
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* ── 절 본문 — 비교 모드 ── */}
          {selectedChapter > 0 && verses.length > 0 && compareMode && compareVerses.length > 0 && (
            <div className="max-w-4xl mx-auto">
              {/* 헤더 */}
              <div className="flex items-center justify-center gap-5 mb-7 pb-4 border-b border-outline/40">
                <button
                  className="w-9 h-9 flex items-center justify-center text-lg rounded-full bg-surface-low border border-outline/40 text-on-surface hover:bg-secondary/10 hover:border-secondary/40 disabled:opacity-30 disabled:cursor-default transition-all"
                  disabled={!canPrev}
                  onClick={() => handleChapterSelect(selectedChapter - 1)}
                >
                  ‹
                </button>
                <h3 className="text-xl font-bold text-on-surface min-w-40 text-center">
                  {selectedBook?.name_kor} {selectedChapter}장
                </h3>
                <button
                  className="w-9 h-9 flex items-center justify-center text-lg rounded-full bg-surface-low border border-outline/40 text-on-surface hover:bg-secondary/10 hover:border-secondary/40 disabled:opacity-30 disabled:cursor-default transition-all"
                  disabled={!canNext}
                  onClick={() => handleChapterSelect(selectedChapter + 1)}
                >
                  ›
                </button>
              </div>

              {/* 버전 레이블 */}
              <div className="grid grid-cols-2 gap-3 px-7 pb-3 border-b border-outline/30 mb-4">
                <div className="flex items-center gap-2">
                  <span className="w-2 h-2 rounded-full bg-secondary flex-shrink-0" />
                  <span className="text-xs font-black uppercase tracking-wider text-primary">
                    {getVersionName(versionId)}
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="w-2 h-2 rounded-full bg-accent-cyan flex-shrink-0" />
                  <span className="text-xs font-black uppercase tracking-wider text-primary">
                    {getVersionName(compareVersionId)}
                  </span>
                </div>
              </div>

              <p className="text-[11px] text-on-surface-variant mb-3 px-1">
                클릭하여 선택 · Shift+클릭으로 범위 선택
              </p>

              {/* 비교 절 목록 */}
              <div className="pb-5">
                {verses.map((v) => {
                  const cv = compareVerses.find((c) => c.verse === v.verse);
                  const isSelected = selectedVerses.has(v.verse);
                  return (
                    <div
                      key={v.verse}
                      className={`grid gap-3 px-1 py-2 rounded-lg cursor-pointer select-none transition-colors ${
                        isSelected
                          ? "bg-secondary/10"
                          : "hover:bg-surface-low"
                      }`}
                      style={{ gridTemplateColumns: "28px 1fr 1fr" }}
                      onClick={(e) => handleVerseClick(e, v.verse)}
                    >
                      <sup className={`text-[11px] font-black text-right pt-0.5 ${
                        isSelected ? "text-secondary" : "text-secondary"
                      }`}>
                        {v.verse}
                      </sup>
                      <div className={`text-[15px] leading-[1.8] ${
                        isSelected ? "text-secondary font-medium" : "text-on-surface"
                      }`}>
                        {v.text}
                      </div>
                      <div className={`text-[15px] leading-[1.8] border-l-2 border-outline/30 pl-3 ${
                        isSelected ? "text-secondary font-medium" : "text-on-surface-variant"
                      }`}>
                        {cv?.text || ""}
                      </div>
                    </div>
                  );
                })}
              </div>

              {/* 푸터 */}
              <div className="flex justify-between pt-5 border-t border-outline/40 mt-3">
                <button
                  className="px-5 py-2.5 text-sm font-medium rounded-lg bg-surface-low border border-outline/40 text-on-surface hover:bg-secondary/10 hover:border-secondary/40 hover:text-secondary disabled:opacity-30 disabled:cursor-default transition-all"
                  disabled={!canPrev}
                  onClick={() => handleChapterSelect(selectedChapter - 1)}
                >
                  ‹ 이전 장
                </button>
                <button
                  className="px-5 py-2.5 text-sm font-medium rounded-lg bg-surface-low border border-outline/40 text-on-surface hover:bg-secondary/10 hover:border-secondary/40 hover:text-secondary disabled:opacity-30 disabled:cursor-default transition-all"
                  disabled={!canNext}
                  onClick={() => handleChapterSelect(selectedChapter + 1)}
                >
                  다음 장 ›
                </button>
              </div>
            </div>
          )}

          {/* ── 절 본문 — 일반 모드 ── */}
          {selectedChapter > 0 && verses.length > 0 && !(compareMode && compareVerses.length > 0) && (
            <div className="max-w-2xl mx-auto">
              {/* 헤더 */}
              <div className="flex items-center justify-center gap-5 mb-7 pb-4 border-b border-outline/40">
                <button
                  className="w-9 h-9 flex items-center justify-center text-lg rounded-full bg-surface-low border border-outline/40 text-on-surface hover:bg-secondary/10 hover:border-secondary/40 disabled:opacity-30 disabled:cursor-default transition-all"
                  disabled={!canPrev}
                  onClick={() => handleChapterSelect(selectedChapter - 1)}
                >
                  ‹
                </button>
                <h3 className="text-xl font-bold text-on-surface min-w-40 text-center">
                  {selectedBook?.name_kor} {selectedChapter}장
                </h3>
                <button
                  className="w-9 h-9 flex items-center justify-center text-lg rounded-full bg-surface-low border border-outline/40 text-on-surface hover:bg-secondary/10 hover:border-secondary/40 disabled:opacity-30 disabled:cursor-default transition-all"
                  disabled={!canNext}
                  onClick={() => handleChapterSelect(selectedChapter + 1)}
                >
                  ›
                </button>
              </div>

              <p className="text-[11px] text-on-surface-variant mb-3 px-1">
                클릭하여 선택 · Shift+클릭으로 범위 선택
              </p>

              {/* 절 목록 */}
              <div className="pb-5">
                {verses.map((v) => {
                  const isSelected = selectedVerses.has(v.verse);
                  return (
                    <p
                      key={v.verse}
                      data-verse={v.verse}
                      className={`group relative text-base leading-loose px-2 py-0.5 rounded-lg cursor-pointer select-none transition-colors ${
                        isSelected
                          ? "bg-secondary/10 text-secondary"
                          : "text-on-surface hover:bg-surface-low"
                      }`}
                      onClick={(e) => handleVerseClick(e, v.verse)}
                    >
                      <sup className={`font-black text-[11px] mr-1.5 ${
                        isSelected ? "text-secondary" : "text-secondary"
                      }`}>
                        {v.verse}
                      </sup>
                      {v.text}
                    </p>
                  );
                })}
              </div>

              {/* 푸터 */}
              <div className="flex justify-between pt-5 border-t border-outline/40 mt-3">
                <button
                  className="px-5 py-2.5 text-sm font-medium rounded-lg bg-surface-low border border-outline/40 text-on-surface hover:bg-secondary/10 hover:border-secondary/40 hover:text-secondary disabled:opacity-30 disabled:cursor-default transition-all"
                  disabled={!canPrev}
                  onClick={() => handleChapterSelect(selectedChapter - 1)}
                >
                  ‹ 이전 장
                </button>
                <button
                  className="px-5 py-2.5 text-sm font-medium rounded-lg bg-surface-low border border-outline/40 text-on-surface hover:bg-secondary/10 hover:border-secondary/40 hover:text-secondary disabled:opacity-30 disabled:cursor-default transition-all"
                  disabled={!canNext}
                  onClick={() => handleChapterSelect(selectedChapter + 1)}
                >
                  다음 장 ›
                </button>
              </div>
            </div>
          )}

          {/* ── 검색 결과 ── */}
          {searchResults.length > 0 && (
            <div className="max-w-2xl mx-auto">
              <h3 className="text-lg font-bold text-on-surface mb-4">
                검색 결과 ({searchResults.length}건)
              </h3>
              {searchResults.map((r, i) => (
                <div
                  key={i}
                  className="px-4 py-3 border border-outline/40 rounded-xl mb-2 cursor-pointer transition-all hover:border-secondary/40 hover:bg-secondary/5 hover:shadow-sm"
                  onClick={() => handleSearchResultClick(r)}
                >
                  <div className="text-[13px] font-bold text-secondary mb-1">
                    {r.book_name} {r.chapter}:{r.verse}
                  </div>
                  <div className="text-sm text-on-surface-variant leading-relaxed">{r.text}</div>
                </div>
              ))}
            </div>
          )}

          {/* ── 빈 상태 ── */}
          {!selectedBook && searchResults.length === 0 && (
            <div className="flex flex-col items-center justify-center h-full min-h-[300px] text-on-surface-variant">
              <div className="w-16 h-16 flex items-center justify-center text-3xl font-light bg-surface-low rounded-full mb-4 text-outline border border-outline/30">
                +
              </div>
              <p className="text-[15px] my-0.5">성경 책을 선택하거나</p>
              <p className="text-[15px] my-0.5">검색어를 입력하세요</p>
            </div>
          )}
        </main>
      </div>
    </div>
  );
}
