import React from "react";
import { useSetRecoilState } from "recoil";
import { selectedDetailState } from "@/recoilState";
import { WorshipOrderItem } from "../page";

export default function SelectedOrder({
  selectedItems,
  setSelectedItems,
}: {
  selectedItems: WorshipOrderItem[];
  setSelectedItems: React.Dispatch<React.SetStateAction<WorshipOrderItem[]>>;
}) {
  const setSelectedDetail = useSetRecoilState(selectedDetailState);

  const handleDeleteItem = (item: WorshipOrderItem) => {
    setSelectedItems((prevItems) => prevItems.filter((el) => el !== item));
  };

  return (
    <section className="card">
      <h2>선택된 예배 순서</h2>
      <div>
        {selectedItems.map((item) => {
          // 각 항목에 대한 상세 정보를 selectedDetail 객체에서 key로 접근
          return (
            <>
              <span
                className="tag"
                onClick={(e) => {
                  e.stopPropagation();
                  setSelectedDetail(item);
                }}
              >
                {item.title}
                <button
                  className="delete-btn"
                  onClick={() => {
                    handleDeleteItem(item);
                  }}
                >
                  x
                </button>
              </span>
            </>
          );
        })}
      </div>
    </section>
  );
}
