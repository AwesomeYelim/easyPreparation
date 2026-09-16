import type { Metadata } from "next";
import Shot from "../components/Shot";
import { IconApple, IconLinux, IconWindows } from "../components/icons";
import { Button, Card, Code, Tag } from "../components/ui";
import {
  ASSETS,
  INSTALL_MAC,
  INSTALL_WIN,
  ISSUES,
  RELEASES,
  RELEASE_LATEST,
  asset,
} from "../lib/links";

export const metadata: Metadata = {
  title: "다운로드 — easyPreparation",
  description:
    "macOS · Windows · Linux. 기본 기능은 전부 무료, 설치 후 자동으로 최신 버전을 유지합니다.",
};

const desktop = [
  {
    icon: <IconApple />,
    name: "macOS",
    desc: "Apple Silicon (M1~M4) · .dmg",
    cta: ".dmg 다운로드",
    file: ASSETS.desktopMac,
    primary: true,
  },
  {
    icon: <IconWindows />,
    name: "Windows",
    desc: "Windows 10 / 11 (64-bit) · 설치 프로그램",
    cta: "setup.exe 다운로드",
    file: ASSETS.desktopWin,
    primary: true,
  },
  {
    icon: <IconLinux />,
    name: "Linux",
    desc: "Ubuntu 22.04+ / Debian 12+ · 실행 파일",
    cta: "다운로드",
    file: ASSETS.desktopLinux,
    primary: false,
  },
];

const more = [
  { label: "macOS (Intel) — 서버 바이너리", file: ASSETS.serverMacIntel },
  { label: "macOS (Apple Silicon) — 서버 바이너리", file: ASSETS.serverMacArm },
  { label: "Linux (x86_64) — 서버 바이너리", file: ASSETS.serverLinux },
  { label: "Windows (x86_64) — 서버 바이너리", file: ASSETS.serverWin },
  { label: "SHA256 체크섬 (checksums.txt)", file: ASSETS.checksums },
];

const firstRun = [
  <>앱을 열면 교회 이름을 묻는 초기 설정이 뜹니다. 한글·영문 이름을 넣고 저장.</>,
  <>
    예배 순서를 고르고 찬송 번호·성경 구절을 채운 뒤 <strong>Display 전송</strong>.
  </>,
  <>
    OBS에 Browser Source로 <Code>http://localhost:8080/display</Code> 를 추가하면
    송출 화면이 됩니다.
  </>,
  <>
    스마트폰으로 <Code>/mobile</Code> QR을 찍으면 리모컨이 됩니다.
  </>,
];

