import RecoilProvider from "@/components/RecoilProvider";
import NavBar from "@/components/NavBar";
import { LocalAuthProvider } from "./lib/LocalAuthContext";
import { LicenseProvider } from "./lib/LicenseContext";
import SetupWizard from "./components/SetupWizard";
import { WebSocketProvider } from "./components/WebSocketProvider";
import GlobalDisplayPanel from "./components/GlobalDisplayPanel";
import SettingsLoader from "./components/SettingsLoader";
import UpdateChecker from "./components/UpdateChecker";
import ToastProvider from "./components/ToastProvider";
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
        <ToastProvider />
        <div className="entire-wrap">
          <RecoilProvider>
            <LocalAuthProvider>
              <LicenseProvider>
                <SetupWizard />
                <WebSocketProvider>
                  <SettingsLoader />
                  <NavBar />
                  <UpdateChecker />
                  {children}
                  <GlobalDisplayPanel />
                </WebSocketProvider>
              </LicenseProvider>
            </LocalAuthProvider>
          </RecoilProvider>
        </div>
      </body>
    </html>
  );
}
