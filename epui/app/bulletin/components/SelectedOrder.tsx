import { useSetRecoilState, useRecoilState } from "recoil";
import { worshipOrderState, selectedDetailState } from "@/recoilState";
import React from "react";
import { WorshipOrderItem } from "./WorshipOrder";

export default function SelectedOrder({
  selectedItems,
  setSelectedItems,
}: {
  selectedItems: WorshipOrderItem[];
  setSelectedItems: React.Dispatch<React.SetStateAction<WorshipOrderItem[]>>;
}) {
  // setWorshipOrder를 useSetRecoilState로 변경하여 setter만 사용합니다.
  const setWorshipOrder = useSetRecoilState(worshipOrderState);
  const setSelectedDetail = useSetRecoilState(selectedDetailState);

  const handleDeleteItem = (item?: WorshipOrderItem) => {
    setSelectedItems((prevItems) => prevItems.filter((el) => el !== item));
    setWorshipOrder((prevOrder: WorshipOrderItem[]) => [...prevOrder, item]);
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
                }}>
                {item.title.split("_")[1]}
                <button
                  className="delete-btn"
                  onClick={(e) => {
                    handleDeleteItem(item);
                  }}>
                  ❌
                </button>
              </span>
            </>
          );
        })}
      </div>
    </section>
  );
}
