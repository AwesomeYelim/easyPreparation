"use client";

import { useEffect } from "react";
import { useRecoilValue, useSetRecoilState } from "recoil";
import { displayPanelOpenState, displayItemsState } from "@/recoilState";
import { apiClient } from "@/lib/apiClient";
import { WorshipOrderItem } from "@/types";
import DisplayControlPanel from "@/bulletin/components/DisplayControlPanel";

export default function GlobalDisplayPanel() {
  const panelOpen = useRecoilValue(displayPanelOpenState);
  const setPanelOpen = useSetRecoilState(displayPanelOpenState);
  const setItems = useSetRecoilState(displayItemsState);

  // 페이지 로드 시 서버에 활성 순서가 있으면 자동 복원
  useEffect(() => {
    apiClient.getDisplayStatus()
      .then((data: any) => {
        if (Array.isArray(data.items) && data.items.length > 0) {
          setItems(data.items as WorshipOrderItem[]);
          setPanelOpen(true);
        }
      })
      .catch((e) => console.error("display 상태 복원 실패:", e));
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  if (!panelOpen) return null;

  return (
    <div className="fixed top-0 right-0 w-80 h-screen z-40 shadow-2xl shadow-navy-dark/5">
      <DisplayControlPanel />
    </div>
  );
}
