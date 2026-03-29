"use client";

import { signIn, useSession } from "next-auth/react";
import { useState } from "react";
import { useRecoilState } from "recoil";
import { displayPanelOpenState } from "@/recoilState";
import NavLink from "./NavLink";
import Sidebar from "./SideBar";
import ProfileButton from "./ProfileButton";
import s from "./NavBar.module.scss";

export default function NavBar() {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const { data: session } = useSession();
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
        {!session?.user ? (
          <i title="login" className={s.login} onClick={() => signIn()} />
        ) : (
          <>
            <ProfileButton
              image={session.user.image!}
              onClick={() => setSidebarOpen(true)}
            />
            <Sidebar
              open={sidebarOpen}
              onClose={() => setSidebarOpen(false)}
              user={session.user}
            />
          </>
        )}
      </div>
    </div>
  );
}
