"use client";

import { signIn, signOut, useSession } from "next-auth/react";
import classNames from "classnames";
import Link from "next/link";
import { usePathname } from "next/navigation";

export default function NavBar() {
  const { data: session } = useSession();
  const location = usePathname();

  return (
    <div className="nav_wrapper">
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

        <i
          title={(() => {
            if (session?.user) {
              return "logout";
            }
            return "login";
          })()}
          className={classNames("login", { logout: !!session?.user })}
          onClick={() => {
            if (session?.user) {
              return signOut();
            }
            return signIn();
          }}
        />
        {session?.user && (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={session?.user.image as string}
            style={{ width: 25, height: 25, borderRadius: "50%" }}
            alt="profile-img"
          />
        )}
      </nav>
    </div>
  );
}
