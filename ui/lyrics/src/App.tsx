import React, { useState } from "react";
import FigmaTokenForm from "./components/tokenKey.tsx";
const SongTitleInput: React.FC = () => {
  const [aboutLyric, setAboutLyric] = useState({
    songTitle: "",
    label: "",
  });

  const handleSubmit = async (secureInfo) => {
    if (!secureInfo.token || !secureInfo.key) {
      alert("Both token and key must be provided.");
      return;
    }

    if (!aboutLyric.songTitle) {
      alert("노래 제목을 입력하세요.");
      return;
    }

    try {
      if (window.sendTokenAndKey) {
        await window.sendTokenAndKey(secureInfo);
      }
      if (window.sendLyrics) {
        await window.sendLyrics(aboutLyric);
      }
    } catch (error) {
      console.error("Error:", error);
    }
  };

  return (
    <>
      <div>
        <h1>교회 이름 입력</h1>
        <input
          type="text"
          value={aboutLyric.label}
          onChange={(e) => setAboutLyric({ ...aboutLyric, label: e.target.value })}
          placeholder="교회 이름을 입력하세요. (ppt label 용)"
          style={{ padding: "10px", width: "300px", marginBottom: "10px" }}
        />
      </div>
      <div>
        <h1>노래 제목 / 가사 입력</h1>
        <p style={{ color: "#ccc" }}>여러개 입력시 (,) 로 구분</p>
        <input
          type="text"
          value={aboutLyric.songTitle}
          onChange={(e) => setAboutLyric({ ...aboutLyric, songTitle: e.target.value })}
          placeholder="노래 제목(또는 가사)을 입력하세요"
          style={{ padding: "10px", width: "300px", marginBottom: "10px" }}
        />
      </div>
      <FigmaTokenForm onSubmit={handleSubmit} />
    </>
  );
};

export default SongTitleInput;
