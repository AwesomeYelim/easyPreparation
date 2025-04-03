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
  const [selectedInfo, setSelectedInfo] =
    useState<WorshipOrderItem[]>(worshipOrder);

  const sendDataToGoServer = async () => {
    try {
      const response = await fetch("http://localhost:8080/api/submit", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ selectedInfo }), // JSON 형식으로 변환
      });

      if (!response.ok) {
        throw new Error("서버 응답 실패");
      }

      const data = await response.json();
      console.log("서버 응답:", data);
      alert("서버로 데이터 전송 성공!");
    } catch (error) {
      console.error("서버 전송 중 오류 발생:", error);
      alert("서버 전송 실패");
    }
  };

  return (
    <div className="bulletin_container">
      <div className="top_bar">
        <button onClick={sendDataToGoServer} className="send_button">
          데이터 전송
        </button>
      </div>
      <div className="bulletin_wrap">
        <div className="editable">
          <WorshipOrder
            selectedItems={selectedInfo}
            setSelectedItems={setSelectedInfo}
          />
          <SelectedOrder
            selectedItems={selectedInfo}
            setSelectedItems={setSelectedInfo}
          />
          <Detail setSelectedItems={setSelectedInfo} />
        </div>
        <div className="result">
          <ResultPart
            selectedItems={selectedInfo}
            setSelectedItems={setSelectedInfo}
          />
        </div>
      </div>
    </div>
  );
}
