"use client";

import { useState } from "react";
import { useRecoilValue, useRecoilState, useSetRecoilState } from "recoil";
import { userInfoState, lyricsSongsState, displayPanelOpenState } from "@/recoilState";
import { apiClient, openDisplayWindow } from "@/lib/apiClient";
import classNames from "classnames";
import "./LyricsManager.scss";

export default function LyricsManager() {
  const [input, setInput] = useState("");
  const [songs, setSongs] = useRecoilState(lyricsSongsState);
  const setDisplayPanelOpen = useSetRecoilState(displayPanelOpenState);
  const [loadingInfo, setLoadingInfo] = useState({ is: false, msg: "" });
  const [dedupResult, setDedupResult] = useState<Record<number, string>>({});

  const userInfo = useRecoilValue(userInfoState);

  const handler = {
    add: () => {
      const trimmed = input.trim();
      if (trimmed && !songs.some((s) => s.title === input)) {
        setSongs([...songs, { title: trimmed, lyrics: "", bpm: 100, expanded: false }]);
        setInput("");
      }
    },
    delete: (index: number) => {
      setSongs(songs.filter((_, i) => i !== index));
    },
    lyricsChange: (index: number, value: string) => {
      setSongs(songs.map((s, i) => i === index ? { ...s, lyrics: value } : s));
    },
    bpmChange: (index: number, value: number) => {
      setSongs(songs.map((s, i) => i === index ? { ...s, bpm: value } : s));
    },
    dedup: (index: number) => {
      const original = songs[index].lyrics;
      const deduped = deduplicateLines(original);
      setSongs(songs.map((s, i) => i === index ? { ...s, lyrics: deduped } : s));

      const changed = original !== deduped;
      const msg = changed ? "중복 병합 완료" : "중복 없음";
      setDedupResult((prev) => ({ ...prev, [index]: msg }));
      setTimeout(() => setDedupResult((prev) => {
        const next = { ...prev };
        delete next[index];
        return next;
      }), 2000);
    },
  };

  const toggleExpand = (index: number) => {
    setSongs(songs.map((s, i) => i === index ? { ...s, expanded: !s.expanded } : s));
  };


  // 중복 제거 — Go SplitTwoLines 방식 포팅
  // 빈 줄 구분 블록이면 블록 단위, 아니면 2줄 블록 단위로 연속 반복 감지 → (xN) 표기
  const deduplicateLines = (text: string): string => {
    const hasBlankLine = /\n\s*\n/.test(text);

    if (hasBlankLine) {
      // 블록 단위: 빈 줄로 나눈 뒤 연속 동일 블록 → (xN)
      const blocks: string[] = [];
      let cur: string[] = [];
      for (const line of text.split("\n")) {
        if (line.trim() === "") {
          if (cur.length > 0) { blocks.push(cur.join("\n")); cur = []; }
          continue;
        }
        cur.push(line);
      }
      if (cur.length > 0) blocks.push(cur.join("\n"));

      const result: string[] = [];
      let prev = "";
      let count = 1;
      for (const block of blocks) {
        if (block.trim() === prev.trim()) {
          count++;
        } else {
          if (count > 1 && result.length > 0) {
            result[result.length - 1] += `\n(x${count})`;
          }
          result.push(block);
          prev = block;
          count = 1;
        }
      }
      if (count > 1 && result.length > 0) {
        result[result.length - 1] += `\n(x${count})`;
      }
      return result.join("\n\n");
    }

    // 빈 줄 없음
    const lines = text.split("\n").filter((l) => l.trim() !== "");

    // 1단계: 단일 줄 연속 반복 → (xN)
    const step1: string[] = [];
    let cnt = 1;
    let hadDups = false;
    for (let i = 0; i < lines.length; i++) {
      if (i + 1 < lines.length && lines[i].trim() === lines[i + 1].trim()) {
        cnt++;
      } else {
        step1.push(lines[i]);
        if (cnt > 1) { step1.push(`(x${cnt})`); hadDups = true; }
        cnt = 1;
      }
    }
    if (hadDups) return step1.join("\n");

    // 2단계: 2줄 블록 단위 연속 반복 → (xN)
    const result: string[] = [];
    let prev = "";
    let count = 1;
    for (let i = 0; i < lines.length; i += 2) {
      const block = i + 1 < lines.length
        ? lines[i] + "\n" + lines[i + 1]
        : lines[i];
      if (block.trim() === prev.trim()) {
        count++;
      } else {
        if (count > 1 && result.length > 0) {
          result[result.length - 1] += `\n(x${count})`;
        }
        result.push(block);
        prev = block;
        count = 1;
      }
    }
    if (count > 1 && result.length > 0) {
      result[result.length - 1] += `\n(x${count})`;
    }
    return result.join("\n");
  };

  const handleSearchLyrics = async () => {
    try {
      setLoadingInfo({ is: true, msg: "전체 가사를 검색 중입니다..." });
      const response = await apiClient.searchLyrics(
        songs.map(({ title, lyrics }) => ({ title, lyrics }))
      );

      if (!response.ok) throw new Error("가사 검색 요청 실패");
      const data = await response.json();

      const updatedSongs = songs.map((song) => {
        const matched = data.searchedSongs.find(
          (s: { title: string }) => s.title === song.title
        );
        const raw = matched?.lyrics || song.lyrics;
        const lyrics = raw ? deduplicateLines(raw) : raw;
        return {
          ...song,
          lyrics,
          expanded: lyrics ? true : song.expanded,
        };
      });
      setSongs(updatedSongs);
    } catch (error) {
      console.error("에러:", error);
    } finally {
      setLoadingInfo({ is: false, msg: "" });
    }
  };

  const handleSubmitLyrics = async () => {
    try {
      setLoadingInfo({ is: true, msg: "가사 기반으로 PDF 생성중입니다..." });

      const response = await apiClient.submitLyrics({
        figmaInfo: userInfo.figmaInfo,
        mark: userInfo.english_name,
        songs: songs.map(({ title, lyrics }) => ({ title, lyrics })),
      });

      if (!response.ok) throw new Error("가사 제출 실패");

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);

      const a = document.createElement("a");
      a.href = url;
      if (songs.length > 1) {
        a.download = `${songs[0].title} 외 ${songs.length - 1}.zip`;
      } else {
        a.download = `${songs[0].title}.zip`;
      }
      document.body.appendChild(a);
      a.click();
      a.remove();
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error("가사 제출 중 에러:", error);
      alert("가사 제출 중 오류가 발생했습니다.");
    } finally {
      setLoadingInfo({ is: false, msg: "" });
    }
  };

  const handleSendToDisplay = async () => {
    const payload = songs
      .filter((s) => s.lyrics.trim())
      .map((s) => ({
        title: s.title,
        lyrics: s.lyrics,
        bpm: s.bpm || 100,
      }));

    if (payload.length === 0) {
      alert("가사가 입력된 곡이 없습니다.");
      return;
    }

    try {
      setLoadingInfo({ is: true, msg: "Display로 가사 전송 중..." });

      // append용 items (서버에서 lyrics_display 전처리)
      const appendItems = payload.map((s) => ({
        title: s.title,
        lyrics: s.lyrics,
        info: "lyrics_display",
        bpm: s.bpm,
      }));

      setDisplayPanelOpen(true);
      openDisplayWindow();

      const res = await apiClient.appendToDisplay(appendItems);
      if (!res.ok) throw new Error("전송 실패");
    } catch (error) {
      console.error("Display 전송 에러:", error);
      alert("Display 전송 중 오류가 발생했습니다.");
    } finally {
      setLoadingInfo({ is: false, msg: "" });
    }
  };

  return (
    <div className="lyrics_container">
      {loadingInfo.is && (
        <div className="loading_overlay">
          <div className="spinner" />
          <p>{loadingInfo.msg}</p>
        </div>
      )}
      <div className="search_wrap">
        <div className="search_row">
          <div className="input_group">
            <input
              type="text"
              placeholder="검색하세요."
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyUp={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  handler.add();
                }
              }}
            />
            <button onClick={handler.add}>+</button>
          </div>

          {songs.length > 0 && (
            <div className="btn_group">
              <button className="search_btn" onClick={handleSearchLyrics}>
                전체 가사 찾기 <span>⌕</span>
              </button>
              <button className="display_btn" onClick={handleSendToDisplay}>
                Display 전송
              </button>
              {songs.every((song) => song.lyrics.trim() !== "") && (
                <button
                  disabled={!userInfo.figmaInfo.key || !userInfo.figmaInfo.token}
                  className={classNames("submit_btn", {
                    disabled:
                      !userInfo.figmaInfo.key || !userInfo.figmaInfo.token,
                  })}
                  onClick={handleSubmitLyrics}
                >
                  가사 제출 <span>✓</span>
                </button>
              )}
            </div>
          )}
        </div>

        {songs.length > 0 && (
          <div className="notice">
            <span>가사를 입력할 경우 자동가사 찾기는 생략됩니다.</span>
            <span className="ref">
              가사 참조 사이트 :
              <a href="https://music.bugs.co.kr/" target="_blank">
                https://music.bugs.co.kr/
              </a>
            </span>
          </div>
        )}
      </div>

      <div className="tags">
        {songs.map((song, idx) => (
          <div key={idx} className="song_block">
            <div
              className={`tag_header${!song.expanded ? " collapsed" : ""}`}
              onClick={() => toggleExpand(idx)}
            >
              <span>
                {song.title}
                <span className="arrow">▼</span>
              </span>
              <div className="tag_header_right">
                <label className="bpm_label" onClick={(e) => e.stopPropagation()}>
                  BPM
                  <input
                    type="number"
                    min={40}
                    max={240}
                    value={song.bpm}
                    onChange={(e) => handler.bpmChange(idx, parseInt(e.target.value) || 100)}
                    className="bpm_input"
                  />
                </label>
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    handler.delete(idx);
                  }}
                >
                  ✕
                </button>
              </div>
            </div>
            {song.expanded && (
              <>
                <textarea
                  rows={5}
                  value={song.lyrics}
                  onChange={(e) => handler.lyricsChange(idx, e.target.value)}
                  placeholder="가사를 입력하세요..."
                  className="lyrics_box"
                />
                {song.lyrics.trim() && (
                  <div className="dedup_row">
                    <button
                      className="dedup_btn"
                      onClick={() => handler.dedup(idx)}
                    >
                      중복 제거
                    </button>
                    {dedupResult[idx] && (
                      <span className={`dedup_result ${dedupResult[idx] === "중복 없음" ? "none" : "merged"}`}>
                        {dedupResult[idx]}
                      </span>
                    )}
                  </div>
                )}
              </>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
