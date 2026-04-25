import RecoilProvider from "@/components/RecoilProvider";
import { LocalAuthProvider } from "./lib/LocalAuthContext";
import { LicenseProvider } from "./lib/LicenseContext";
import SetupWizard from "./components/SetupWizard";
import { WebSocketProvider } from "./components/WebSocketProvider";
import GlobalDisplayPanel from "./components/GlobalDisplayPanel";
import SettingsLoader from "./components/SettingsLoader";
import ToastProvider from "./components/ToastProvider";
import WebSocketToastBridge from "./components/WebSocketToastBridge";
import ProShell from "@/components/ProShell";
import ProTopBar from "@/components/ProTopBar";
import ProIconBar from "@/components/ProIconBar";
import ProMainArea from "@/components/ProMainArea";
import ProSequencePanel from "@/components/ProSequencePanel";
import ProInspectorPanel from "@/components/ProInspectorPanel";
import ProTimeline from "@/components/ProTimeline";
import ClientOnly from "@/components/ClientOnly";
import "@/globals.css";
import type { Viewport, Metadata } from "next";

export const metadata: Metadata = {
  title: "easyPreparation",
  description: "예배 준비 자동화",
  icons: {
    icon: [
      { url: "/images/ep-logo.svg", type: "image/svg+xml" },
      { url: "/favicon.ico", sizes: "any" },
    ],
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
    <html lang="ko" suppressHydrationWarning translate="no">
      <body suppressHydrationWarning>
        <ClientOnly>
          <ToastProvider />
          <RecoilProvider>
            <LocalAuthProvider>
              <LicenseProvider>
                <SetupWizard />
                <WebSocketProvider>
                  <WebSocketToastBridge />
                  <SettingsLoader />
                  <ProShell>
                    <ProTopBar />
                    <ProIconBar />
                    <ProSequencePanel />
                    <ProMainArea>
                      {children}
                    </ProMainArea>
                    <ProInspectorPanel />
                    <ProTimeline />
                  </ProShell>
                  <GlobalDisplayPanel />
                </WebSocketProvider>
              </LicenseProvider>
            </LocalAuthProvider>
          </RecoilProvider>
        </ClientOnly>
      </body>
    </html>
  );
}
