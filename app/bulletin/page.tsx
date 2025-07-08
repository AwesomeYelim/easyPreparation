"use client";

import { WorshipOrder } from "./components/WorshipOrder";
import SelectedOrder from "./components/SelectedOrder";
import Detail from "./components/Detail";
import { useState, useEffect } from "react";
import classNames from "classnames";
import { WorshipType, userInfoState, worshipOrderState } from "@/recoilState";
import { useRecoilValue } from "recoil";
import { ResultPart } from "./components/ResultPage";
import { useWS } from "@/components/WebSocketProvider";

export type WorshipOrderItem = {
  key: string;
  title: string;
  obj: string;
  info: string;
  lead?: string;
  children?: WorshipOrderItem[];
};

export default function Bulletin() {
  const [selectedWorshipType, setSelectedWorshipType] = useState<WorshipType>("main_worship");
  const worshipOrder = useRecoilValue(worshipOrderState);
  const [selectedInfo, setSelectedInfo] = useState<WorshipOrderItem[]>(worshipOrder[selectedWorshipType]);
  const userInfo = useRecoilValue(userInfoState);
  const { message, isOpen } = useWS();

  const [loading, setLoading] = useState(false);
  const [wsMessage, setWsMessage] = useState("");

  // 예배 종류에 따른 초기값 설정
  useEffect(() => {
    if (worshipOrder[selectedWorshipType]) {
      setSelectedInfo(worshipOrder[selectedWorshipType]);
    }
  }, [selectedWorshipType, worshipOrder]);

  useEffect(() => {
    if (!message) return;

    if (message.type === "done" && message.target === "main_worship") {
      downloadZip(message.fileName);
      setWsMessage("Success !!");
      setLoading(false);
    } else {
      console.log(message);
      setWsMessage(message.message || "");
    }
  }, [message]);

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

  const removeEmptyNodes = (items: WorshipOrderItem[]): WorshipOrderItem[] => {
    return items
      .map((item) => {
        // 자식이 있으면 자식 먼저 처리
        if (item.children) {
          item = { ...item, children: removeEmptyNodes(item.children) };
        }
        return item;
      })
      .filter((item) => {
        // key가 ".0"으로 끝나고 title이 "-"인 경우 제외 (삭제)
        const isKeyEndsWithZero = item.key.endsWith(".0");
        const isTitleDash = item.title === "-";
        return !(isKeyEndsWithZero && isTitleDash);
      });
  };

  const processSelectedInfo = (data: WorshipOrderItem[]): WorshipOrderItem[] => {
    return data.map((item) => {
      if (item.title === "교회소식" && item.children) {
        return {
          ...item,
          children: removeEmptyNodes(item.children),
        };
      }
      return item;
    });
  };

  const sendDataToGoServer = async () => {
    try {
      setLoading(true);
      setWsMessage("");

      const baseUrl = process.env.NEXT_PUBLIC_API_BASE_URL;
      const response = await fetch(`${baseUrl}/submit`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          mark: userInfo.english_name,
          targetInfo: processSelectedInfo(selectedInfo),
          target: selectedWorshipType,
          figmaInfo: userInfo.figmaInfo,
        }),
      });

      if (!response.ok) throw new Error("서버 응답 실패");

      const data = await response.json();
      console.log("서버 응답:", data);
      alert("서버로 데이터 전송 성공!");
    } catch (error) {
      console.error("서버 전송 중 오류 발생:", error);
      alert("서버 전송 실패");
      setLoading(false);
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
        <select
          value={selectedWorshipType}
          onChange={(e) => setSelectedWorshipType(e.target.value as WorshipType)}
          className="worship_select">
          <option value="main_worship">주일예배</option>
          <option value="after_worship">오후예배</option>
          <option value="wed_worship">수요예배</option>
        </select>

        <button
          disabled={!userInfo.figmaInfo.key || !userInfo.figmaInfo.token}
          onClick={sendDataToGoServer}
          className={classNames("send_button", {
            disabled: !userInfo.figmaInfo.key || !userInfo.figmaInfo.token,
          })}>
          예배 자료 생성하기
        </button>
      </div>

      <div className="bulletin_wrap">
        <div className="editable">
          <WorshipOrder selectedItems={selectedInfo} setSelectedItems={setSelectedInfo} />
          <SelectedOrder selectedItems={selectedInfo} setSelectedItems={setSelectedInfo} />
          <Detail setSelectedItems={setSelectedInfo} />
        </div>
        <div className="result">
          <ResultPart selectedItems={selectedInfo} />
        </div>
      </div>
    </div>
  );
}
