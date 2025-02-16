// EditableData.tsx
import React, { useState } from "react";
import BibleSelect from "./bibleSelect.tsx";
import initialData from "../data.json"; // JSON 파일 import

interface Info {
  title: string;
  obj: string;
  info: string;
  lead?: string;
  children?: Info[];
}

const EditableData: React.FC = () => {
  const [title, setTitle] = useState("main_worship");
  const [data, setData] = useState(initialData);

  const handleValueChange = (key: string, newObj: string) => {
    const updateData = (items: Info[], keyParts: string[]): Info[] => {
      const [currentIndex, ...restKeyParts] = keyParts;
      if (!currentIndex) return items;
      return items.map((item, index) => {
        if (index === parseInt(currentIndex)) {
          if (restKeyParts.length === 0) {
            switch (item.info) {
              case "b_edit":
                return { ...item, obj: newObj };
              case "c_edit":
                return { ...item, obj: newObj };
              case "r_edit":
                return { ...item, lead: newObj };
            }
          }
          if (item.children) {
            return {
              ...item,
              children: updateData(item.children, restKeyParts),
            };
          }
        }
        return item;
      });
    };

    setData((prevData) => updateData(prevData, key.split("-")));
  };

  const handleSubmit = async () => {
    try {
      if (window.sendContentsDate) {
        await window.sendContentsDate(title, data);
      }
    } catch (error) {
      console.error("Error:", error);
    }
  };

  const renderItems = (items: Info[], parentIndex: string = "") => {
    return items.map((item, index) => {
      const key = parentIndex ? `${parentIndex}-${index}` : `${index}`;
      if (!item.info.includes("-")) {
        return (
          <div
            key={key}
            style={{
              marginBottom: "15px",
              display: "block",
              maxWidth: 400,
              flexWrap: "wrap",
              justifyContent: "start",
            }}
          >
            <label
              style={{
                marginTop: "10px",
                color: item.info.includes("edit") ? "#000" : "#ccc",
              }}
            >
              {item.title}
            </label>
            {item.info.includes("edit") && !item.info.includes("b_") && item.obj != "" && (
              <input
                type="text"
                onChange={(e) => handleValueChange(key, e.target.value)}
                placeholder={item.info == "r_edit" ? item.lead : item.obj}
                style={{
                  marginTop: "5px",
                  padding: "10px",
                  width: "100%",
                  maxWidth: "400px",
                  border: "1px solid #ccc",
                  borderRadius: "4px",
                }}
              />
            )}
            {item.info.includes("edit") && item.info.includes("b_") && (
              <BibleSelect
                handleValueChange={handleValueChange}
                parentKey={key}
              />
            )}
            {item.children && (
              <div style={{ marginLeft: "20px", marginTop: "10px" }}>
                {renderItems(item.children, key)}
              </div>
            )}
          </div>
        );
      }
      return <></>;
    });
  };

  return (
    <div>
      <div style={{ marginBottom: "15px" }}>
        <label>예배 종류</label>
        <input
          type="text"
          onChange={(e) => setTitle(e.target.value)}
          placeholder="main_worship"
          style={{
            display: "block",
            marginTop: "10px",
            padding: "10px",
            width: "100%",
            maxWidth: "400px",
            border: "1px solid #ccc",
            borderRadius: "4px",
          }}
        />
      </div>
      {renderItems(data)}
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
