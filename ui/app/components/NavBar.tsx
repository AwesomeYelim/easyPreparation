"use client";

import { useState } from "react";
import { useRecoilState } from "recoil";
import { displayPanelOpenState } from "@/recoilState";
import { useAuth } from "@/lib/LocalAuthContext";
import { openDisplayWindow } from "@/lib/apiClient";
import NavLink from "./NavLink";
import Sidebar from "./SideBar";
import s from "./NavBar.module.scss";

export default function NavBar() {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const { church } = useAuth();
  const [panelOpen, setPanelOpen] = useRecoilState(displayPanelOpenState);

  const handlePanelToggle = () => {
    setPanelOpen(!panelOpen);
  };

  const handleOpenWindow = () => {
    openDisplayWindow();
  };

  return (
    <div className={s.nav_wrapper}>
      <nav>
        <NavLink href="/bulletin">주보</NavLink>
        <NavLink href="/lyrics">찬양</NavLink>
        <NavLink href="/bible">성경</NavLink>
      </nav>

      <div className={s.auth_wrap}>
        <div className={s.display_btn_group}>
          <button
            className={`${s.display_open_btn}${panelOpen ? ` ${s.active}` : ""}`}
            onClick={handlePanelToggle}
            title="예배 화면 제어판 열기/닫기"
          >
            예배 화면
          </button>
          <button
            className={s.display_window_btn}
            onClick={handleOpenWindow}
            title="예배 화면을 새 창으로 열기"
          >
            <svg width="14" height="14" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M6 2H3C2.44772 2 2 2.44772 2 3V13C2 13.5523 2.44772 14 3 14H13C13.5523 14 14 13.5523 14 13V10M10 2H14V6M14 2L7 9" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
          </button>
        </div>
        <button
          className={s.menu_btn}
          onClick={() => setSidebarOpen(true)}
          title="메뉴"
        >
          {church?.name || "설정"}
        </button>
        <Sidebar
          open={sidebarOpen}
          onClose={() => setSidebarOpen(false)}
        />
      </div>
    </div>
  );
}