export default function DownloadPage() {
  return (
    <section className="mx-auto flex max-w-[960px] flex-col gap-14 px-6 pb-24 pt-20">
      <div className="flex flex-col items-center gap-3.5 text-center">
        <h1 className="text-3xl font-extrabold tracking-[-0.02em] text-white md:text-[44px]">
          다운로드
        </h1>
        <p className="max-w-[560px] text-base leading-[1.7] text-muted md:text-[17px]">
          운영체제를 고르면 바로 받습니다. 기본 기능은 전부 무료, 설치 후
          자동으로 최신 버전을 유지합니다.
        </p>
        <p className="text-[13px] text-dim">
          항상 최신 안정 버전을 내려받습니다 ·{" "}
          <a
            href={RELEASE_LATEST}
            target="_blank"
            rel="noopener noreferrer"
            className="text-accent hover:text-accent-light"
          >
            릴리즈 노트 보기
          </a>
        </p>
      </div>

      <div className="grid gap-5 sm:grid-cols-3">
        {desktop.map((d) => (
          <Card key={d.name} className="flex flex-col items-center gap-3.5 px-5 py-7">
            {d.icon}
            <div className="flex flex-col gap-1 text-center">
              <p className="text-[17px] font-bold text-white">{d.name}</p>
              <p className="text-xs text-dim">{d.desc}</p>
            </div>
            <Button
              href={asset(d.file)}
              variant={d.primary ? "primary" : "ghost"}
              className="w-full"
            >
              {d.cta}
            </Button>
          </Card>
        ))}
      </div>

      <div id="install" className="flex flex-col gap-3.5 scroll-mt-24">
        <h2 className="text-lg font-bold text-white">
          터미널 한 줄 설치{" "}
          <span className="text-[13px] font-medium text-dim">
            — 보안 경고(Gatekeeper / SmartScreen)까지 자동 처리
          </span>
        </h2>
        <div className="grid gap-4 md:grid-cols-2">
          {[
            { label: "macOS · Terminal", cmd: INSTALL_MAC },
            { label: "Windows · PowerShell", cmd: INSTALL_WIN },
          ].map((t) => (
            <div
              key={t.label}
              className="flex flex-col gap-1.5 rounded-card border border-line bg-raised px-[18px] py-4"
            >
              <span className="text-xs text-[#94a3b8]">{t.label}</span>
              <code className="whitespace-pre-wrap break-all font-mono text-[13px] text-[#86efac]">
                {t.cmd}
              </code>
            </div>
          ))}
        </div>
      </div>

      <div className="grid gap-5 md:grid-cols-2">
        <Card className="flex flex-col gap-3 p-6">
          <h3 className="text-base font-bold text-white">더 많은 플랫폼</h3>
          <ul className="flex flex-col gap-2 text-sm text-list">
            {more.map((m) => (
              <li key={m.file} className="flex justify-between gap-4">
                <span>{m.label}</span>
                <a
                  href={asset(m.file)}
                  className="whitespace-nowrap font-semibold text-accent hover:text-accent-light"
                >
                  받기
                </a>
              </li>
            ))}
          </ul>
          <p className="text-xs leading-relaxed text-dim">
            서버 바이너리는 데스크톱 앱 없이 터미널에서 실행하고 브라우저로{" "}
            <Code>http://localhost:8080</Code> 에 접속합니다.
          </p>
        </Card>

        <Card highlight className="flex flex-col gap-3 p-6">
          <div id="beta" className="flex scroll-mt-24 items-center gap-2">
            <Tag>BETA</Tag>
            <h3 className="text-base font-bold text-white">베타 채널</h3>
          </div>
          <p className="text-sm leading-relaxed text-list">
            새 기능을 먼저 써보고 피드백을 주실 교회를 찾습니다. 베타 빌드는
            자동 업데이트로 배포되지 않고, GitHub Releases에서{" "}
            <strong className="text-white">Pre-release</strong>로 표시된
            버전을 직접 받습니다.
          </p>
          <div className="mt-auto flex flex-wrap gap-x-4 gap-y-1 text-sm font-semibold">
            <a
              href={RELEASES}
              target="_blank"
              rel="noopener noreferrer"
              className="text-accent hover:text-accent-light"
            >
              베타 받기 →
            </a>
            <a
              href={ISSUES}
              target="_blank"
              rel="noopener noreferrer"
              className="text-accent hover:text-accent-light"
            >
              참여 · 피드백 (GitHub Issues) →
            </a>
          </div>
        </Card>
      </div>

      <div className="flex flex-col gap-5">
        <h2 className="text-[22px] font-extrabold text-white">설치 후 5분</h2>
        <div className="grid items-center gap-7 md:grid-cols-[1.1fr_1fr]">
          <div className="h-[280px] overflow-hidden rounded-xl">
            <Shot label={"[스크린샷 1 또는 3]\n첫 실행 · 예배 순서 화면"} className="h-full rounded-xl" />
          </div>
          <ol className="flex flex-col gap-3.5 text-[15px] text-list">
            {firstRun.map((step, i) => (
              <li key={i} className="flex gap-3">
                <span className="num inline-flex h-[26px] w-[26px] flex-none items-center justify-center rounded-full bg-accent text-xs font-bold text-bg">
                  {i + 1}
                </span>
                <span className="leading-relaxed">{step}</span>
              </li>
            ))}
          </ol>
        </div>
      </div>

      <div className="flex flex-col gap-3 rounded-card border border-[rgba(232,184,109,0.4)] bg-[rgba(232,184,109,0.08)] px-6 py-[22px]">
        <h3 className="text-[15px] font-bold text-[#f3cf95]">
          첫 실행에 보안 경고가 뜨면
        </h3>
        <p className="text-sm leading-[1.65] text-[#d9c6a3]">
          macOS: Finder에서 앱 우클릭 → 열기 → 열기. Windows: &ldquo;추가
          정보&rdquo; → &ldquo;실행&rdquo;. 위의 터미널 한 줄 설치를 쓰면 이
          단계가 필요 없습니다. 한 번 허용하면 이후엔 나오지 않습니다.
        </p>
      </div>
    </section>
  );
}
