import React, { useState } from "react";
import { WorshipOrderItem } from "../page";
import EditChildNews from "./EditChildNews";

export interface ChurchNewsProps {
  handleValueChange: (key: string, newObj: string) => void;
  selectedDetail: WorshipOrderItem;
  setSelectedDetail: React.Dispatch<React.SetStateAction<WorshipOrderItem>>;
  setSelectedItems: React.Dispatch<React.SetStateAction<WorshipOrderItem[]>>;
}

const ChurchNews = ({ handleValueChange, selectedDetail, setSelectedDetail, setSelectedItems }: ChurchNewsProps) => {
  const [selectedChild, setSelectedChild] = useState<WorshipOrderItem>(selectedDetail);
  const [expandedKeys, setExpandedKeys] = useState(new Set<string>());

  const handleDeleteChild = (childKey: string) => {
    // selectedDetail 업데이트
    setSelectedDetail((prev) => {
      if (!prev) return prev;

      const deleteRecursive = (items: WorshipOrderItem[]): WorshipOrderItem[] =>
        items
          .map((child) => {
            const newChild = { ...child };
            if (child.key === childKey) return null;

            if (child.children) {
              newChild.children = deleteRecursive(child.children);
            }
            return newChild;
          })
          .filter(Boolean) as WorshipOrderItem[];

      return { ...prev, children: deleteRecursive(prev.children || []) };
    });

    // selectedItems에서 삭제
    setSelectedItems((prevItems) => {
      const deleteItemRecursive = (items: WorshipOrderItem[]): WorshipOrderItem[] =>
        items
          .map((item) => {
            const newItem = { ...item };
            if (item.key === childKey) return null;

            if (item.children) {
              newItem.children = deleteItemRecursive(item.children);
            }
            return newItem;
          })
          .filter(Boolean) as WorshipOrderItem[];

      return deleteItemRecursive(prevItems);
    });
  };

  const handleSelectChild = (child: WorshipOrderItem) => {
    setSelectedChild(child);
  };

  const toggleExpand = (key: string) => {
    setExpandedKeys((prev) => {
      const newKeys = new Set(prev);
      newKeys.has(key) ? newKeys.delete(key) : newKeys.add(key);
      return newKeys;
    });
  };

  const renderNewsItem = (news: WorshipOrderItem) => (
    <div key={news.key} className="news-tag-wrapper">
      <span className="tag" onClick={() => handleSelectChild(news)}>
        {news.title}
        <button
          className="delete-btn"
          onClick={(e) => {
            e.stopPropagation();
            handleDeleteChild(news.key);
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

      {news.children && expandedKeys.has(news.key) && (
        <div className="sub-news">
          {news.children.map((child) => (
            <div key={child.key} className="news-tag-wrapper sub-item">
              <span className="tag" onClick={() => handleSelectChild(child)}>
                {child.title}
                <button
                  className="delete-btn"
                  onClick={(e) => {
                    e.stopPropagation();
                    handleDeleteChild(child.key);
                  }}>
                  x
                </button>
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  );

  return (
    <>
      <div className="church-news-container">{selectedDetail?.children?.map(renderNewsItem)}</div>

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
