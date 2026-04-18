import RecoilProvider from "@/components/RecoilProvider";
import { LocalAuthProvider } from "./lib/LocalAuthContext";
import { LicenseProvider } from "./lib/LicenseContext";
import SetupWizard from "./components/SetupWizard";
import { WebSocketProvider } from "./components/WebSocketProvider";
import GlobalDisplayPanel from "./components/GlobalDisplayPanel";
import SettingsLoader from "./components/SettingsLoader";
import ToastProvider from "./components/ToastProvider";
import AppShell from "./components/AppShell";
import "@/globals.css";
import type { Viewport, Metadata } from "next";

export const metadata: Metadata = {
  title: "easyPreparation",
  description: "예배 준비 자동화",
  icons: {
    icon: "/favicon.ico",
    apple: "/images/apple-touch-icon.png",
  },
};

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
        <RecoilProvider>
          <LocalAuthProvider>
            <LicenseProvider>
              <SetupWizard />
              <WebSocketProvider>
                <SettingsLoader />
                <AppShell>
                  {children}
                </AppShell>
                <GlobalDisplayPanel />
              </WebSocketProvider>
            </LicenseProvider>
          </LocalAuthProvider>
        </RecoilProvider>
      </body>
    </html>
  );
}
