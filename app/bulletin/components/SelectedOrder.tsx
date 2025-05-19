import React from "react";
import { useRecoilState, useSetRecoilState } from "recoil";
import { selectedDetailState } from "@/recoilState";
import { WorshipOrderItem } from "../page";
import classNames from "classnames";

export default function SelectedOrder({
  selectedItems,
  setSelectedItems,
}: {
  selectedItems: WorshipOrderItem[];
  setSelectedItems: React.Dispatch<React.SetStateAction<WorshipOrderItem[]>>;
}) {
  const [selectedDetail, setSelectedDetail] =
    useRecoilState(selectedDetailState);

  const handleDeleteItem = (item: WorshipOrderItem) => {
    setSelectedItems((prevItems) => prevItems.filter((el) => el !== item));
  };

  return (
    <section className="card">
      <h2>선택된 예배 순서</h2>
      <div>
        {selectedItems.map((item) => {
          return (
            <span
              key={item.key}
              className={classNames("tag", {
                selected:
                  selectedDetail.key === item.key &&
                  selectedDetail.title == item.title,
              })}
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
          );
        })}
      </div>
    </section>
  );
}
