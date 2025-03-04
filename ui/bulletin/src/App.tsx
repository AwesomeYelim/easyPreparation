import React, { useState } from "react";
import EditableData from "./components/contents.tsx";
import FigmaTokenForm from "./components/tokenKey.tsx";

const FigmaTokenInput: React.FC = () => {
  const [isSubmitted, setIsSubmitted] = useState(false);

  const handleSubmit = async (secureInfo) => {
    if (!secureInfo.token || !secureInfo.key) {
      alert("Both token and key must be provided.");
      return;
    }

    try {
      // 실제 Go와 연동 시 사용될 함수 호출
      // Go로 데이터를 보내기 위해 window.sendTokenAndKey 사용
      if (window.sendTokenAndKey) {
        await window.sendTokenAndKey(secureInfo);
        setIsSubmitted(true); // UI 표시 상태 업데이트
      }
    } catch (error) {
      console.error("Error:", error);
    }
  };

  return (
    <div style={{ margin: "20px", fontFamily: "Arial, sans-serif" }}>
      {/* {!isSubmitted ? (
        <FigmaTokenForm onSubmit={handleSubmit} />
      ) : ( */}
      <div>
        <h1>Editable Data</h1>
        <EditableData />
      </div>
      {/* )} */}
    </div>
  );
};

export default FigmaTokenInput;
