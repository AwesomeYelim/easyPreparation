"use client";

import { useState } from "react";
import { useRecoilValue, useRecoilState, useSetRecoilState } from "recoil";
import { userInfoState, lyricsSongsState, displayPanelOpenState, userSettingsState } from "@/recoilState";
import { apiClient, openDisplayWindow } from "@/lib/apiClient";
import toast from "react-hot-toast";

export default function LyricsManager() {
  const [input, setInput] = useState("");
  const [songs, setSongs] = useRecoilState(lyricsSongsState);
  const setDisplayPanelOpen = useSetRecoilState(displayPanelOpenState);
  const [loadingInfo, setLoadingInfo] = useState({ is: false, msg: "" });
  const [dedupResult, setDedupResult] = useState<Record<number, string>>({});

  const userInfo = useRecoilValue(userInfoState);
  const settings = useRecoilValue(userSettingsState);

  const handler = {
    add: () => {
      const trimmed = input.trim();
      if (trimmed && !songs.some((s) => s.title === input)) {
        setSongs([...songs, { title: trimmed, lyrics: "", bpm: settings.default_bpm || 100, expanded: false }]);
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
        mark: userInfo.english_name,
        songs: songs.map(({ title, lyrics }) => ({ title, lyrics })),
        email: userInfo.email,
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
      toast.error("가사 제출 중 오류가 발생했습니다.");
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
      toast("가사가 입력된 곡이 없습니다.");
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

      const res = await apiClient.appendToDisplay(appendItems, "lyrics");
      if (!res.ok) throw new Error("전송 실패");
    } catch (error) {
      console.error("Display 전송 에러:", error);
      toast.error("Display 전송 중 오류가 발생했습니다.");
    } finally {
      setLoadingInfo({ is: false, msg: "" });
    }
  };

  return (
    <div className="p-6 flex flex-col gap-6">
      {/* 로딩 오버레이 */}
      {loadingInfo.is && (
        <div className="fixed inset-0 bg-black/50 flex flex-col justify-center items-center z-[9999]">
          <div className="w-14 h-14 border-4 border-white/30 border-t-white rounded-full animate-spin mb-5" />
          <p className="text-yellow-300 text-base font-medium">{loadingInfo.msg}</p>
        </div>
      )}

      {/* 검색 섹션 */}
      <div className="flex flex-col gap-3">
        {/* 입력 + 버튼 행 */}
        <div className="flex flex-wrap items-center gap-3">
          {/* 입력 그룹 */}
          <div className="flex items-stretch w-80 max-w-full">
            <input
              type="text"
              placeholder="곡 제목을 입력하세요 (예: 은혜 아니면)"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyUp={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  handler.add();
                }
              }}
              className="flex-1 px-4 py-2.5 text-sm border border-outline/50 border-r-0 rounded-l-lg bg-surface focus:outline-none focus:ring-1 focus:ring-secondary focus:border-secondary transition-all text-on-surface placeholder:text-on-surface-variant"
            />
            <button
              onClick={handler.add}
              title="곡 추가"
              className="px-4 bg-secondary text-white text-xl font-bold rounded-r-lg border border-secondary hover:bg-secondary/90 active:scale-95 transition-all leading-none flex items-center"
            >
              +
            </button>
          </div>

          {/* 액션 버튼들 */}
          {songs.length > 0 && (
            <div className="flex gap-2 flex-wrap">
              <button
                className="flex items-center gap-1.5 px-4 py-2.5 bg-surface-high text-primary text-sm font-bold rounded-lg border border-outline/30 hover:bg-surface-highest transition-all shadow-sm"
                onClick={handleSearchLyrics}
              >
                전체 가사 찾기
                <span className="text-base">⌕</span>
              </button>
              <button
                className="flex items-center gap-1.5 px-4 py-2.5 bg-secondary text-white text-sm font-bold rounded-lg hover:bg-secondary/90 active:scale-95 transition-all shadow-sm"
                onClick={handleSendToDisplay}
              >
                Display 전송
              </button>
              {songs.every((song) => song.lyrics.trim() !== "") && (
                <button
                  className="flex items-center gap-1.5 px-4 py-2.5 bg-primary text-white text-sm font-bold rounded-lg hover:bg-primary/80 active:scale-95 transition-all shadow-sm"
                  onClick={handleSubmitLyrics}
                >
                  가사 제출
                  <span>✓</span>
                </button>
              )}
            </div>
          )}
        </div>

        {/* 안내 문구 */}
        {songs.length > 0 && (
          <div className="flex flex-wrap gap-3 text-xs text-on-surface-variant">
            <span>가사를 입력할 경우 자동가사 찾기는 생략됩니다.</span>
            <span>
              가사 참조 사이트:{" "}
              <a
                href="https://music.bugs.co.kr/"
                target="_blank"
                className="text-secondary hover:underline"
              >
                https://music.bugs.co.kr/
              </a>
            </span>
          </div>
        )}
      </div>

      {/* 곡 목록 */}
      <div className="flex flex-wrap gap-4">
        {songs.length === 0 && (
          <div className="w-full text-center py-10 text-on-surface-variant">
            <p className="text-sm">아직 추가된 곡이 없습니다.</p>
            <p className="text-xs mt-2">위 검색창에서 곡 제목을 입력해 추가하세요.</p>
          </div>
        )}
        {songs.map((song, idx) => (
          <div
            key={idx}
            className="min-w-[280px] flex-1 basis-[300px] max-w-[500px] h-fit bg-surface-low border border-outline/30 rounded-xl shadow-sm transition-all"
          >
            {/* 곡 헤더 */}
            <div
              className={`flex justify-between items-center px-4 py-3 cursor-pointer select-none rounded-xl ${
                song.expanded ? "rounded-b-none border-b border-outline/20" : ""
              }`}
              onClick={() => toggleExpand(idx)}
            >
              <span className="flex items-center gap-2 font-semibold text-[15px] text-on-surface">
                <span
                  className={`text-sm text-on-surface-variant transition-transform duration-200 ${
                    song.expanded ? "rotate-0" : "-rotate-90"
                  }`}
                >
                  ▼
                </span>
                {song.title}
              </span>
              <div className="flex items-center gap-2">
                <label
                  className="flex items-center gap-1.5 text-xs font-medium text-on-surface-variant cursor-default"
                  onClick={(e) => e.stopPropagation()}
                >
                  BPM
                  <input
                    type="number"
                    min={40}
                    max={240}
                    value={song.bpm}
                    onChange={(e) => handler.bpmChange(idx, parseInt(e.target.value) || 100)}
                    className="w-14 px-1.5 py-1 text-center text-sm border border-outline/40 rounded bg-white focus:outline-none focus:ring-1 focus:ring-secondary"
                  />
                </label>
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    handler.delete(idx);
                  }}
                  className="text-error font-bold text-sm px-1 hover:text-error/70 transition-colors"
                >
                  ✕
                </button>
              </div>
            </div>

            {/* 가사 영역 */}
            {song.expanded && (
              <div className="px-4 pb-4 flex flex-col gap-2">
                <textarea
                  rows={5}
                  value={song.lyrics}
                  onChange={(e) => handler.lyricsChange(idx, e.target.value)}
                  placeholder="가사를 입력하세요..."
                  className="w-full mt-3 px-3 py-2.5 text-sm font-sans border border-outline/40 rounded-lg bg-white text-on-surface placeholder:text-on-surface-variant focus:outline-none focus:ring-1 focus:ring-secondary resize-none transition-all"
                />
                {song.lyrics.trim() && (
                  <div className="flex items-center gap-2">
                    <button
                      className="px-3 py-1.5 text-xs font-semibold bg-electric-blue/10 text-secondary border border-secondary/20 rounded-md hover:bg-electric-blue/20 transition-colors"
                      onClick={() => handler.dedup(idx)}
                    >
                      중복 제거
                    </button>
                    {dedupResult[idx] && (
                      <span
                        className={`text-xs font-semibold animate-pulse ${
                          dedupResult[idx] === "중복 없음"
                            ? "text-on-surface-variant"
                            : "text-secondary"
                        }`}
                      >
                        {dedupResult[idx]}
                      </span>
                    )}
                  </div>
                )}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
