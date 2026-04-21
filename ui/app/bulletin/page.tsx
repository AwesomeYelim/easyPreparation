"use client";

import { WorshipOrder } from "./components/WorshipOrder";
import SelectedOrder from "./components/SelectedOrder";
import Detail from "./components/Detail";
import { useState, useEffect, useRef } from "react";
import { WorshipType, userInfoState, worshipOrderState, displayPanelOpenState } from "@/recoilState";
import { WorshipOrderItem } from "@/types";
import { apiClient, openDisplayWindow } from "@/lib/apiClient";
import toast from "react-hot-toast";
import { useRecoilState, useRecoilValue, useSetRecoilState } from "recoil";
import { ResultPart } from "./components/ResultPage";
import { useWS } from "@/components/WebSocketProvider";

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
  const setDisplayPanelOpen = useSetRecoilState(displayPanelOpenState);

  const [loading, setLoading] = useState(false);
  const [wsMessage, setWsMessage] = useState("");
  const [wsLogs, setWsLogs] = useState<string[]>([]);
  const [displayLoading, setDisplayLoading] = useState(false);
  const [displayProgress, setDisplayProgress] = useState("");
  const msgQueueRef = useRef<string[]>([]);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const lastMsgRef = useRef("");
  const processingRef = useRef(false); // "done" 중복 처리 방지 (StrictMode 이중 WS 연결)

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

  // 예배 순서 API에서 로드 (아직 로드 안 된 타입만 — 이미 있으면 편집 내용 유지)
  useEffect(() => {
    if (worshipOrder[selectedWorshipType]?.length > 0) return;
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
        if (!processingRef.current) return; // 이미 처리했으면 중복 무시
        processingRef.current = false;
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

  const downloadZip = async (fileName: string) => {
    try {
      await apiClient.downloadFile(fileName);
      toast.success("다운로드 폴더에 저장되었습니다.");
    } catch (e) {
      const msg = e instanceof Error && e.message.includes("주보를 생성")
        ? e.message
        : "다운로드 중 오류가 발생했습니다.";
      toast.error(msg);
    }
  };

  const sendToDisplay = async () => {
    const processedInfo = processSelectedInfo(selectedInfo);
    if (processedInfo.length === 0) {
      toast.error("예배 순서가 비어 있습니다. 먼저 순서를 불러오세요.");
      setDisplayPanelOpen(true);
      return;
    }
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
      processingRef.current = true;
      setWsMessage("");
      setWsLogs([]);

      const processedInfo = processSelectedInfo(selectedInfo);

      const saverRes = await apiClient.saveWorshipOrder(selectedWorshipType, processedInfo);
      if (!saverRes.ok) throw new Error("저장 실패");

      const response = await apiClient.submitBulletin({
        mark: userInfo.english_name,
        targetInfo: processedInfo,
        target: selectedWorshipType,
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

  const worshipTypeLabels: Record<WorshipType, string> = {
    main_worship: "주일예배",
    after_worship: "오후예배",
    wed_worship: "수요예배",
    fri_worship: "금요예배",
  };

  return (
    <div className="flex flex-col w-full min-h-full">
      {/* 로딩 오버레이 (주보 PDF 생성 전용 — display 전송은 제어판 loadingMsg로 표시) */}
      {loading && (
        <div className="fixed inset-0 z-50 flex flex-col items-center justify-center bg-navy-dark/80 backdrop-blur-sm">
          <div className="w-12 h-12 border-4 border-white/20 border-t-electric-blue rounded-full animate-spin mb-6" />
          {wsLogs.length > 1 && (
            <div className="text-white/50 text-xs font-mono mb-2 flex flex-col items-center gap-1">
              {wsLogs.slice(0, -1).map((log, i) => (
                <div key={i}>{log}</div>
              ))}
            </div>
          )}
          <div className="text-white text-sm font-bold tracking-wide">
            {wsMessage}
          </div>
        </div>
      )}

      {/* 상단 헤더 */}
      <div className="flex items-center justify-between mb-8">
        <h1 className="text-3xl font-black tracking-tight text-primary">Worship Sequence</h1>
        <div className="flex items-center gap-3">
          {/* 예배 타입 드롭다운 */}
          <select
            value={selectedWorshipType}
            onChange={(e) => setSelectedWorshipType(e.target.value as WorshipType)}
            className="bg-white text-navy-dark font-semibold text-sm px-4 py-2.5 border border-slate-200 rounded-xl cursor-pointer transition-all hover:border-electric-blue focus:outline-none focus:border-electric-blue focus:ring-2 focus:ring-electric-blue/20 shadow-sm flex-shrink-0"
          >
            <option value="main_worship">주일예배</option>
            <option value="after_worship">오후예배</option>
            <option value="wed_worship">수요예배</option>
            <option value="fri_worship">금요예배</option>
          </select>

          {/* 주보 PDF 생성 버튼 */}
          <button
            onClick={sendDataToGoServer}
            title="주보/예배 PDF 파일을 생성하여 다운로드합니다"
            className="flex items-center gap-2 bg-white text-navy-dark px-5 py-2.5 rounded-xl font-bold text-sm border border-slate-200 shadow-sm hover:bg-slate-50 transition-all whitespace-nowrap"
          >
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M8 2V10M8 10L5 7M8 10L11 7M3 13H13" stroke="#3B82F6" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            주보/예배 PDF 생성
          </button>

          {/* 프로젝터 전송 버튼 */}
          <button
            onClick={sendToDisplay}
            title="예배 순서를 프로젝터 화면에 전송합니다"
            className="flex items-center gap-2 bg-electric-blue text-white px-5 py-2.5 rounded-xl font-bold text-sm shadow-sm shadow-electric-blue/30 hover:bg-secondary transition-all active:scale-[0.98] whitespace-nowrap"
          >
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M2 3H14V11H2V3Z" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
              <path d="M5 14H11M8 11V14" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            프로젝터에 보내기
          </button>
        </div>
      </div>

      {/* 메인 콘텐츠 그리드 */}
      <div className="flex gap-6 w-full min-h-0">
        {/* 좌측: 편집 영역 */}
        <div className="flex flex-col gap-5 flex-1 min-w-0">
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
        {/* 우측: 미리보기 */}
        <div className="w-80 flex-shrink-0">
          <ResultPart selectedItems={selectedInfo} />
        </div>
      </div>
    </div>
  );
}
