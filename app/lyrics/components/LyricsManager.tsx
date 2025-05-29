"use client";

import { useState } from "react";
import "./LyricsManager.scss";
import { useRecoilValue } from "recoil";
import { userInfoState } from "@/recoilState";

interface SongBlock {
  title: string;
  lyrics: string;
  expanded: boolean;
}

export default function LyricsManager() {
  const [input, setInput] = useState("");
  const [songs, setSongs] = useState<SongBlock[]>([]);
  const userInfo = useRecoilValue(userInfoState);

  const handler = {
    add: () => {
      const trimmed = input.trim();
      if (trimmed && !songs.some((s) => s.title === input)) {
        setSongs([...songs, { title: trimmed, lyrics: "", expanded: false }]);
        setInput("");
      }
    },
    delete: (index: number) => {
      setSongs(songs.filter((_, i) => i !== index));
    },
    lyricsChange: (index: number, value: string) => {
      const newSongs = [...songs];
      newSongs[index].lyrics = value;
      setSongs(newSongs);
    },
  };

  const toggleExpand = (index: number) => {
    const newSongs = [...songs];
    newSongs[index].expanded = !newSongs[index].expanded;
    setSongs(newSongs);
  };

  const handleSearchLyrics = async () => {
    try {
      const baseUrl = process.env.NEXT_PUBLIC_API_BASE_URL;
      const response = await fetch(`${baseUrl}/searchLyrics`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          songs: songs.map(({ title, lyrics }) => ({ title, lyrics })),
        }),
      });

      if (!response.ok) throw new Error("가사 검색 요청 실패");
      const data = await response.json();

      const updatedSongs = songs.map((song) => {
        const matched = data.searchedSongs.find(
          (s: { title: string }) => s.title === song.title
        );
        return {
          ...song,
          lyrics: matched?.lyrics || song.lyrics,
        };
      });
      setSongs(updatedSongs);
    } catch (error) {
      console.error("에러:", error);
    }
  };

  const handleSubmitLyrics = async () => {
    try {
      const baseUrl = process.env.NEXT_PUBLIC_API_BASE_URL;
      const response = await fetch(`${baseUrl}/submitLyrics`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          figmaInfo: userInfo.figmaInfo,
          mark: userInfo.english_name,
          songs: songs.map(({ title, lyrics }) => ({ title, lyrics })),
        }),
      });

      if (!response.ok) throw new Error("가사 제출 실패");

      alert("가사가 성공적으로 제출되었습니다!");
    } catch (error) {
      console.error("가사 제출 중 에러:", error);
      alert("가사 제출 중 오류가 발생했습니다.");
    }
  };

  return (
    <div className="container">
      <div className="search_wrap">
        <div className="inputGroup">
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

        {songs.length > 0 && (
          <div className="btnGroup">
            <button className="searchBtn" onClick={handleSearchLyrics}>
              전체 가사 찾기 <span>⌕</span>
            </button>

            {songs.every((song) => song.lyrics.trim() !== "") && (
              <button className="submitBtn" onClick={handleSubmitLyrics}>
                가사 제출 <span>✓</span>
              </button>
            )}
          </div>
        )}
      </div>

      <div className="tags">
        {songs.map((song, idx) => (
          <div key={idx} className="songBlock">
            <div
              className={`tagHeader${!song.expanded ? " collapsed" : ""}`}
              onClick={() => toggleExpand(idx)}
            >
              <span>
                {song.title}
                <span className="arrow">▼</span>
              </span>
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  handler.delete(idx);
                }}
              >
                ✕
              </button>
            </div>
            {song.expanded && (
              <textarea
                rows={5}
                value={song.lyrics}
                onChange={(e) => handler.lyricsChange(idx, e.target.value)}
                placeholder="가사를 입력하세요..."
                className="lyricsBox"
              />
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
