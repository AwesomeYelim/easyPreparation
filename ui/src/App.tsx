import React, { useState } from "react";
import EditableData from "./components/contents.tsx";

const FigmaTokenInput: React.FC = () => {
  const [secureInfo, setSecureInfo] = useState({ token: "", key: "" });
  const [isSubmitted, setIsSubmitted] = useState(false);

  const handleSubmit = async () => {
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
      {!isSubmitted ? (
        <div>
          <h1>Figma Token and Key</h1>
          <div style={{ marginBottom: "20px" }}>
            <input
              type="text"
              value={secureInfo.token}
              onChange={(e) => setSecureInfo({ ...secureInfo, token: e.target.value })}
              placeholder="Enter your token"
              style={{
                margin: "10px 0",
                padding: "10px",
                width: "100%",
                maxWidth: "400px",
                border: "1px solid #ccc",
                borderRadius: "4px",
              }}
            />
            <input
              type="text"
              value={secureInfo.key}
              onChange={(e) => setSecureInfo({ ...secureInfo, key: e.target.value })}
              placeholder="Enter your key"
              style={{
                margin: "10px 0",
                padding: "10px",
                width: "100%",
                maxWidth: "400px",
                border: "1px solid #ccc",
                borderRadius: "4px",
              }}
            />
            <button
              onClick={handleSubmit}
              style={{
                padding: "10px 20px",
                backgroundColor: "#4CAF50",
                color: "white",
                border: "none",
                borderRadius: "4px",
                cursor: "pointer",
              }}>
              Submit
            </button>
          </div>
        </div>
      ) : (
        <div>
          <h1>Editable Data</h1>
          <EditableData />
        </div>
      )}
    </div>
  );
};

export default FigmaTokenInput;
