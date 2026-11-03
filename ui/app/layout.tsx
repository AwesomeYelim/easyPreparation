import RecoilProvider from "@/components/RecoilProvider";
import { LocalAuthProvider } from "./lib/LocalAuthContext";
import { LicenseProvider } from "./lib/LicenseContext";
import ToastProvider from "./components/ToastProvider";
import AppChrome from "@/components/AppChrome";
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
      <head>
        <link rel="preload" href="/font/NanumBrush.ttf" as="font" type="font/truetype" crossOrigin="anonymous" />
        <link rel="preload" href="/font/NanumGothic-regular.ttf" as="font" type="font/truetype" crossOrigin="anonymous" />
        <link rel="preload" href="/font/NanumGothic-800.ttf" as="font" type="font/truetype" crossOrigin="anonymous" />
        <link rel="preload" href="/font/JacquesFrancois-regular.ttf" as="font" type="font/truetype" crossOrigin="anonymous" />
      </head>
      <body suppressHydrationWarning>
        <ClientOnly>
          <ToastProvider />
          <RecoilProvider>
            <LocalAuthProvider>
              <LicenseProvider>
                <AppChrome>{children}</AppChrome>
              </LicenseProvider>
            </LocalAuthProvider>
          </RecoilProvider>
        </ClientOnly>
      </body>
    </html>
  );
}
