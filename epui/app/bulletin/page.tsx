"use client";

import { WorshipOrder } from "./components/WorshipOrder";
import SelectedOrder from "./components/SelectedOrder";
import Detail from "./components/Detail";
import { useState } from "react";
import { worshipOrderState } from "../recoilState";
import { useRecoilValue } from "recoil";
import { ResultPart } from "./components/ResultPage";

export type WorshipOrderItem = {
  key: string;
  title: string;
  obj: string;
  info: string;
  lead?: string;
  children?: WorshipOrderItem[];
};

export default function Bulletin() {
  const worshipOrder = useRecoilValue(worshipOrderState);
  const [selectedInfo, setSelectedInfo] = useState<WorshipOrderItem[]>(worshipOrder);

  console.log(selectedInfo);
  return (
    <div className="bulletin_wrap">
      <div className="editable">
        <WorshipOrder selectedItems={selectedInfo} setSelectedItems={setSelectedInfo} />
        <SelectedOrder selectedItems={selectedInfo} setSelectedItems={setSelectedInfo} />
        <Detail setSelectedItems={setSelectedInfo} />
      </div>
      <div className="result">
        <ResultPart selectedItems={selectedInfo} setSelectedItems={setSelectedInfo} />
      </div>
    </div>
  );
}
