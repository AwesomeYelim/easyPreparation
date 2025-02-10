import React, { useState } from "react";

interface FigmaTokenFormProps {
  onSubmit: (token: string, key: string) => void;
}

const FigmaTokenForm: React.FC<FigmaTokenFormProps> = ({ onSubmit }) => {
  const [secureInfo, setSecureInfo] = useState({ token: "", key: "" });

  const handleSubmit = () => {
    if (!secureInfo.token || !secureInfo.key) {
      alert("Both token and key must be provided.");
      return;
    }

    // 부모 컴포넌트에서 전달받은 onSubmit 함수 호출
    onSubmit(secureInfo);
  };

  return (
    <div style={{ marginBottom: "20px" }}>
      <h1>Figma Token and Key</h1>
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
  );
};

export default FigmaTokenForm;
