const RELEASE_URL =
  "https://github.com/AwesomeYelim/easyPreparation/releases/latest";

const platforms = [
  {
    name: "macOS (Apple Silicon)",
    desc: "M1 / M2 / M3 / M4",
    file: "easyPreparation_darwin_arm64",
    icon: "\u{f8ff}",
  },
  {
    name: "macOS (Intel)",
    desc: "Intel Mac",
    file: "easyPreparation_darwin_amd64",
    icon: "\u{f8ff}",
  },
  {
    name: "Windows",
    desc: "Windows 10 / 11 (64-bit)",
    file: "easyPreparation_windows_amd64.exe",
    icon: "W",
  },
  {
    name: "Linux",
    desc: "Ubuntu / Debian / Fedora (64-bit)",
    file: "easyPreparation_linux_amd64",
    icon: "L",
  },
];

export default function DownloadPage() {
  return (
    <section className="mx-auto max-w-4xl px-6 py-20">
      <h1 className="text-center text-3xl font-bold text-navy">다운로드</h1>
      <p className="mt-3 text-center text-gray-600">
        운영체제에 맞는 버전을 선택하세요. 모두 무료입니다.
      </p>

      <div className="mt-12 grid gap-6 sm:grid-cols-2">
        {platforms.map((p) => (
          <a
            key={p.file}
            href={RELEASE_URL}
            target="_blank"
            rel="noopener noreferrer"
            className="group flex items-center gap-4 rounded-xl border border-gray-200 p-5 shadow-sm hover:border-navy hover:shadow-md"
          >
            <span className="flex h-12 w-12 items-center justify-center rounded-lg bg-gray-100 text-lg font-bold text-navy group-hover:bg-navy group-hover:text-white">
              {p.icon}
            </span>
            <div>
              <p className="font-semibold text-gray-900">{p.name}</p>
              <p className="text-sm text-gray-500">{p.desc}</p>
            </div>
          </a>
        ))}
      </div>

      <div className="mt-12 rounded-xl border border-gray-200 bg-gray-50 p-6 text-sm text-gray-600">
        <h2 className="font-semibold text-gray-900">설치 후 시작하기</h2>
        <ol className="mt-3 list-inside list-decimal space-y-1.5">
          <li>다운로드한 파일을 실행합니다.</li>
          <li>
            브라우저에서{" "}
            <code className="rounded bg-gray-200 px-1.5 py-0.5 text-xs">
              http://localhost:8080
            </code>{" "}
            에 접속합니다.
          </li>
          <li>예배 순서를 입력하고 Display 화면을 열어 확인합니다.</li>
          <li>
            OBS Browser Source에{" "}
            <code className="rounded bg-gray-200 px-1.5 py-0.5 text-xs">
              http://localhost:8080/display
            </code>{" "}
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
