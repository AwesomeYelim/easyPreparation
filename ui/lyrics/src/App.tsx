import React, { useState } from "react";

const SongTitleInput: React.FC = () => {
  const [songTitle, setSongTitle] = useState("");

  const handleSubmit = async () => {
    if (!songTitle) {
      alert("노래 제목을 입력하세요.");
      return;
    }

    try {
      // 실제 Go와 연동 시 사용될 함수 호출
      // Go로 데이터를 보내기 위해 window.sendSongTitle 사용
      if (window.sendSongTitle) {
        await window.sendSongTitle(songTitle);
      }
    } catch (error) {
      console.error("Error:", error);
    }
  };

  return (
    <div>
      <h1>노래 제목 입력</h1>
      <input
        type="text"
        value={songTitle}
        onChange={(e) => setSongTitle(e.target.value)}
        placeholder="노래 제목을 입력하세요"
        style={{ padding: "10px", width: "300px", marginBottom: "10px" }}
      />
      <button onClick={handleSubmit} style={{ padding: "10px 20px" }}>
        제출
      </button>
    </div>
  );
};

export default SongTitleInput;
