import React from "react";
import initialData from "../App.tsx";

const EditableData: React.FC<{
  data: typeof initialData;
  onChange: (index: number, newContent: string) => void;
}> = ({ data, onChange }) => {


  const handleSubmit = async () => {

    try {
      // 실제 Go와 연동 시 사용될 함수 호출
      // Go로 데이터를 보내기 위해 window.sendTokenAndKey 사용
      if (window.sendContentsDate) {
        await window.sendContentsDate(data)
      }
    } catch (error) {
      console.error("Error:", error);
    }
  };



  return (
    <div>
      {data.map((item, index) => (
        <div key={index} style={{ marginBottom: "15px" }}>
          <label>{item.title}</label>
          <input
            type="text"
            onChange={(e) => onChange(index, e.target.value)}
            placeholder={item.content}
            style={{
              display: "block",
              marginTop: "5px",
              padding: "10px",
              width: "100%",
              maxWidth: "400px",
              border: "1px solid #ccc",
              borderRadius: "4px",
            }}
          />
        </div>
      ))}

      <button
        onClick={handleSubmit}
        style={{
          padding: "10px 20px",
          backgroundColor: "#4CAF50",
          color: "white",
          border: "none",
          borderRadius: "4px",
          cursor: "pointer",
        }}
      >
        Submit
      </button>
    </div>
  );
};

export default EditableData;
