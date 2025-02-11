import React, { useState } from "react";
import FigmaTokenForm from "./components/tokenKey.tsx";
const SongTitleInput: React.FC = () => {
  const [songTitle, setSongTitle] = useState("");

  const handleSubmit = async (secureInfo) => {
    if (!secureInfo.token || !secureInfo.key) {
      alert("Both token and key must be provided.");
      return;
    }

    if (!songTitle) {
      alert("노래 제목을 입력하세요.");
      return;
    }

    try {
      if (window.sendTokenAndKey) {
        await window.sendTokenAndKey(secureInfo);
      }
      if (window.sendSongTitle) {
        await window.sendSongTitle(songTitle);
      }
    } catch (error) {
      console.error("Error:", error);
    }
  };

  return (
    <>
      <div>
        <h1>노래 제목 입력</h1>
        <input
          type="text"
          value={songTitle}
          onChange={(e) => setSongTitle(e.target.value)}
          placeholder="노래 제목을 입력하세요"
          style={{ padding: "10px", width: "300px", marginBottom: "10px" }}
        />
      </div>
      <FigmaTokenForm onSubmit={handleSubmit} />
    </>
  );
};

export default SongTitleInput;
