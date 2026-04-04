"use client";

import { useEffect } from "react";
import { useSession } from "next-auth/react";
import { useSetRecoilState } from "recoil";
import { userSettingsState } from "@/recoilState";
import { apiClient } from "@/lib/apiClient";

export default function SettingsLoader() {
  const { data: session } = useSession();
  const setSettings = useSetRecoilState(userSettingsState);

  useEffect(() => {
    const email = session?.user?.email;
    if (!email) return;

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
  }, [session?.user?.email, setSettings]);

  return null;
}
