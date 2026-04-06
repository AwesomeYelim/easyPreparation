"use client";

import { useState } from "react";
import { useRecoilState } from "recoil";
import { displayPanelOpenState } from "@/recoilState";
import { useAuth } from "@/lib/LocalAuthContext";
import NavLink from "./NavLink";
import Sidebar from "./SideBar";
import s from "./NavBar.module.scss";

export default function NavBar() {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const { church } = useAuth();
  const [panelOpen, setPanelOpen] = useRecoilState(displayPanelOpenState);

  const handleDisplayClick = () => {
    window.open(
      `${process.env.NEXT_PUBLIC_API_BASE_URL}/display`,
      "display_window"
    );
    setPanelOpen(!panelOpen);
  };

  return (
    <div className={s.nav_wrapper}>
      <nav>
        <NavLink href="/bulletin">Bulletin</NavLink>
        <NavLink href="/lyrics">Lyrics</NavLink>
        <NavLink href="/bible">Bible</NavLink>
      </nav>

      <div className={s.auth_wrap}>
        <button
          className={`${s.display_open_btn}${panelOpen ? ` ${s.active}` : ""}`}
          onClick={handleDisplayClick}
          title="Display 창 열기 / 제어판 토글"
        >
          Display
        </button>
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
