"use client";

import { useState, useCallback } from "react";
import { useSetRecoilState } from "recoil";
import { displayPanelOpenState } from "@/recoilState";
import { apiClient, openDisplayWindow } from "@/lib/apiClient";
import { Hymn } from "@/types";

export default function HymnSearch() {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<Hymn[]>([]);
  const [searching, setSearching] = useState(false);
  const [selected, setSelected] = useState<Hymn | null>(null);
  const [pickedHymns, setPickedHymns] = useState<Hymn[]>([]);
  const setDisplayPanelOpen = useSetRecoilState(displayPanelOpenState);

  const handleSearch = useCallback(() => {
    if (!query.trim()) return;
    setSearching(true);
    setSelected(null);
    apiClient
        .searchHymns(query)
        .then((data: Hymn[]) => {
          const newData = Array.isArray(data) ? data : [];
          setResults((prev) => {
            const existing = new Set(prev.map((h) => `${h.hymnbook}-${h.number}`));
            const unique = newData.filter((h) => !existing.has(`${h.hymnbook}-${h.number}`));
            return [...prev, ...unique];
          });
        })
        .catch(() => {})
        .finally(() => setSearching(false));
  }, [query]);

  const handleSelect = useCallback((hymn: Hymn) => {
    if (hymn.lyrics) {
      setSelected(hymn);
    } else {
      apiClient
          .getHymnDetail(hymn.number, hymn.hymnbook)
          .then((data: Hymn) => setSelected(data))
          .catch(() => setSelected(hymn));
    }
  }, []);

  const handleSendToDisplay = useCallback(async () => {
    if (!selected) return;
    // 중복 방지 후 picked 리스트에 즉시 추가
    setPickedHymns((prev) => {
      if (prev.some((h) => h.number === selected.number && h.hymnbook === selected.hymnbook)) return prev;
      return [...prev, selected];
    });
    try {
      setDisplayPanelOpen(true);
      openDisplayWindow();
      await apiClient.appendToDisplay([
        {
          title: "찬송",
          info: "c_edit",
          obj: `${selected.number}장`,
        },
      ], "lyrics");
    } catch (e) {
      console.error("Display 전송 에러:", e);
    }
  }, [selected, setDisplayPanelOpen]);

  return (
      <div className="flex h-full min-h-[500px]">
        {/* 좌측: 검색 & 목록 */}
        <section className="w-80 flex-shrink-0 border-r border-outline/30 flex flex-col bg-surface-low rounded-bl-xl">
          {/* 검색 헤더 */}
          <div className="p-5 space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-black tracking-tight text-primary">찬송가 검색</h2>
              {results.length > 0 && (
                  <span className="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider">
                {results.length}곡
              </span>
              )}
            </div>

            {/* 검색 입력 */}
            <div className="relative">
              <input
                  type="text"
                  placeholder="번호 또는 제목을 입력하세요..."
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  onKeyUp={(e) => e.key === "Enter" && handleSearch()}
                  className="w-full h-11 pl-10 pr-4 bg-white border border-outline/40 rounded-lg text-sm text-on-surface placeholder:text-on-surface-variant focus:outline-none focus:ring-1 focus:ring-secondary focus:border-secondary transition-all"
              />
              <svg
                  className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-on-surface-variant"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
              >
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </div>

            <button
                onClick={handleSearch}
                disabled={searching}
                className="w-full py-2.5 bg-secondary text-white text-sm font-bold rounded-lg hover:bg-secondary/90 active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
            >
              {searching ? "검색 중..." : "검색"}
            </button>
          </div>

          {/* 결과 목록 */}
          <div className="flex-1 overflow-y-auto divide-y divide-outline/10 [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-thumb]:bg-outline/40 [&::-webkit-scrollbar-thumb]:rounded-full">
            {results.length === 0 ? (
                <div className="flex items-center justify-center min-h-[200px] text-sm text-on-surface-variant px-4 text-center">
                  {query ? "검색 결과가 없습니다" : "번호나 제목으로 찬송가를 검색하세요"}
                </div>
            ) : (
                results.map((h) => {
                  const isActive =
                      selected?.number === h.number && selected?.hymnbook === h.hymnbook;
                  return (
                      <div
                          key={`${h.hymnbook}-${h.number}`}
                          onClick={() => handleSelect(h)}
                          className={`flex items-center gap-3 px-5 py-3.5 cursor-pointer transition-colors group ${
                              isActive
                                  ? "bg-secondary/5 border-l-4 border-secondary"
                                  : "hover:bg-white border-l-4 border-transparent"
                          }`}
                      >
                  <span
                      className={`text-sm font-black w-8 text-center flex-shrink-0 ${
                          isActive ? "text-secondary" : "text-on-surface-variant"
                      }`}
                  >
                    {h.number}
                  </span>
                        <div className="flex-1 min-w-0">
                          <div
                              className={`text-sm font-bold truncate ${
                                  isActive ? "text-primary" : "text-on-surface"
                              }`}
                          >
                            {h.title}
                          </div>
                          {h.first_line && (
                              <div className="text-[11px] text-on-surface-variant truncate mt-0.5">
                                {h.first_line}
                              </div>
                          )}
                        </div>
                        <svg
                            className="w-4 h-4 text-on-surface-variant opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                        >
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                        </svg>
                      </div>
                  );
                })
            )}
          </div>

          {/* 전송된 찬송가 목록 */}
          {pickedHymns.length > 0 && (
              <div className="border-t border-outline/30 bg-white">
                <div className="px-5 py-3 flex items-center justify-between">
              <span className="text-[10px] font-black text-secondary uppercase tracking-widest">
                전송 목록 ({pickedHymns.length})
              </span>
                  <button
                      onClick={() => setPickedHymns([])}
                      className="text-[10px] text-on-surface-variant hover:text-red-500 transition-colors"
                  >
                    전체 삭제
                  </button>
                </div>
                <div className="max-h-40 overflow-y-auto px-3 pb-3 space-y-1 [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-thumb]:bg-outline/40 [&::-webkit-scrollbar-thumb]:rounded-full">
                  {pickedHymns.map((h) => (
                      <div
                          key={`picked-${h.hymnbook}-${h.number}`}
                          className="flex items-center gap-2 px-3 py-2 bg-secondary/5 rounded-lg group"
                      >
                  <span className="text-xs font-black text-secondary w-7 text-center flex-shrink-0">
                    {h.number}
                  </span>
                        <span className="text-xs font-medium text-on-surface truncate flex-1">
                    {h.title}
                  </span>
                        <button
                            onClick={() =>
                                setPickedHymns((prev) =>
                                    prev.filter((p) => !(p.number === h.number && p.hymnbook === h.hymnbook))
                                )
                            }
                            className="w-4 h-4 flex items-center justify-center rounded-full text-on-surface-variant hover:bg-red-100 hover:text-red-500 opacity-0 group-hover:opacity-100 transition-all flex-shrink-0 text-[10px]"
                        >
                          ×
                        </button>
                      </div>
                  ))}
                </div>
              </div>
          )}
        </section>

        {/* 우측: 상세 뷰 */}
        <section className="flex-1 flex flex-col min-w-0 overflow-hidden">
          {selected ? (
              <>
                {/* 상세 헤더 */}
                <div className="px-8 py-5 border-b border-outline/20 bg-white flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
                  <div>
                    <div className="flex items-center gap-2 mb-2">
                  <span className="px-2 py-0.5 bg-secondary/10 text-secondary text-[10px] font-black rounded uppercase tracking-widest">
                    찬송 #{selected.number}
                  </span>
                    </div>
                    <h3 className="text-2xl font-black text-primary tracking-tight">
                      {selected.title}
                    </h3>
                    <p className="text-sm text-on-surface-variant mt-1">
                      {selected.number}장
                    </p>
                  </div>
                  <button
                      onClick={handleSendToDisplay}
                      className="flex items-center gap-2 px-5 py-2.5 bg-secondary text-white text-sm font-bold rounded-lg hover:bg-secondary/90 active:scale-95 transition-all shadow-md flex-shrink-0"
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7" />
                    </svg>
                    Display 전송
                  </button>
                </div>

                {/* 가사 내용 */}
                <div className="flex-1 overflow-y-auto p-8 [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-thumb]:bg-outline/40 [&::-webkit-scrollbar-thumb]:rounded-full">
                  {selected.lyrics ? (
                      <div className="max-w-2xl">
                        {/* 코러스/절 구분: 빈 줄 기준 블록 분리 */}
                        {selected.lyrics.split(/\n\s*\n/).map((block, blockIdx) => {
                          const lines = block.trim().split("\n");
                          const isChorus =
                              lines[0]?.trim().match(/^(후렴|코러스|chorus|refrain)/i) !== null;
                          return (
                              <div
                                  key={blockIdx}
                                  className={`mb-7 ${
                                      isChorus
                                          ? "pl-5 border-l-4 border-secondary bg-secondary/5 py-4 pr-4 rounded-r-lg"
                                          : ""
                                  }`}
                              >
                        <pre className={`font-sans whitespace-pre-wrap leading-relaxed ${
                            isChorus
                                ? "text-lg font-bold text-primary"
                                : "text-base text-on-surface font-medium"
                        }`}>
                          {block.trim()}
                        </pre>
                              </div>
                          );
                        })}
                      </div>
                  ) : (
                      <div className="flex items-center justify-center min-h-[200px] text-sm text-on-surface-variant">
                        가사 데이터가 없습니다
                      </div>
                  )}
                </div>
              </>
          ) : (
              /* 미선택 상태 */
              <div className="flex-1 flex flex-col items-center justify-center gap-3 text-on-surface-variant">
                <div className="w-16 h-16 rounded-full bg-surface-high flex items-center justify-center">
                  <svg className="w-8 h-8 text-on-surface-variant/50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3" />
                  </svg>
                </div>
                <p className="text-sm font-medium">찬송가를 선택하세요</p>
                <p className="text-xs text-on-surface-variant/60">좌측 목록에서 찬송가를 클릭하면 가사가 표시됩니다</p>
              </div>
          )}
        </section>
      </div>
  );
}
