"use client";

import { useEffect } from "react";
import { useAuth } from "@/lib/LocalAuthContext";
import { useSetRecoilState } from "recoil";
import { userSettingsState, userInfoState } from "@/recoilState";
import { apiClient } from "@/lib/apiClient";

const BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";

export default function SettingsLoader() {
  const { church, isLoading } = useAuth();
  const setSettings = useSetRecoilState(userSettingsState);
  const setUserInfo = useSetRecoilState(userInfoState);

  useEffect(() => {
    if (isLoading) return;

    const email = church?.email || "local@localhost";

    // userInfo 로드 (교회 정보)
    fetch(`${BASE_URL}/api/user?email=${encodeURIComponent(email)}`)
      .then((r) => r.json())
      .then((data) => {
        if (data && !data.error) {
          setUserInfo(data);
        }
      })
      .catch(() => {});

    // settings 로드
    apiClient
      .getSettings(email)
      .then((data) => {
        if (data && !data.error) {
          setSettings(data);
          // 테마 적용
          document.documentElement.setAttribute("data-theme", data.theme || "light");
          // 폰트 크기 적용
          if (data.font_size) {
            document.documentElement.style.setProperty("--user-font-size", `${data.font_size}px`);
          }
        }
      })
      .catch(() => {});
  }, [church, isLoading, setSettings, setUserInfo]);

  return null;
}
