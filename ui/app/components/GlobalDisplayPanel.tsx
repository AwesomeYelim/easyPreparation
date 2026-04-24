"use client";

import { useEffect } from "react";
import { useSetRecoilState } from "recoil";
import { displayPanelOpenState, displayItemsState } from "@/recoilState";
import { apiClient } from "@/lib/apiClient";
import { WorshipOrderItem } from "@/types";

export default function GlobalDisplayPanel() {
  const setPanelOpen = useSetRecoilState(displayPanelOpenState);
  const setItems = useSetRecoilState(displayItemsState);

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

  return null;
}
