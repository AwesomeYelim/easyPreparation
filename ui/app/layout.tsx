import RecoilProvider from "@/components/RecoilProvider";
import NavBar from "@/components/NavBar";
import AuthProvider from "./lib/next-auth";
import { WebSocketProvider } from "./components/WebSocketProvider";
import GlobalDisplayPanel from "./components/GlobalDisplayPanel";
import SettingsLoader from "./components/SettingsLoader";
import "@/globals.css";
import type { Viewport } from "next";

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  maximumScale: 1,
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="ko">
      <body>
        <div className="entire-wrap">
          <RecoilProvider>
            <AuthProvider>
              <WebSocketProvider>
                <SettingsLoader />
                <NavBar />
                {children}
                <GlobalDisplayPanel />
              </WebSocketProvider>
            </AuthProvider>
          </RecoilProvider>
        </div>
      </body>
    </html>
  );
}
