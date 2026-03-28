import React, { useEffect, useState } from "react";
import { WorshipOrderItem } from "@/types";
import { deleteNode, insertSiblingNode } from "@/lib/treeUtils";
import EditChildNews from "./EditChildNews";

export interface ChurchNewsProps {
  handleValueChange: (key: string, { newObj, newLead }: { newObj: string; newLead?: string }) => void;
  selectedDetail: WorshipOrderItem;
  setSelectedDetail: React.Dispatch<React.SetStateAction<WorshipOrderItem>>;
  setSelectedItems: React.Dispatch<React.SetStateAction<WorshipOrderItem[]>>;
}

const ChurchNews = ({ handleValueChange, selectedDetail, setSelectedDetail, setSelectedItems }: ChurchNewsProps) => {
  const [selectedChild, setSelectedChild] = useState<WorshipOrderItem | null>(null);
  const [expandedKeys, setExpandedKeys] = useState(new Set<string>());
  const [addContent, setAddContent] = useState<WorshipOrderItem | null>(null);

  const handleModifyChild = (action: "DELETE" | "ADD", childKey: string) => {
    switch (action) {
      case "DELETE":
        setSelectedDetail((prev) => {
          if (!prev) return prev;
          return { ...prev, children: deleteNode(prev.children || [], childKey) };
        });

        setSelectedItems((prevItems) => deleteNode(prevItems, childKey));
        break;
      case "ADD":
        const keys = childKey.split(".");
        const lastKey = parseInt(keys[keys.length - 1], 10);
        const newKey = `${keys.slice(0, -1).join(".")}.${lastKey + 1}`;

        const newChild: WorshipOrderItem = {
          key: newKey,
          title: "",
          info: "c-edit",
          obj: "",
          children: [
            {
              key: `${newKey}.0`,
              title: "-",
              info: "c-edit",
              obj: "",
            },
          ],
        };

        setAddContent(newChild);
    }
  };

  const handleAddNewItem = (argAdd: WorshipOrderItem) => {
    if (addContent) {
      const updatedChild = {
        ...addContent,
        title: argAdd.title,
        obj: argAdd.obj,
      };

      setSelectedDetail((prev) => {
        if (!prev) return prev;
        return { ...prev, children: insertSiblingNode(prev.children || [], updatedChild) };
      });

      setSelectedItems((prevItems) => insertSiblingNode(prevItems, updatedChild));

      setAddContent(null);
    }
  };

  const toggleExpand = (key: string) => {
    setExpandedKeys((prev) => {
      const newKeys = new Set(prev);
      newKeys.has(key) ? newKeys.delete(key) : newKeys.add(key);
      return newKeys;
    });
  };

  const renderNewsList = (newsList: WorshipOrderItem[], depth = 0) => {
    return newsList.map((news, i) => {
      const hue = 215;
      const saturation = depth === 0 ? 55 : 40;
      const lightness = depth === 0 ? 30 : Math.min(85, 45 + depth * 15);
      const backgroundColor = `hsl(${hue}, ${saturation}%, ${lightness}%)`;
      const depthLabel = depth === 0 ? "▪ " : "└ ";

      return (
        <div
          key={news.key}
          className={`news-tag-wrapper depth-${depth}`}
          style={{
            boxShadow: !depth ? "2px 3px rgba(0, 0, 0, 0.1)" : "none",
            border: !depth ? "1px solid #e5e5e5" : "none",
            margin: !depth ? "6px 0" : "none",
            borderRadius: "5px",
            paddingLeft: depth > 0 ? `${depth * 24}px` : undefined,
          }}>
          {news.title !== "-" && (
            <span
              className="tag"
              onClick={() => setSelectedChild(news)}
              style={{
                backgroundColor,
                color: lightness > 60 ? "#000" : "#fff",
                padding: "7px 14px",
                borderRadius: "5px",
                fontSize: depth === 0 ? "14px" : "13px",
              }}>
              {depthLabel}{news.title}
              <button
                className="delete-btn"
                onClick={(e) => {
                  e.stopPropagation();
                  if (!window.confirm(`'${news.title}' 소식을 삭제하시겠습니까?`)) return;
                  handleModifyChild("DELETE", news.key);
                }}>
                x
              </button>
            </span>
          )}

          {news.children && (
            <button
              className="expand-btn"
              onClick={(e) => {
                e.stopPropagation();
                toggleExpand(news.key);
              }}>
              {expandedKeys.has(news.key) ? "▼" : "◀"}
            </button>
          )}

          {i === newsList.length - 1 && (
            <span
              className="tag"
              style={{
                backgroundColor: "transparent",
                color: "rgb(130 130 130)",
                padding: "5px 10px",
                border: "1px dashed #ccc",
                borderRadius: "5px",
              }}>
              <button
                className="plus-btn"
                onClick={(e) => {
                  e.stopPropagation();
                  handleModifyChild("ADD", news.key);
                }}>
                +
              </button>
              추가
            </span>
          )}

          {news.children && expandedKeys.has(news.key) && (
            <div className="sub-news">{renderNewsList(news.children, depth + 1)}</div>
          )}
        </div>
      );
    });
  };

  return (
    <>
      <div className="church-news-container">{selectedDetail?.children && renderNewsList(selectedDetail.children)}</div>

      {addContent && (
        <div
          className="add-item-form"
          style={{
            backgroundColor: "#fff",
            padding: "16px",
            borderRadius: "8px",
            boxShadow: "0 4px 6px rgba(0, 0, 0, 0.1)",
            display: "flex",
            flexDirection: "column",
            gap: "5px",
          }}>
          <input
            type="text"
            placeholder="타이틀을 입력하세요"
            value={addContent.title}
            onChange={(e) =>
              setAddContent((prev) => ({
                ...prev!,
                title: e.target.value,
              }))
            }
            style={{
              width: "100%",
              padding: "3px",
              border: "1px solid #ccc",
              borderRadius: "3px",
              outline: "none",
              fontSize: "0.8rem",
            }}
          />
          <input
            placeholder="내용을 입력하세요"
            value={addContent.obj}
            onChange={(e) =>
              setAddContent((prev) => ({
                ...prev!,
                obj: e.target.value,
              }))
            }
            style={{
              width: "100%",
              height: "50px",
              padding: "3px",
              border: "1px solid #ccc",
              borderRadius: "3px",
              outline: "none",
              fontSize: "0.8rem",
            }}
          />
          <button
            onClick={() => handleAddNewItem(addContent)}
            style={{
              backgroundColor: "#007bff",
              color: "#fff",
              padding: "10px",
              borderRadius: "6px",
              fontSize: "16px",
              border: "none",
              cursor: "pointer",
              transition: "background 0.3s",
            }}
            onMouseOver={(e) => (e.currentTarget.style.backgroundColor = "#0056b3")}
            onMouseOut={(e) => (e.currentTarget.style.backgroundColor = "#007bff")}>
            항목 추가
          </button>
        </div>
      )}

      {selectedChild && (
        <EditChildNews
          selectedDetail={selectedDetail}
          selectedChild={selectedChild}
          setSelectedChild={setSelectedChild}
          handleValueChange={handleValueChange}
        />
      )}
    </>
  );
};

export default ChurchNews;
