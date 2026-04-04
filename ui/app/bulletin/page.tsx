"use client";

import { WorshipOrder } from "./components/WorshipOrder";
import SelectedOrder from "./components/SelectedOrder";
import Detail from "./components/Detail";
import { useState, useEffect, useRef } from "react";
import classNames from "classnames";
import { WorshipType, userInfoState, worshipOrderState, displayPanelOpenState, displayItemsState } from "@/recoilState";
import { WorshipOrderItem } from "@/types";
import { apiClient, openDisplayWindow } from "@/lib/apiClient";
import { useRecoilState, useRecoilValue, useSetRecoilState } from "recoil";
import { ResultPart } from "./components/ResultPage";
import { useWS } from "@/components/WebSocketProvider";
import s from "./bulletin.module.scss";

export default function Bulletin() {
  const [selectedWorshipType, setSelectedWorshipType] =
    useState<WorshipType>("main_worship");
  const [worshipOrder, setWorshipOrder] = useRecoilState(worshipOrderState);
  const selectedInfo = worshipOrder[selectedWorshipType];
  const setSelectedInfo: React.Dispatch<React.SetStateAction<WorshipOrderItem[]>> = (updater) => {
    setWorshipOrder((prev) => ({
      ...prev,
      [selectedWorshipType]: typeof updater === "function" ? updater(prev[selectedWorshipType]) : updater,
    }));
  };
  const userInfo = useRecoilValue(userInfoState);
  const { subscribe } = useWS();
  const setDisplayItems = useSetRecoilState(displayItemsState);
  const setDisplayPanelOpen = useSetRecoilState(displayPanelOpenState);

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

  // selectedInfo는 worshipOrder[selectedWorshipType]에서 직접 파생
  // — 드롭다운 전환 시 자동으로 해당 타입의 데이터를 보여줌

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

  const [displayLoading, setDisplayLoading] = useState(false);
  const [displayProgress, setDisplayProgress] = useState("");

  const sendToDisplay = async () => {
    const processedInfo = processSelectedInfo(selectedInfo);
    setDisplayItems(processedInfo);
    setDisplayPanelOpen(true);
    openDisplayWindow();
    try {
      setDisplayLoading(true);
      setDisplayProgress("예배 순서 전송 중...");
      const res = await apiClient.startDisplay(processedInfo, userInfo.english_name, userInfo.email);
      if (!res.ok) throw new Error("Display 전송 실패");
    } catch (error) {
      console.error("Display 전송 에러:", error);
      alert("Display 전송 중 오류가 발생했습니다.");
    } finally {
      setDisplayLoading(false);
      setDisplayProgress("");
    }
  };

  const removeEmptyNodes = (items: WorshipOrderItem[]): WorshipOrderItem[] => {
    return items
      .map((item) => {
        if (item.children) {
          item = { ...item, children: removeEmptyNodes(item.children) };
        }
        return item;
      })
      .filter((item) => {
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
      if (!saverRes.ok) throw new Error("저장 실패");

      const response = await apiClient.submitBulletin({
        mark: userInfo.english_name,
        targetInfo: processedInfo,
        target: selectedWorshipType,
        figmaInfo: userInfo.figmaInfo,
        email: userInfo.email,
      });

      if (!response.ok) throw new Error("서버 응답 실패");

      alert("서버로 데이터 전송 성공!");
    } catch (error) {
      console.error("서버 전송 중 오류 발생:", error);
      alert("서버 전송 실패");
      setLoading(false);
    }
  };

  return (
    <div className={s.bulletin_container}>
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

      <div className={s.top_bar}>
        <select
          value={selectedWorshipType}
          onChange={(e) =>
            setSelectedWorshipType(e.target.value as WorshipType)
          }
          className={s.worship_select}
        >
          <option value="main_worship">주일예배</option>
          <option value="after_worship">오후예배</option>
          <option value="wed_worship">수요예배</option>
          <option value="fri_worship">금요예배</option>
        </select>

        <button
          disabled={!userInfo.figmaInfo.key || !userInfo.figmaInfo.token}
          onClick={sendDataToGoServer}
          className={classNames(s.send_button, {
            [s.disabled]: !userInfo.figmaInfo.key || !userInfo.figmaInfo.token,
          })}
        >
          예배 자료 생성하기
        </button>

        <button
          onClick={sendToDisplay}
          className={`${s.send_button} ${s.display_send_btn}`}
        >
          Display 전송
        </button>
      </div>

      <div className={s.bulletin_wrap}>
        <div className={s.editable}>
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
        <div className={s.result}>
          <ResultPart selectedItems={selectedInfo} />
        </div>
      </div>

    </div>
  );
}
