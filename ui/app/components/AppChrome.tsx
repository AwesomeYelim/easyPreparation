"use client";
import { usePathname } from "next/navigation";
import SetupWizard from "@/components/SetupWizard";
import { WebSocketProvider } from "@/components/WebSocketProvider";
import GlobalDisplayPanel from "@/components/GlobalDisplayPanel";
import SettingsLoader from "@/components/SettingsLoader";
import WebSocketToastBridge from "@/components/WebSocketToastBridge";
import WelcomeTour from "@/components/WelcomeTour";
import ProShell from "@/components/ProShell";
import ProTopBar from "@/components/ProTopBar";
import ProIconBar from "@/components/ProIconBar";
import ProMainArea from "@/components/ProMainArea";
import ProSequencePanel from "@/components/ProSequencePanel";
import ProInspectorPanel from "@/components/ProInspectorPanel";
import ProTimeline from "@/components/ProTimeline";
import LicensePanelMount from "@/components/LicensePanelMount";

/**
 * 앱 외곽 구성.
 * - /mobile : 폰 리모컨 (QR 접속) — 마찬가지로 ProShell 없이.
 * - /remote : 분리된 순서 패널 창 — ProShell(탑바·아이콘바·인스펙터·타임라인)·마법사·투어 없이 패널만.
 * - 그 외   : 기존 ProShell 전체.
 */
export default function AppChrome({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();

  // /remote(분리 창)·/mobile(폰 리모컨)은 ProShell 없이 단독 화면
  if (pathname?.startsWith("/remote") || pathname?.startsWith("/mobile")) {
    return <WebSocketProvider>{children}</WebSocketProvider>;
  }

  return (
    <>
      <SetupWizard />
      <WebSocketProvider>
        <WebSocketToastBridge />
        <WelcomeTour />
        <SettingsLoader />
        <ProShell>
          <ProTopBar />
          <ProIconBar />
          <ProSequencePanel />
          <ProMainArea>{children}</ProMainArea>
          <ProInspectorPanel />
          <ProTimeline />
        </ProShell>
        <GlobalDisplayPanel />
        <LicensePanelMount />
      </WebSocketProvider>
    </>
  );
}
