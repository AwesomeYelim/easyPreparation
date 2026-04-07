const RELEASE_BASE =
  "https://github.com/AwesomeYelim/easyPreparation/releases/latest/download";
const RELEASE_URL =
  "https://github.com/AwesomeYelim/easyPreparation/releases/latest";

const desktopApps = [
  {
    name: "macOS (Apple Silicon)",
    desc: "M1 / M2 / M3 / M4",
    file: "easyPreparation_desktop_darwin_arm64.zip",
    icon: "\u{f8ff}",
    install: "압축 해제 → Applications 폴더로 이동",
  },
  {
    name: "Windows",
    desc: "Windows 10 / 11 (64-bit)",
    file: "easyPreparation_desktop_windows_amd64_setup.exe",
    icon: "W",
    install: "실행하여 설치",
  },
  {
    name: "Linux",
    desc: "Ubuntu 22.04+ / Debian 12+ (64-bit)",
    file: "easyPreparation_desktop_linux_amd64",
    icon: "L",
    install: "chmod +x 후 실행",
  },
];

const serverBinaries = [
  {
    name: "macOS (Apple Silicon)",
    file: "easyPreparation_server_darwin_arm64",
  },
  { name: "macOS (Intel)", file: "easyPreparation_server_darwin_amd64" },
  { name: "Linux (x86_64)", file: "easyPreparation_server_linux_amd64" },
  { name: "Windows (x86_64)", file: "easyPreparation_server_windows_amd64.exe" },
];

export default function DownloadPage() {
  return (
    <section className="mx-auto max-w-4xl px-6 py-20">
      <h1 className="text-center text-3xl font-bold text-navy">다운로드</h1>
      <p className="mt-3 text-center text-gray-600">
        운영체제에 맞는 버전을 선택하세요. 기본 기능은 모두 무료입니다.
      </p>

      {/* Desktop 앱 (권장) */}
      <h2 className="mt-12 text-xl font-bold text-navy">
        Desktop 앱{" "}
        <span className="ml-2 rounded-full bg-navy/10 px-2.5 py-0.5 text-xs font-medium text-navy">
          권장
        </span>
      </h2>
      <p className="mt-1 text-sm text-gray-500">
        설치 한 번으로 바로 사용. 자동 업데이트를 지원합니다.
      </p>

      <div className="mt-6 grid gap-4 sm:grid-cols-3">
        {desktopApps.map((p) => (
          <a
            key={p.file}
            href={`${RELEASE_BASE}/${p.file}`}
            className="group flex flex-col items-center gap-3 rounded-xl border border-gray-200 p-5 shadow-sm transition hover:border-navy hover:shadow-md"
          >
            <span className="flex h-12 w-12 items-center justify-center rounded-lg bg-gray-100 text-lg font-bold text-navy group-hover:bg-navy group-hover:text-white">
              {p.icon}
            </span>
            <div className="text-center">
              <p className="font-semibold text-gray-900">{p.name}</p>
              <p className="text-xs text-gray-500">{p.desc}</p>
              <p className="mt-1 text-xs text-gray-400">{p.install}</p>
            </div>
          </a>
        ))}
      </div>

      {/* Server 바이너리 */}
      <details className="mt-10">
        <summary className="cursor-pointer text-lg font-bold text-navy hover:text-navy-light">
          Server 바이너리 (고급 사용자)
        </summary>
        <p className="mt-2 text-sm text-gray-500">
          Desktop 앱 없이 터미널에서 직접 실행합니다. 브라우저에서{" "}
          <code className="rounded bg-gray-100 px-1.5 py-0.5 text-xs">
            http://localhost:8080
          </code>
          으로 접속합니다.
        </p>
        <div className="mt-4 grid gap-3 sm:grid-cols-2">
          {serverBinaries.map((p) => (
            <a
              key={p.file}
              href={`${RELEASE_BASE}/${p.file}`}
              className="flex items-center justify-between rounded-lg border border-gray-200 px-4 py-3 text-sm hover:border-navy hover:bg-gray-50"
            >
              <span className="font-medium text-gray-700">{p.name}</span>
              <span className="text-xs text-gray-400">{p.file}</span>
            </a>
          ))}
        </div>
        <p className="mt-3 text-right">
          <a
            href={RELEASE_URL}
            target="_blank"
            rel="noopener noreferrer"
            className="text-sm text-navy underline"
          >
            모든 릴리즈 보기 &rarr;
          </a>
        </p>
      </details>

      {/* 설치 후 시작하기 */}
      <div className="mt-12 rounded-xl border border-gray-200 bg-gray-50 p-6 text-sm text-gray-600">
        <h2 className="font-semibold text-gray-900">설치 후 시작하기</h2>
        <ol className="mt-3 list-inside list-decimal space-y-1.5">
          <li>다운로드한 Desktop 앱을 설치 및 실행합니다.</li>
          <li>자동으로 브라우저가 열리며 예배 준비 화면이 표시됩니다.</li>
          <li>예배 순서를 입력하고 Display 화면을 열어 확인합니다.</li>
          <li>
            OBS 방송 연동 시, Browser Source에{" "}
            <code className="rounded bg-gray-200 px-1.5 py-0.5 text-xs">
              http://localhost:8080/display
            </code>
            를 등록합니다.
          </li>
        </ol>
        <p className="mt-4">
          자세한 사용법은{" "}
          <a
            href="https://github.com/AwesomeYelim/easyPreparation"
            target="_blank"
            rel="noopener noreferrer"
            className="text-navy underline"
          >
            GitHub README
          </a>
          를 참고하세요.
        </p>
      </div>
    </section>
  );
}
