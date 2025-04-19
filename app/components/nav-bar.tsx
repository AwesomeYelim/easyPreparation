"use client";

import { signIn, signOut, useSession } from "next-auth/react";
import classNames from "classnames";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState } from "react";

export default function NavBar() {
  const [sidebarOpen, setSidebarOpen] = useState(false);
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
      </nav>

      <div className="auth_wrap">
        {!session?.user && (
          <i
            title="login"
            className="login"
            onClick={() => {
              return signIn();
            }}
          />
        )}

        {session?.user && (
          <>
            <div style={{ position: "relative", display: "inline-block" }}>
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={session.user.image as string}
                alt="profile-img"
                style={{
                  width: 25,
                  height: 25,
                  borderRadius: "50%",
                  cursor: "pointer",
                }}
                onClick={() => setSidebarOpen(true)}
              />
            </div>

            {/* Overlay Background */}
            <div
              style={{
                position: "fixed",
                top: 0,
                left: 0,
                width: "100%",
                height: "100vh",
                backgroundColor: "rgba(0, 0, 0, 0.6)",
                opacity: sidebarOpen ? 1 : 0,
                visibility: sidebarOpen ? "visible" : "hidden",
                transition:
                  "opacity 0.3s ease-in-out, visibility 0.3s ease-in-out",
                zIndex: 998,
              }}
              onClick={() => setSidebarOpen(false)}
            />

            {/* Sidebar Panel */}
            <div
              style={{
                position: "fixed",
                top: 0,
                right: sidebarOpen ? 0 : "-320px",
                width: "320px",
                height: "100vh",
                backgroundColor: "#8a8a8a",
                padding: "20px",
                boxShadow: "-2px 0 6px rgba(0,0,0,0.3)",
                display: "flex",
                flexDirection: "column",
                alignItems: "center",
                transition: "right 0.3s ease-in-out",
                zIndex: 999,
              }}
            >
              {/* Close Button */}
              <button
                style={{
                  alignSelf: "flex-end",
                  background: "none",
                  border: "none",
                  fontSize: "20px",
                  color: "#fff",
                  cursor: "pointer",
                }}
                onClick={() => setSidebarOpen(false)}
              >
                ✕
              </button>

              {/* Profile Section */}
              <img
                src={session.user.image || "/default-profile.png"}
                alt="profile"
                style={{
                  width: "80px",
                  height: "80px",
                  borderRadius: "50%",
                  marginTop: "10px",
                }}
              />
              <div
                style={{
                  marginTop: "10px",
                  fontWeight: "bold",
                  fontSize: "16px",
                  color: "#204d87",
                }}
              >
                {session.user.name}
              </div>
              <div
                style={{
                  fontSize: "12px",
                  color: "#ddd",
                  marginBottom: "20px",
                }}
              >
                {session.user.email}
              </div>

              {/* Menu List */}
              <div
                style={{
                  backgroundColor: "rgba(255, 255, 255, 0.1)",
                  padding: "20px",
                  borderRadius: "12px",
                  width: "100%",
                }}
              >
                {[
                  ["소속교회", "동남생명의 빛 교회"],
                  ["교회표기", "Light of Light Church"],
                  ["라이선스 정보"],
                  ["주보 생성 내역"],
                  ["PPT 생성 내역"],
                  ["가사 PPT 생성 내역"],
                ].map(([title, value]) => (
                  <div
                    key={title}
                    style={{
                      display: "flex",
                      justifyContent: "space-between",
                      alignItems: "center",
                      padding: "10px 0",
                      borderBottom: "1px solid rgba(255,255,255,0.2)",
                      color: "#fff",
                    }}
                  >
                    <div>
                      <div style={{ fontWeight: "bold", fontSize: "15px" }}>
                        {title}
                      </div>
                      {value && (
                        <div style={{ fontSize: "13px", opacity: 0.8 }}>
                          {value}
                        </div>
                      )}
                    </div>
                    <div style={{ fontSize: "18px", opacity: 0.7 }}>›</div>
                  </div>
                ))}
              </div>

              {/* Logout */}
              <div
                style={{
                  marginTop: "auto",
                  color: "#b65050",
                  fontWeight: "bold",
                  cursor: "pointer",
                }}
                onClick={() => {
                  setSidebarOpen(false);
                  signOut();
                }}
              >
                logout
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
