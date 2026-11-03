import type { Metadata, Viewport } from "next";
import MobileRemote from "@/components/remote/MobileRemote";

// 폰 리모컨 (QR 로 접속). PWA 설치 정보는 Go 가 /mobile/manifest.json 으로 제공한다.
export const metadata: Metadata = {
  title: "EP Remote",
  manifest: "/mobile/manifest.json",
  appleWebApp: { capable: true, title: "EP Remote", statusBarStyle: "default" },
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  maximumScale: 1,
  userScalable: false,
  themeColor: "#002045",
};

export default function MobilePage() {
  return <MobileRemote />;
}
