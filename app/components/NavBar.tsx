"use client";

import { signIn, useSession } from "next-auth/react";
import { useState } from "react";
import NavLink from "./NavLink";
import Sidebar from "./SideBar";
import ProfileButton from "./ProfileButton";

export default function NavBar() {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const { data: session } = useSession();

  return (
    <div className="nav_wrapper">
      <nav>
        <NavLink href="/bulletin">Bulletin</NavLink>
        <NavLink href="/lyrics">Lyrics</NavLink>
      </nav>

      <div className="auth_wrap">
        {!session?.user ? (
          <i title="login" className="login" onClick={() => signIn()} />
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
