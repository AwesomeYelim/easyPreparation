import type { Metadata } from "next";
import { Noto_Sans_KR } from "next/font/google";
import "./globals.css";
import { GITHUB, ISSUES, README, RELEASES, USAGE_GUIDE } from "./lib/links";

const noto = Noto_Sans_KR({
  subsets: ["latin"],
  weight: ["400", "500", "700", "800"],
  display: "swap",
  variable: "--font-noto",
});

export const metadata: Metadata = {
  title: "easyPreparation — 교회 예배 준비 소프트웨어",
  description:
    "주보 PDF, 찬송 악보, 성경 슬라이드, OBS 송출까지 한 화면에서. 교회 미디어팀을 위한 예배 준비 데스크톱 앱.",
  icons: {
    icon: "/favicon.ico",
    apple: "/apple-touch-icon.png",
  },
};

const nav = [
  { label: "기능", href: "/#features" },
  { label: "요금제", href: "/pricing" },
  { label: "다운로드", href: "/download" },
  { label: "문서", href: README, external: true },
];

function Header() {
  return (
    <header className="sticky top-0 z-50 border-b border-line bg-bg/85 backdrop-blur">
      <div className="mx-auto flex max-w-site items-center justify-between px-6 py-4">
        <a
          href="/"
          className="flex items-center gap-2.5 text-lg font-bold text-white"
        >
          <img
            src="/ep-logo.svg"
            alt=""
            width={28}
            height={28}
            className="rounded-[7px] shadow-[0_0_0_1px_rgba(255,255,255,0.12)]"
          />
          easyPreparation
        </a>
        <nav className="flex items-center gap-4 text-sm font-medium text-muted sm:gap-7">
          {nav.map((n) => (
            <a
              key={n.label}
              href={n.href}
              className="hidden hover:text-white sm:inline"
              {...(n.external
                ? { target: "_blank", rel: "noopener noreferrer" }
                : {})}
            >
              {n.label}
            </a>
          ))}
          <a
            href="/download"
            className="inline-flex h-[38px] items-center rounded-btn bg-white px-4 font-bold text-bg transition hover:bg-accent-light"
          >
            무료 다운로드
          </a>
        </nav>
      </div>
    </header>
  );
}

const footerCols: { title: string; links: { label: string; href: string; external?: boolean }[] }[] = [
  {
    title: "제품",
    links: [
      { label: "기능", href: "/#features" },
      { label: "요금제", href: "/pricing" },
      { label: "다운로드", href: "/download" },
      { label: "릴리즈 노트", href: RELEASES, external: true },
    ],
  },
  {
    title: "사용 안내",
    links: [
      { label: "설치 가이드", href: "/download#install" },
      { label: "사용 설명서", href: USAGE_GUIDE, external: true },
      { label: "GitHub", href: GITHUB, external: true },
    ],
  },
  {
    title: "베타 · 문의",
    links: [
      { label: "베타 참여하기", href: "/download#beta" },
      { label: "버그 · 문의 (GitHub Issues)", href: ISSUES, external: true },
    ],
  },
];

function Footer() {
  return (
    <footer className="border-t border-line bg-band">
      <div className="mx-auto grid max-w-site gap-8 px-6 pb-8 pt-14 text-sm text-muted sm:grid-cols-2 md:grid-cols-[1.4fr_repeat(3,minmax(0,1fr))]">
        <div className="flex flex-col gap-3">
          <div className="flex items-center gap-2 text-base font-bold text-white">
            <img
              src="/ep-logo.svg"
              alt=""
              width={22}
              height={22}
              className="rounded-md"
            />
            easyPreparation
          </div>
          <p className="leading-relaxed text-dim">
            교회 예배 준비 자동화 데스크톱 앱
          </p>
        </div>
        {footerCols.map((c) => (
          <div key={c.title} className="flex flex-col gap-2.5">
            <span className="font-bold text-white">{c.title}</span>
            {c.links.map((l) => (
              <a
                key={l.label}
                href={l.href}
                className="hover:text-white"
                {...(l.external
                  ? { target: "_blank", rel: "noopener noreferrer" }
                  : {})}
              >
                {l.label}
              </a>
            ))}
          </div>
        ))}
      </div>
      <div className="mx-auto max-w-site px-6 pb-8 text-xs text-faint">
        © {new Date().getFullYear()} easyPreparation
      </div>
    </footer>
  );
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="ko" className={noto.variable}>
      <body className="bg-bg font-sans text-ink antialiased">
        <Header />
        <main>{children}</main>
        <Footer />
      </body>
    </html>
  );
}
