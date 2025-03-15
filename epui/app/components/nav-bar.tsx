"use client";

// import { signIn, signOut, useSession } from "next-auth/react";
import classNames from "classnames";
import Cookies from "js-cookie";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useCallback, useEffect, useState } from "react";

export default function NavBar() {
  // const { data: session } = useSession();
  const location = usePathname();

  const {
    navInfo: [navInfo, setNavInfo],
    dark: [dark, setDark],
  } = {
    navInfo: useState({ to: 0, hide: true }),
    dark: useState(Cookies.get("theme") === "dark"),
  };

  // const admin = session?.user?.email === "uiop01900@gmail.com";

  // const scrollToHide = useCallback(() => {
  //   setNavInfo((prev) => ({ ...prev, to: window.pageYOffset }));
  //   //  스크롤 아래로 내릴때
  //   if (navInfo.to < window.pageYOffset) {
  //     setNavInfo({ to: window.pageYOffset, hide: false });
  //     //  스크롤 위로 올릴때
  //   } else if (navInfo.to > window.pageYOffset) {
  //     setNavInfo({ to: window.pageYOffset, hide: true });
  //   }
  // }, [navInfo]);

  // useEffect(() => {
  //   document.addEventListener("scroll", scrollToHide);
  //   return () => document.removeEventListener("scroll", scrollToHide);
  // }, [navInfo]);

  // useEffect(() => {
  //   if (dark) {
  //     document.documentElement.dataset.theme = "dark";
  //   } else {
  //     document.documentElement.dataset.theme = "light";
  //   }
  // }, []);

  return (
    <div
      className="nav_wrapper"
      // style={{ height: navInfo.hide ? "111px" : "0px" }}
    >
      <nav>
        <Link
          href="/bulletin"
          className={classNames({ active: location?.includes("/bulletin") })}
        >
          Bulletin
        </Link>

        <Link
          href="/lyrics"
          className={classNames({ active: location?.includes("/lyrics") })}
        >
          Lyrics
        </Link>
      </nav>
    </div>
  );
}
