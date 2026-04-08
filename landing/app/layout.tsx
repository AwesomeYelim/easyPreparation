import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "easyPreparation — 예배 준비 자동화",
  description:
    "찬양 악보, 주보 PDF, OBS 방송 송출까지. 교회 예배 준비를 하나의 도구로.",
  icons: {
    icon: "/favicon.ico",
    apple: "/apple-touch-icon.png",
  },
};

function Header() {
  return (
    <header className="sticky top-0 z-50 border-b border-gray-200 bg-white/80 backdrop-blur">
      <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
        <a href="/" className="flex items-center gap-2.5 text-xl font-bold text-navy">
          <img src="/ep-logo-192.png" alt="EP" width={28} height={28} />
          easyPreparation
        </a>
        <nav className="flex items-center gap-6 text-sm font-medium text-gray-600">
          <a href="/pricing" className="hover:text-navy">
            요금제
          </a>
          <a href="/download" className="hover:text-navy">
            다운로드
          </a>
          <a
            href="https://github.com/AwesomeYelim/easyPreparation"
            target="_blank"
            rel="noopener noreferrer"
            className="hover:text-navy"
          >
            GitHub
          </a>
        </nav>
      </div>
    </header>
  );
}

function Footer() {
  return (
    <footer className="border-t border-gray-200 bg-gray-50 py-10 text-center text-sm text-gray-500">
      <p>easyPreparation</p>
      <p className="mt-1">
        <a
          href="https://github.com/AwesomeYelim/easyPreparation"
          className="underline hover:text-navy"
        >
          GitHub
        </a>
      </p>
    </footer>
  );
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="ko">
      <body className="font-sans text-gray-900 antialiased">
        <Header />
        <main>{children}</main>
        <Footer />
      </body>
    </html>
  );
}
