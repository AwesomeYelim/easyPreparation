import React, { useState } from "react";
import { WorshipOrderItem } from "../page";
import EditChildNews from "./EditChildNews";

export interface ChurchNewsProps {
  handleValueChange: (key: string, newObj: string) => void;
  selectedDetail: WorshipOrderItem;
  setSelectedDetail: React.Dispatch<React.SetStateAction<WorshipOrderItem>>;
  setSelectedItems: React.Dispatch<React.SetStateAction<WorshipOrderItem[]>>;
}

const ChurchNews = ({
  handleValueChange,
  selectedDetail,
  setSelectedDetail,
  setSelectedItems,
}: ChurchNewsProps) => {
  const [selectedChild, setSelectedChild] =
    useState<WorshipOrderItem>(selectedDetail);
  const [expandedKeys, setExpandedKeys] = useState(new Set<string>());
  const [addContent, setAddContent] = useState<WorshipOrderItem | null>(null);

  const handleModifyChild = (action: "DELETE" | "PLUS", childKey: string) => {
    if (action === "DELETE") {
      // 삭제 로직
      setSelectedDetail((prev) => {
        if (!prev) return prev;

        const deleteRecursive = (
          items: WorshipOrderItem[]
        ): WorshipOrderItem[] =>
          items
            .map((child) => {
              if (child.key === childKey) return null;
              return child.children
                ? { ...child, children: deleteRecursive(child.children) }
                : child;
            })
            .filter(Boolean) as WorshipOrderItem[];

        return { ...prev, children: deleteRecursive(prev.children || []) };
      });

      setSelectedItems((prevItems) => {
        const deleteItemRecursive = (
          items: WorshipOrderItem[]
        ): WorshipOrderItem[] =>
          items
            .map((item) => {
              if (item.key === childKey) return null;
              return item.children
                ? { ...item, children: deleteItemRecursive(item.children) }
                : item;
            })
            .filter(Boolean) as WorshipOrderItem[];

        return deleteItemRecursive(prevItems);
      });
    } else if (action === "PLUS") {
      const keys = childKey.split(".");
      const lastKey = parseInt(keys[keys.length - 1], 10);
      const newKey = `${keys.slice(0, -1).join(".")}.${lastKey + 1}`;

      const newChild: WorshipOrderItem = {
        key: newKey,
        title: "새로운 항목",
        info: "c-edit",
        obj: "",
      };

      setAddContent(newChild);
    }
  };

  const handleAddNewItem = (title: string, obj: string) => {
    if (addContent) {
      const updatedChild = { ...addContent, title, obj };

      setSelectedDetail((prev) => {
        if (!prev) return prev;

        const insertSibling = (
          items: WorshipOrderItem[]
        ): WorshipOrderItem[] => {
          return items.map((item) => {
            const keys = addContent.key.split(".");
            const lastKey = parseInt(keys[keys.length - 1], 10);
            const beforeKey = `${keys.slice(0, -1).join(".")}.${lastKey - 1}`;

            if (item.key === beforeKey) {
              return {
                ...item,
                children: [...(item.children || []), updatedChild],
              };
            } else if (item.children) {
              return { ...item, children: insertSibling(item.children) };
            }
            return item;
          });
        };

        return { ...prev, children: insertSibling(prev.children || []) };
      });

      setSelectedItems((prevItems) => {
        const insertSibling = (
          items: WorshipOrderItem[]
        ): WorshipOrderItem[] => {
          return items.map((item) => {
            const keys = addContent.key.split(".");
            const lastKey = parseInt(keys[keys.length - 1], 10);
            const beforeKey = `${keys.slice(0, -1).join(".")}.${lastKey - 1}`;

            if (item.key === beforeKey) {
              return {
                ...item,
                children: [...(item.children || []), updatedChild],
              };
            } else if (item.children) {
              return { ...item, children: insertSibling(item.children) };
            }
            return item;
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
        <div key={news.key} className="news-tag-wrapper">
          <span
            className="tag"
            onClick={() => setSelectedChild(news)}
            style={{
              backgroundColor,
              color: lightness > 60 ? "#000" : "#fff",
              padding: "5px 10px",
              borderRadius: "5px",
            }}
          >
            {news.title}
            <button
              className="delete-btn"
              onClick={(e) => {
                e.stopPropagation();
                handleModifyChild("DELETE", news.key);
              }}
            >
              x
            </button>
          </span>

          {news.children && (
            <button
              className="expand-btn"
              onClick={(e) => {
                e.stopPropagation();
                toggleExpand(news.key);
              }}
            >
              {expandedKeys.has(news.key) ? "▼" : "◀"}
            </button>
          )}

          {i === newsList.length - 1 && (
            <span
              className="tag"
              style={{
                backgroundColor,
                padding: "5px 10px",
                borderRadius: "5px",
              }}
            >
              <button
                className="plus-btn"
                onClick={(e) => {
                  e.stopPropagation();
                  handleModifyChild("PLUS", news.key);
                }}
              >
                +
              </button>
              추가
            </span>
          )}

          {news.children && expandedKeys.has(news.key) && (
            <div className="sub-news">
              {renderNewsList(news.children, depth + 1)}
            </div>
          )}
        </div>
      );
    });
  };

  return (
    <>
      <div className="church-news-container">
        {selectedDetail?.children && renderNewsList(selectedDetail.children)}
      </div>

      {addContent && (
        <div className="add-item-form">
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
          />
          <textarea
            placeholder="내용을 입력하세요"
            value={addContent.obj}
            onChange={(e) =>
              setAddContent((prev) => ({
                ...prev!,
                obj: e.target.value,
              }))
            }
          />
          <button
            onClick={() => handleAddNewItem(addContent.title, addContent.obj)}
          >
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
