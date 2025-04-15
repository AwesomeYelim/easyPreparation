"use client";
import RecoilProvider from "@/components/recoil-provider";
import NavBar from "@/components/nav-bar";
import { SessionProvider } from "next-auth/react";
import "@/globals.css";
import "@/styles.scss";

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="ko">
      <body>
        <div className="entire-wrap">
          <SessionProvider>
            <NavBar />
            <RecoilProvider>{children}</RecoilProvider>
          </SessionProvider>
        </div>
      </body>
    </html>
  );
}
