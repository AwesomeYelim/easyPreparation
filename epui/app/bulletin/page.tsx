"use client";

import { WorshipOrder, WorshipOrderItem } from "./components/WorshipOrder";
import SelectedOrder from "./components/SelectedOrder";
import Detail from "./components/Detail";
import { useState } from "react";
import { worshipOrderState } from "../recoilState";
import { useRecoilValue } from "recoil";

export default function Bulletin() {
  const worshipOrder = useRecoilValue(worshipOrderState);
  const [selectedItems, setSelectedItems] =
    useState<WorshipOrderItem[]>(worshipOrder);

  return (
    <div className="bulletin_wrap">
      <div className="editable">
        <WorshipOrder
          selectedItems={selectedItems}
          setSelectedItems={setSelectedItems}
        />
        <SelectedOrder
          selectedItems={selectedItems}
          setSelectedItems={setSelectedItems}
        />
        <Detail />
      </div>
      <div className="result">
        <h2>생성된 예배 내용</h2>
        <div className="contents">
          {selectedItems.map((el) => {
            return (
              <div className="row">
                <div className="title">{el.title.split("_")[1]}</div>
                <div className="obj">{el.obj}</div>
                <div className="lead">{el.lead}</div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
