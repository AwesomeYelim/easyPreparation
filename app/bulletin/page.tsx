"use client";

import { WorshipOrder } from "./components/WorshipOrder";
import SelectedOrder from "./components/SelectedOrder";
import Detail from "./components/Detail";
import { useState, useEffect, useRef } from "react";
import classNames from "classnames";
import { userInfoState, worshipOrderState } from "@/recoilState";
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
  const userInfo = useRecoilValue(userInfoState);
  const ws = useRef<WebSocket | null>(null);

  const [loading, setLoading] = useState(false);
  const [wsMessage, setWsMessage] = useState("");

  useEffect(() => {
    let pingInterval: NodeJS.Timeout;

    const mount = () => {
      const baseUrl = process.env.NEXT_PUBLIC_WS_URL;
      ws.current = new WebSocket(baseUrl as string);

      ws.current.onopen = () => {
        console.log("WebSocket connected");
        // Ping 메시지 30초마다 보내기
        pingInterval = setInterval(() => {
          if (ws.current?.readyState === WebSocket.OPEN) {
            ws.current.send(JSON.stringify({ type: "ping" }));
          }
        }, 30000); // 30초
      };

      ws.current.onmessage = (event) => {
        console.log("WebSocket Received Message:", event.data);
        const message = JSON.parse(event.data);
        setWsMessage(message.message);

        if (message.type === "done" && message.target === "main_worship") {
          downloadZip(message.fileName);
          setWsMessage("Success !!");
          setLoading(false);
        }
      };

      ws.current.onclose = () => {
        console.log("WebSocket closed");
        setLoading(false);
        clearInterval(pingInterval);
      };
    };

    mount();

    return () => {
      if (ws.current?.readyState === WebSocket.OPEN) {
        ws.current.close();
      }
      clearInterval(pingInterval);
    };
  }, []);

  const downloadZip = (fileName: string) => {
    const baseUrl = process.env.NEXT_PUBLIC_API_BASE_URL;
    const link = document.createElement("a");
    link.href = `${baseUrl}/download?target=${fileName}`;
    link.download = fileName;
    link.target = "_self";
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  const sendDataToGoServer = async () => {
    try {
      setLoading(true); // 로딩 시작
      setWsMessage(""); // 이전 메시지 초기화

      const baseUrl = process.env.NEXT_PUBLIC_API_BASE_URL;
      const response = await fetch(`${baseUrl}/submit`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          mark: userInfo.english_name,
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

      // 로딩은 유지 (WebSocket 메시지 기다리는 중)
    } catch (error) {
      console.error("서버 전송 중 오류 발생:", error);
      alert("서버 전송 실패");
      setLoading(false); // 실패 시 로딩 종료
    }
  };

  return (
    <div className="bulletin_container">
      {loading && (
        <div className="loading_overlay">
          <div className="spinner"></div>
          {wsMessage && <div className="ws_message">{wsMessage}</div>}
        </div>
      )}

      <div className="top_bar">
        <button
          disabled={!userInfo.figmaInfo.key || !userInfo.figmaInfo.token}
          onClick={sendDataToGoServer}
          className={classNames("send_button", {
            disabled: !userInfo.figmaInfo.key || !userInfo.figmaInfo.token,
          })}
        >
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
