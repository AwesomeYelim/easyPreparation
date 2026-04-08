import React, { useState } from "react";
import { useRecoilState } from "recoil";
import { selectedDetailState } from "@/recoilState";
import { WorshipOrderItem } from "@/types";
import ConfirmModal from "@/components/ConfirmModal";

export default function SelectedOrder({
  selectedItems,
  setSelectedItems,
}: {
  selectedItems: WorshipOrderItem[];
  setSelectedItems: React.Dispatch<React.SetStateAction<WorshipOrderItem[]>>;
}) {
  const [selectedDetail, setSelectedDetail] =
    useRecoilState(selectedDetailState);
  const [dragIndex, setDragIndex] = useState<number | null>(null);
  const [overIndex, setOverIndex] = useState<number | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<{ idx: number; title: string } | null>(null);

  const handleDeleteItem = (idx: number) => {
    setSelectedItems((prevItems) => prevItems.filter((_, i) => i !== idx));
  };

  const handleDrop = (dropIdx: number) => {
    if (dragIndex === null || dragIndex === dropIdx) return;
    setSelectedItems((prev) => {
      const next = [...prev];
      const [moved] = next.splice(dragIndex, 1);
      next.splice(dropIdx, 0, moved);
      return next;
    });
    setDragIndex(null);
    setOverIndex(null);
  };

  return (
    <section className="bg-white rounded-2xl border border-slate-100 shadow-sm p-6">
      <h2 className="text-xs font-black uppercase tracking-[0.2em] text-on-surface-variant mb-4">
        선택된 예배 순서
      </h2>
      <div className="flex flex-wrap gap-2">
        {selectedItems.length === 0 ? (
          <p className="text-on-surface-variant text-sm py-3">
            위에서 예배 순서를 클릭하여 추가하세요
          </p>
        ) : (
          selectedItems.map((item, i) => {
            const isSelected =
              selectedDetail.key === item.key &&
              selectedDetail.title === item.title;
            const isDragging = dragIndex === i;
            const isDragOver = overIndex === i && dragIndex !== i;

            return (
              <span
                key={item.key}
                draggable
                className={[
                  "inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-sm font-semibold cursor-pointer select-none transition-all border",
                  isSelected
                    ? "bg-electric-blue text-white border-electric-blue shadow-sm shadow-electric-blue/20"
                    : "bg-surface-low text-navy-dark border-slate-200 hover:border-electric-blue hover:bg-electric-blue/5",
                  isDragging ? "opacity-40" : "",
                  isDragOver ? "border-t-2 border-t-electric-blue" : "",
                ].join(" ")}
                onDragStart={() => setDragIndex(i)}
                onDragOver={(e) => {
                  e.preventDefault();
                  setOverIndex(i);
                }}
                onDragLeave={() => setOverIndex(null)}
                onDrop={(e) => {
                  e.preventDefault();
                  handleDrop(i);
                }}
                onDragEnd={() => {
                  setDragIndex(null);
                  setOverIndex(null);
                }}
                onClick={(e) => {
                  e.stopPropagation();
                  setSelectedDetail(item);
                }}
              >
                <span className="cursor-grab text-on-surface-variant/40 text-xs">⠿</span>
                {item.title}
                <button
                  className={[
                    "w-4 h-4 flex items-center justify-center rounded-full text-[9px] font-black transition-colors",
                    isSelected
                      ? "bg-white/20 text-white hover:bg-white/30"
                      : "bg-slate-200 text-slate-500 hover:bg-red-100 hover:text-red-500",
                  ].join(" ")}
                  onClick={(e) => {
                    e.stopPropagation();
                    setConfirmDelete({ idx: i, title: item.title });
                  }}
                >
                  ×
                </button>
              </span>
            );
          })
        )}
      </div>
      <ConfirmModal
        open={confirmDelete !== null}
        message={`'${confirmDelete?.title}' 항목을 삭제하시겠습니까?`}
        confirmLabel="삭제"
        danger
        onConfirm={() => {
          if (confirmDelete) handleDeleteItem(confirmDelete.idx);
          setConfirmDelete(null);
        }}
        onCancel={() => setConfirmDelete(null)}
      />
    </section>
  );
}
