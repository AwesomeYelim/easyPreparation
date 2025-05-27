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
  //   const userInfo = useRecoilValue(userInfoState);

  const handler = {
    add: () => {
      console.log(input);

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
          //   figmaInfo: userInfo.figmaInfo,
          //   mark: userInfo.english_name,
          songs: songs.map(({ title, lyrics }) => ({ title, lyrics })),
        }),
      });

      if (!response.ok) throw new Error("가사 검색 요청 실패");
      const data = await response.json();
      console.log("가사 검색 결과:", data);

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
            <span className="notice">
              가사를 입력할 경우 자동가사 찾기는 생략됩니다.
            </span>
          )}
        </div>

        {songs.length > 0 && (
          <button className="searchBtn" onClick={handleSearchLyrics}>
            전체 가사 찾기 <span>⌕</span>
          </button>
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
