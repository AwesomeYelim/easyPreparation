"use client";

import { WorshipOrder } from "./components/WorshipOrder";
import SelectedOrder from "./components/SelectedOrder";
import Detail from "./components/Detail";
import { useState } from "react";
import { userInfoState, worshipOrderState } from "@/recoilState";
import { useRecoilValue } from "recoil";
import { ResultPart } from "./components/ResultPage";
import { useRef } from "react";
import { useEffect } from "react";

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
  const userInfo = useRecoilValue(userInfoState);

  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    const baseUrl = process.env.NEXT_PUBLIC_WS_URL;
    const ws = new WebSocket(baseUrl || `ws://localhost:8080/ws`);

    ws.onopen = () => {
      console.log("WebSocket 연결 성공");
    };

    ws.onmessage = (event) => {
      console.log("WebSocket 메시지 수신:", event.data);
      const message = JSON.parse(event.data);

      // 완료 알림 수신시 PDF 자동 다운로드
      if (message.type === "done" && message.target === "main_worship") {
        downloadPDF(message.fileName);
      }
    };

    ws.onclose = () => {
      console.log("WebSocket 연결 종료");
    };

    return () => {
      ws.close();
    };
  }, []);

  const downloadPDF = (fileName: string) => {
    const baseUrl = process.env.NEXT_PUBLIC_API_BASE_URL;
    const link = document.createElement("a");
    link.href = `${baseUrl}/download?target=${fileName}`;
    link.download = fileName;
    link.target = "_self"; // 새로운 페이지가 열리지 않도록 _self로 설정

    // 링크 클릭하여 다운로드
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  const sendDataToGoServer = async () => {
    try {
      const baseUrl = process.env.NEXT_PUBLIC_API_BASE_URL;
      const response = await fetch(`${baseUrl}/submit`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          targetInfo: selectedInfo,
          target: "main_worship",
          figmaInfo: userInfo.figmaInfo,
        }),
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
          Submit
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
