import React, { useState } from "react";
import { WorshipOrderItem } from "@/types";
import { deleteNode, insertSiblingNode } from "@/lib/treeUtils";
import ConfirmModal from "@/components/ConfirmModal";
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
  const [confirmDelete, setConfirmDelete] = useState<{ key: string; title: string } | null>(null);

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
          className={`depth-${depth}`}
          style={{
            boxShadow: !depth ? "2px 3px rgba(0, 0, 0, 0.1)" : "none",
            border: !depth ? "1px solid #e5e5e5" : "none",
            margin: !depth ? "6px 0" : undefined,
            marginTop: depth > 0 ? "6px" : undefined,
            borderRadius: "8px",
            paddingLeft: depth > 0 ? `${depth * 24}px` : undefined,
          }}>
          {news.title !== "-" && (
            <span
              className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg cursor-pointer"
              onClick={() => setSelectedChild(news)}
              style={{
                backgroundColor,
                color: lightness > 60 ? "#000" : "#fff",
                padding: "7px 14px",
                borderRadius: "8px",
                fontSize: depth === 0 ? "14px" : "13px",
              }}>
              {depthLabel}{news.title}
              <button
                className="w-4 h-4 flex items-center justify-center rounded-full bg-white/20 text-white text-[9px] font-black hover:bg-white/40 transition-colors ml-1"
                onClick={(e) => {
                  e.stopPropagation();
                  setConfirmDelete({ key: news.key, title: news.title });
                }}>
                ×
              </button>
            </span>
          )}

          {news.children && (
            <button
              className="text-electric-blue hover:text-secondary transition-colors text-lg px-1 border-none bg-transparent cursor-pointer"
              onClick={(e) => {
                e.stopPropagation();
                toggleExpand(news.key);
              }}>
              {expandedKeys.has(news.key) ? "▼" : "◀"}
            </button>
          )}

          {i === newsList.length - 1 && (
            <span
              className="inline-flex items-center gap-1 px-3 py-1.5 rounded-lg cursor-pointer border border-dashed border-slate-300 text-slate-400 text-sm ml-1"
              style={{ backgroundColor: "transparent" }}>
              <button
                className="w-4 h-4 flex items-center justify-center rounded-full bg-slate-600 text-white text-[10px] font-black hover:bg-navy-dark transition-colors"
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
            <div className="pl-6 mt-1">{renderNewsList(news.children, depth + 1)}</div>
          )}
        </div>
      );
    });
  };

  return (
    <>
      <div className="flex flex-col gap-3">
        {selectedDetail?.children && renderNewsList(selectedDetail.children)}
      </div>

      {addContent && (
        <div className="mt-4 bg-surface-low p-4 rounded-xl border border-slate-200 flex flex-col gap-3">
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
            className="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm text-navy-dark focus:outline-none focus:border-electric-blue focus:ring-2 focus:ring-electric-blue/20 transition-all"
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
            className="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm text-navy-dark h-12 focus:outline-none focus:border-electric-blue focus:ring-2 focus:ring-electric-blue/20 transition-all"
          />
          <button
            onClick={() => handleAddNewItem(addContent)}
            className="w-full bg-electric-blue text-white py-2.5 rounded-xl font-bold text-sm hover:bg-secondary transition-all active:scale-[0.98] shadow-sm shadow-electric-blue/20"
          >
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

      <ConfirmModal
        open={confirmDelete !== null}
        message={`'${confirmDelete?.title}' 소식을 삭제하시겠습니까?`}
        confirmLabel="삭제"
        danger
        onConfirm={() => {
          if (confirmDelete) handleModifyChild("DELETE", confirmDelete.key);
          setConfirmDelete(null);
        }}
        onCancel={() => setConfirmDelete(null)}
      />
    </>
  );
};

export default ChurchNews;
