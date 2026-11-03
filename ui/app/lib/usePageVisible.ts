"use client";
import { useEffect, useState } from "react";

/**
 * 페이지가 화면에 보이는지 (document.visibilityState).
 * 분리 창은 숨긴 채 미리 떠 있으므로, 숨겨진 동안 무거운 작업(미리보기 iframe·폴링)을 쉬게 하는 데 쓴다.
 */
export function usePageVisible(): boolean {
  const [visible, setVisible] = useState(
    typeof document === "undefined" ? true : document.visibilityState === "visible",
  );
  useEffect(() => {
    const onChange = () => setVisible(document.visibilityState === "visible");
    document.addEventListener("visibilitychange", onChange);
    return () => document.removeEventListener("visibilitychange", onChange);
  }, []);
  return visible;
}
