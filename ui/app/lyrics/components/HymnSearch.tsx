"use client";

import { useState, useCallback } from "react";
import { useSetRecoilState } from "recoil";
import { displayPanelOpenState } from "@/recoilState";
import { apiClient, openDisplayWindow } from "@/lib/apiClient";
import { Hymn } from "@/types";
import "./HymnSearch.scss";

export default function HymnSearch() {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<Hymn[]>([]);
  const [searching, setSearching] = useState(false);
  const [selected, setSelected] = useState<Hymn | null>(null);
  const setDisplayPanelOpen = useSetRecoilState(displayPanelOpenState);

  const handleSearch = useCallback(() => {
    if (!query.trim()) return;
    setSearching(true);
    setSelected(null);
    apiClient
      .searchHymns(query)
      .then((data: Hymn[]) => setResults(Array.isArray(data) ? data : []))
      .catch(() => setResults([]))
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
    try {
      setDisplayPanelOpen(true);
      openDisplayWindow();
      await apiClient.appendToDisplay([
        {
          title: "찬송",
          info: "c_edit",
          obj: `${selected.number}장`,
        },
      ]);
    } catch (e) {
      console.error("Display 전송 에러:", e);
    }
  }, [selected, setDisplayPanelOpen]);

  return (
    <div className="hymn_search_container">
      <div className="hymn_search_bar">
        <input
          type="text"
          placeholder="찬송가 번호 또는 제목을 입력하세요..."
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyUp={(e) => e.key === "Enter" && handleSearch()}
        />
        <button onClick={handleSearch} disabled={searching}>
          {searching ? "..." : "검색"}
        </button>
      </div>

      <div className="hymn_search_body">
        {/* 검색 결과 목록 */}
        <div className="hymn_results">
          {results.length === 0 ? (
            <div className="hymn_empty">
              {query ? "검색 결과가 없습니다" : "번호나 제목으로 찬송가를 검색하세요"}
            </div>
          ) : (
            results.map((h) => (
              <div
                key={`${h.hymnbook}-${h.number}`}
                className={`hymn_result_item${
                  selected?.number === h.number && selected?.hymnbook === h.hymnbook
                    ? " active"
                    : ""
                }`}
                onClick={() => handleSelect(h)}
              >
                <span className="hymn_result_num">{h.number}</span>
                <div className="hymn_result_info">
                  <div className="hymn_result_title">{h.title}</div>
                  {h.first_line && (
                    <div className="hymn_result_line">{h.first_line}</div>
                  )}
                </div>
              </div>
            ))
          )}
        </div>

        {/* 상세 뷰 */}
        <div className="hymn_detail_pane">
          {selected ? (
            <>
              <div className="hymn_detail_top">
                <h3>
                  {selected.number}장 — {selected.title}
                </h3>
                <button className="hymn_send_btn" onClick={handleSendToDisplay}>
                  Display 전송
                </button>
              </div>
              {selected.lyrics ? (
                <pre className="hymn_lyrics_text">{selected.lyrics}</pre>
              ) : (
                <div className="hymn_no_lyrics">가사 데이터가 없습니다</div>
              )}
            </>
          ) : (
            <div className="hymn_empty">찬송가를 선택하세요</div>
          )}
        </div>
      </div>
    </div>
  );
}
