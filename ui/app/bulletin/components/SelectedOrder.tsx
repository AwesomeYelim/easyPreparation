import React, { useState } from "react";
import { useRecoilState } from "recoil";
import { selectedDetailState } from "@/recoilState";
import { WorshipOrderItem } from "@/types";
import classNames from "classnames";
import s from "../bulletin.module.scss";

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

  const handleDeleteItem = (idx: number, title: string) => {
    if (!window.confirm(`'${title}' 항목을 삭제하시겠습니까?`)) return;
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
    <section className={s.card}>
      <h2>선택된 예배 순서</h2>
      <div>
        {selectedItems.length === 0 ? (
          <p className={s.empty_guide}>위에서 예배 순서를 클릭하여 추가하세요</p>
        ) : (
          selectedItems.map((item, i) => (
            <span
              key={item.key}
              draggable
              className={classNames(s.tag, {
                [s.selected]:
                  selectedDetail.key === item.key &&
                  selectedDetail.title === item.title,
                [s.dragging]: dragIndex === i,
                [s.drag_over]: overIndex === i && dragIndex !== i,
              })}
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
              <span className={s.drag_handle}>⠿</span>
              {item.title}
              <button
                className={s.delete_btn}
                onClick={(e) => {
                  e.stopPropagation();
                  handleDeleteItem(i, item.title);
                }}
              >
                x
              </button>
            </span>
          ))
        )}
      </div>
    </section>
  );
}
