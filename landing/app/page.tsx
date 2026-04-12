const features = [
  {
    title: "악보 자동 표시",
    desc: "찬송가 번호만 입력하면 악보 PDF를 자동으로 다운로드하고 화면에 표시합니다. 645곡 새찬송가와 교독문을 지원합니다.",
  },
  {
    title: "OBS 방송 연동",
    desc: "OBS WebSocket으로 예배 항목별 씬 자동 전환. Display 화면을 Browser Source로 바로 송출할 수 있습니다.",
  },
  {
    title: "자동 스케줄러",
    desc: "예배 시간을 설정하면 카운트다운 후 자동으로 예배 순서를 로드하고 OBS 스트리밍을 시작합니다.",
  },
  {
    title: "모바일 리모컨",
    desc: "같은 WiFi에서 스마트폰으로 슬라이드를 제어합니다. QR 코드로 즉시 연결, PWA로 홈 화면에 추가 가능.",
  },
  {
    title: "주보 PDF 생성",
    desc: "예배 순서를 입력하면 인쇄용 주보 PDF를 자동 생성합니다. 교회소식, 광고를 포함한 레이아웃.",
  },
  {
    title: "성경 본문 조회",
    desc: "7개 번역본의 성경을 지원합니다. 3절 단위 페이징으로 화면에 깔끔하게 표시됩니다.",
  },
];

export default function Home() {
  return (
    <>
      {/* 히어로 */}
      <section className="bg-gradient-to-b from-white to-gray-50 px-6 pb-20 pt-24 text-center">
        <h1 className="text-4xl font-extrabold leading-tight tracking-tight text-navy sm:text-5xl">
          예배 준비, 이제 쉽게
        </h1>
        <p className="mx-auto mt-5 max-w-xl text-lg text-gray-600">
          찬양 악보, 주보 PDF, OBS 방송 송출까지.
          <br />
          교회 예배 준비를 하나의 도구로 해결합니다.
        </p>
        <div className="mt-8 flex flex-wrap justify-center gap-4">
          <a
            href="/download"
            className="rounded-lg bg-navy px-6 py-3 text-sm font-semibold text-white shadow-md hover:bg-navy-light"
          >
            무료 다운로드
          </a>
          <a
            href="/pricing"
            className="rounded-lg border border-gray-300 px-6 py-3 text-sm font-semibold text-gray-700 hover:bg-gray-100"
          >
            요금제 보기
          </a>
        </div>

        {/* 핵심 수치 */}
        <div className="mx-auto mt-14 grid max-w-2xl grid-cols-3 gap-6">
          <div className="rounded-xl border border-gray-200 bg-white px-6 py-5 shadow-sm">
            <p className="text-3xl font-extrabold text-navy">645곡</p>
            <p className="mt-1 text-sm text-gray-500">새찬송가 전곡 지원</p>
          </div>
          <div className="rounded-xl border border-gray-200 bg-white px-6 py-5 shadow-sm">
            <p className="text-3xl font-extrabold text-navy">7개</p>
            <p className="mt-1 text-sm text-gray-500">성경 번역본</p>
          </div>
          <div className="rounded-xl border border-gray-200 bg-white px-6 py-5 shadow-sm">
            <p className="text-3xl font-extrabold text-navy">OBS</p>
            <p className="mt-1 text-sm text-gray-500">방송 자동 연동</p>
          </div>
        </div>
      </section>

      {/* 기능 소개 */}
      <section className="mx-auto max-w-6xl px-6 py-20">
        <h2 className="text-center text-3xl font-bold text-navy">주요 기능</h2>
        <div className="mt-12 grid gap-8 sm:grid-cols-2 lg:grid-cols-3">
          {features.map((f) => (
            <div
              key={f.title}
              className="rounded-xl border border-gray-200 p-6 shadow-sm"
            >
              <h3 className="text-lg font-semibold text-navy">{f.title}</h3>
              <p className="mt-2 text-sm leading-relaxed text-gray-600">
                {f.desc}
              </p>
            </div>
          ))}
        </div>
      </section>

      {/* CTA */}
      <section className="bg-navy px-6 py-16 text-center text-white">
        <h2 className="text-2xl font-bold">지금 시작하세요</h2>
        <p className="mt-3 text-navy-light">
          Free 플랜으로 기본 기능을 모두 사용할 수 있습니다.
        </p>
        <a
          href="/download"
          className="mt-6 inline-block rounded-lg bg-white px-6 py-3 text-sm font-semibold text-navy shadow hover:bg-gray-100"
        >
          무료 다운로드
        </a>
      </section>
    </>
  );
}
