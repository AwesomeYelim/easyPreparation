"use client";

import { WorshipOrder } from "./components/WorshipOrder";
import SelectedOrder from "./components/SelectedOrder";
import Detail from "./components/Detail";
import { useState, useEffect, useRef } from "react";
import classNames from "classnames";
import { WorshipType, userInfoState, worshipOrderState, displayPanelOpenState, displayItemsState } from "@/recoilState";
import { WorshipOrderItem } from "@/types";
import { apiClient, openDisplayWindow } from "@/lib/apiClient";
import toast from "react-hot-toast";
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
  const [displayLoading, setDisplayLoading] = useState(false);
  const [displayProgress, setDisplayProgress] = useState("");
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

  // 예배 순서 API에서 로드
  useEffect(() => {
    apiClient.getWorshipOrder(selectedWorshipType).then((data) => {
      if (Array.isArray(data) && data.length > 0) {
        setWorshipOrder((prev) => ({ ...prev, [selectedWorshipType]: data }));
      }
    }).catch((err) => console.error("예배 순서 로드 실패:", err));
  }, [selectedWorshipType]);

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
      toast.error("Display 전송 중 오류가 발생했습니다.");
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

      const saverRes = await apiClient.saveWorshipOrder(selectedWorshipType, processedInfo);
      if (!saverRes.ok) throw new Error("저장 실패");

      const response = await apiClient.submitBulletin({
        mark: userInfo.english_name,
        targetInfo: processedInfo,
        target: selectedWorshipType,
        figmaInfo: userInfo.figmaInfo,
        email: userInfo.email,
      });

      if (!response.ok) throw new Error("서버 응답 실패");

      toast.success("서버로 데이터 전송 성공!");
    } catch (error) {
      console.error("서버 전송 중 오류 발생:", error);
      toast.error("서버 전송 실패");
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
          onClick={sendDataToGoServer}
          className={s.send_button}
          title="주보 PDF 파일을 생성하여 다운로드합니다"
        >
          <svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg" style={{verticalAlign: "middle", marginRight: "6px"}}>
            <path d="M8 2V10M8 10L5 7M8 10L11 7M3 13H13" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
          주보 PDF 생성
        </button>

        <button
          onClick={sendToDisplay}
          className={`${s.send_button} ${s.display_send_btn}`}
          title="예배 순서를 프로젝터 화면에 전송합니다"
        >
          <svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg" style={{verticalAlign: "middle", marginRight: "6px"}}>
            <path d="M2 3H14V11H2V3Z" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
            <path d="M5 14H11M8 11V14" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
          프로젝터에 보내기
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
