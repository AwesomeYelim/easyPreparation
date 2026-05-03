"use client";

import { useEffect, useRef } from "react";

/**
 * 디바운스 자동저장 훅.
 * value가 null이면 아직 로드 전으로 간주.
 * 처음 non-null이 되는 순간을 "로드 완료"로 표시하고,
 * 이후 변경부터 delay ms 디바운스 후 saveFn 호출.
 */
export function useAutoSave<T>(
  value: T | null,
  saveFn: (v: T) => void,
  delay = 600
): void {
  const loadedRef = useRef(false);

  useEffect(() => {
    if (!loadedRef.current) {
      if (value !== null) loadedRef.current = true;
      return;
    }
    if (value === null) return;
    const t = setTimeout(() => saveFn(value), delay);
    return () => clearTimeout(t);
  }, [value]); // eslint-disable-line react-hooks/exhaustive-deps
}
