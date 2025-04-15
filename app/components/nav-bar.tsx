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
