"use client";

import { WorshipOrder } from "./components/WorshipOrder";
import SelectedOrder from "./components/SelectedOrder";
import Detail from "./components/Detail";
import { useState, useEffect, useRef } from "react";
import classNames from "classnames";
import { WorshipType, userInfoState, worshipOrderState } from "@/recoilState";
import { WorshipOrderItem } from "@/types";
import { apiClient } from "@/lib/apiClient";
import { useRecoilValue } from "recoil";
import { ResultPart } from "./components/ResultPage";
import { useWS } from "@/components/WebSocketProvider";
import DisplayControlPanel from "./components/DisplayControlPanel";

export default function Bulletin() {
  const [selectedWorshipType, setSelectedWorshipType] =
    useState<WorshipType>("main_worship");
  const worshipOrder = useRecoilValue(worshipOrderState);
  const [selectedInfo, setSelectedInfo] = useState<WorshipOrderItem[]>(
    worshipOrder[selectedWorshipType]
  );
  const userInfo = useRecoilValue(userInfoState);
  const { subscribe } = useWS();

  const [loading, setLoading] = useState(false);
  const [wsMessage, setWsMessage] = useState("");
  const [wsLogs, setWsLogs] = useState<string[]>([]);
  const msgQueueRef = useRef<string[]>([]);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const lastMsgRef = useRef("");

  const flushQueue = () => {
    if (timerRef.current) return;
    timerRef.current = setInterval(() => {
      const next = msgQueueRef.current.shift();
      if (!next) {
        clearInterval(timerRef.current!);
        timerRef.current = null;
        return;
      }
      setWsMessage(next);
      setWsLogs((prev) => [...prev.slice(-9), next]);
    }, 400);
  };

  const enqueueMsg = (msg: string) => {
    if (!msg || msg === lastMsgRef.current) return;
    lastMsgRef.current = msg;
    msgQueueRef.current.push(msg);
    flushQueue();
  };

  // 예배 종류에 따른 초기값 설정
  useEffect(() => {
    if (worshipOrder[selectedWorshipType]) {
      setSelectedInfo(worshipOrder[selectedWorshipType]);
    }
  }, [selectedWorshipType, worshipOrder]);

  useEffect(() => {
    const ignored = new Set(["navigate", "order", "keepalive", "position"]);

    return subscribe((message) => {
      if (message.type === "display_loading") {
        if (message.done) {
          setDisplayLoading(false);
          setDisplayProgress("");
        } else {
          setDisplayProgress(message.message || "");
        }
        return;
      }

      if (message.type === "done") {
        // 큐 즉시 비우고 완료 표시
        msgQueueRef.current = [];
        if (timerRef.current) { clearInterval(timerRef.current); timerRef.current = null; }
        downloadZip(message.fileName);
        setWsMessage("Success !!");
        setWsLogs([]);
        setLoading(false);
      } else if (!ignored.has(message.type)) {
        enqueueMsg(message.message || "");
      }
    });
  }, [subscribe]);

  const downloadZip = (fileName: string) => apiClient.downloadFile(fileName);

  const [panelOpen, setPanelOpen] = useState(false);
  const displayStarted = useRef(false);
  const [displayItems, setDisplayItems] = useState<WorshipOrderItem[]>([]);
  const [displayLoading, setDisplayLoading] = useState(false);
  const [displayProgress, setDisplayProgress] = useState("");

  const toggleDisplay = async () => {
    if (panelOpen) {
      setPanelOpen(false);
      return;
    }

    // 최초 1회만 display 창 열기 + order 전송
    if (!displayStarted.current) {
      const processedInfo = processSelectedInfo(selectedInfo);
      setDisplayItems(processedInfo);
      setDisplayLoading(true);
      setDisplayProgress("예배 화면 준비 중...");
      window.open(
        `${window.location.protocol}//${window.location.hostname}:8080/display`,
        "displayWindow",
        "width=1280,height=720"
      );
      await apiClient.startDisplay(processedInfo);
      displayStarted.current = true;
      setDisplayLoading(false);
    }

    setPanelOpen(true);
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

  const processSelectedInfo = (
    data: WorshipOrderItem[]
  ): WorshipOrderItem[] => {
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
      setWsLogs([]);

      const processedInfo = processSelectedInfo(selectedInfo);

      const saverRes = await apiClient.saveBulletin(selectedWorshipType, processedInfo);
      if (saverRes.status == 500) throw new Error("저장 실패");

      const response = await apiClient.submitBulletin({
        mark: userInfo.english_name,
        targetInfo: processedInfo,
        target: selectedWorshipType,
        figmaInfo: userInfo.figmaInfo,
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
    <div className={classNames("bulletin_container", { panel_open: panelOpen })}>
      {(loading || displayLoading) && (
        <div className="loading_overlay">
          <div className="spinner"></div>
          {!displayLoading && wsLogs.length > 1 && (
            <div className="ws_logs">
              {wsLogs.slice(0, -1).map((log, i) => (
                <div key={i}>{log}</div>
              ))}
            </div>
          )}
          <div className="ws_message">
            {displayLoading ? displayProgress : wsMessage}
          </div>
        </div>
      )}

      <div className="top_bar">
        <select
          value={selectedWorshipType}
          onChange={(e) =>
            setSelectedWorshipType(e.target.value as WorshipType)
          }
          className="worship_select"
        >
          <option value="main_worship">주일예배</option>
          <option value="after_worship">오후예배</option>
          <option value="wed_worship">수요예배</option>
        </select>

        <button
          disabled={!userInfo.figmaInfo.key || !userInfo.figmaInfo.token}
          onClick={sendDataToGoServer}
          className={classNames("send_button", {
            disabled: !userInfo.figmaInfo.key || !userInfo.figmaInfo.token,
          })}
        >
          예배 자료 생성하기
        </button>

        <button
          onClick={toggleDisplay}
          className={classNames("send_button display_start_btn", { active: panelOpen })}
        >
          {panelOpen ? "제어판 닫기" : "예배 화면 시작"}
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
          <ResultPart selectedItems={selectedInfo} />
        </div>
      </div>

      {panelOpen && (
        <DisplayControlPanel
          items={displayItems}
          onClose={() => setPanelOpen(false)}
        />
      )}
    </div>
  );
}
