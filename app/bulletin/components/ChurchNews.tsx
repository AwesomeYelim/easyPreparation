import React, { useState } from "react";
import { WorshipOrderItem } from "../page";
import EditChildNews from "./EditChildNews";

export interface ChurchNewsProps {
  handleValueChange: (key: string, { newObj, newLead }: { newObj: string; newLead?: string }) => void;
  selectedDetail: WorshipOrderItem;
  setSelectedDetail: React.Dispatch<React.SetStateAction<WorshipOrderItem>>;
  setSelectedItems: React.Dispatch<React.SetStateAction<WorshipOrderItem[]>>;
}

const ChurchNews = ({ handleValueChange, selectedDetail, setSelectedDetail, setSelectedItems }: ChurchNewsProps) => {
  const [selectedChild, setSelectedChild] = useState<WorshipOrderItem>(selectedDetail);
  const [expandedKeys, setExpandedKeys] = useState(new Set<string>());
  const [addContent, setAddContent] = useState<WorshipOrderItem | null>(null);

  const handleModifyChild = (action: "DELETE" | "PLUS", childKey: string) => {
    switch (action) {
      case "DELETE":
        setSelectedDetail((prev) => {
          if (!prev) return prev;

          const deleteRecursive = (items: WorshipOrderItem[]): WorshipOrderItem[] =>
            items
              .map((child) => {
                if (child.key === childKey) return null;
                return child.children ? { ...child, children: deleteRecursive(child.children) } : child;
              })
              .filter(Boolean) as WorshipOrderItem[];

          return { ...prev, children: deleteRecursive(prev.children || []) };
        });

        setSelectedItems((prevItems) => {
          const deleteItemRecursive = (items: WorshipOrderItem[]): WorshipOrderItem[] =>
            items
              .map((item) => {
                if (item.key === childKey) return null;
                return item.children ? { ...item, children: deleteItemRecursive(item.children) } : item;
              })
              .filter(Boolean) as WorshipOrderItem[];

          return deleteItemRecursive(prevItems);
        });
      case "PLUS":
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
              key: `${newKey}.1`,
              title: "sample",
              info: "c-edit",
              obj: "",
            },
          ],
        };

        setAddContent(newChild);
    }
  };

  const handleAddNewItem = (title: string, obj: string) => {
    if (addContent) {
      const updatedChild = { ...addContent, title, obj };
      setSelectedDetail((prev) => {
        if (!prev) return prev;

        const insertSibling = (items: WorshipOrderItem[]): WorshipOrderItem[] => {
          return items.flatMap((item) => {
            const keys = addContent.key.split(".");
            const lastKey = parseInt(keys[keys.length - 1], 10);
            const beforeKey = `${keys.slice(0, -1).join(".")}.${lastKey - 1}`;

            if (item.key === beforeKey) {
              return [item, updatedChild];
            } else if (item.children) {
              return [{ ...item, children: insertSibling(item.children) }];
            }
            return [item];
          });
        };

        return { ...prev, children: insertSibling(prev.children || []) };
      });

      setSelectedItems((prevItems) => {
        const insertSibling = (items: WorshipOrderItem[]): WorshipOrderItem[] => {
          return items.flatMap((item) => {
            const keys = addContent.key.split(".");
            const lastKey = parseInt(keys[keys.length - 1], 10);
            const beforeKey = `${keys.slice(0, -1).join(".")}.${lastKey - 1}`;

            if (item.key === beforeKey) {
              return [item, updatedChild];
            } else if (item.children) {
              return [{ ...item, children: insertSibling(item.children) }];
            }
            return [item];
          });
        };
        return insertSibling(prevItems);
      });

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
      const saturation = 50;
      const lightness = Math.min(90, 25 + depth * 10);
      const backgroundColor = `hsl(${hue}, ${saturation}%, ${lightness}%)`;

      return (
        <div
          key={news.key}
          className="news-tag-wrapper"
          style={{
            boxShadow: !depth ? "2px 3px rgba(0, 0, 0, 0.1)" : "none",
            border: !depth ? "1px solid #e5e5e5" : "none",
            margin: !depth ? "5px" : "none",
            borderRadius: "5px",
          }}>
          <span
            className="tag"
            onClick={() => setSelectedChild(news)}
            style={{
              backgroundColor,
              color: lightness > 60 ? "#000" : "#fff",
              padding: "5px 10px",
              borderRadius: "5px",
            }}>
            {news.title}
            <button
              className="delete-btn"
              onClick={(e) => {
                e.stopPropagation();
                handleModifyChild("DELETE", news.key);
              }}>
              x
            </button>
          </span>

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
                  handleModifyChild("PLUS", news.key);
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
            onClick={() => handleAddNewItem(addContent.title, addContent.obj)}
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
          selectedChild={selectedChild}
          setSelectedChild={setSelectedChild}
          handleValueChange={handleValueChange}
        />
      )}
    </>
  );
};

export default ChurchNews;
