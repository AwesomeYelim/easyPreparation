const plans = [
  {
    name: "Free",
    price: "무료",
    desc: "기본 기능으로 시작",
    features: [
      "찬양 악보 Display",
      "주보 PDF 생성",
      "성경 본문 조회 (7개 번역본)",
      "교독문 표시",
      "모바일 리모컨 (PWA)",
    ],
    cta: "무료 다운로드",
    href: "/download",
    highlight: false,
  },
  {
    name: "Pro",
    price: "월 9,900원",
    priceAnnual: "연 99,000원 (2개월 무료)",
    desc: "방송과 자동화까지",
    features: [
      "Free 기능 전부 포함",
      "OBS 씬 자동 전환",
      "자동 스케줄러 (예배 시간 감지)",
      "OBS 스트리밍 자동 시작/중지",
      "YouTube 방송 자동 생성",
      "썸네일 자동 생성 + 업로드",
      "다중 예배 타입 지원",
    ],
    cta: "Pro 시작하기",
    href: "/download",
    highlight: true,
  },
];

const comparison = [
  { feature: "찬양 악보 Display", free: true, pro: true },
  { feature: "주보 PDF 생성", free: true, pro: true },
  { feature: "성경 본문 조회", free: true, pro: true },
  { feature: "교독문 표시", free: true, pro: true },
  { feature: "모바일 리모컨 (PWA)", free: true, pro: true },
  { feature: "OBS 씬 자동 전환", free: false, pro: true },
  { feature: "자동 스케줄러", free: false, pro: true },
  { feature: "OBS 스트리밍 제어", free: false, pro: true },
  { feature: "YouTube 방송 자동 생성", free: false, pro: true },
  { feature: "썸네일 자동 생성", free: false, pro: true },
  { feature: "다중 예배 타입", free: false, pro: true },
];

export default function PricingPage() {
  return (
    <section className="mx-auto max-w-5xl px-6 py-20">
      <h1 className="text-center text-3xl font-bold text-navy">요금제</h1>
      <p className="mt-3 text-center text-gray-600">
        기본 기능은 무료입니다. 방송 자동화가 필요하면 Pro를 선택하세요.
      </p>

      {/* 플랜 카드 */}
      <div className="mt-12 grid gap-8 sm:grid-cols-2">
        {plans.map((p) => (
          <div
            key={p.name}
            className={`rounded-xl border p-8 shadow-sm ${
              p.highlight
                ? "border-navy bg-navy/5 ring-2 ring-navy"
                : "border-gray-200"
            }`}
          >
            <h2 className="text-xl font-bold text-navy">{p.name}</h2>
            <p className="mt-1 text-2xl font-extrabold">{p.price}</p>
            {p.priceAnnual && (
              <p className="mt-1 text-sm text-gray-500">{p.priceAnnual}</p>
            )}
            <p className="mt-2 text-sm text-gray-600">{p.desc}</p>
            <ul className="mt-6 space-y-2 text-sm text-gray-700">
              {p.features.map((f) => (
                <li key={f} className="flex items-start gap-2">
                  <span className="mt-0.5 text-navy">&#10003;</span>
                  {f}
                </li>
              ))}
            </ul>
            <a
              href={p.href}
              className={`mt-8 block rounded-lg py-2.5 text-center text-sm font-semibold ${
                p.highlight
                  ? "bg-navy text-white hover:bg-navy-light"
                  : "border border-gray-300 text-gray-700 hover:bg-gray-100"
              }`}
            >
              {p.cta}
            </a>
          </div>
        ))}
      </div>

      {/* 기능 비교 테이블 */}
      <h2 className="mt-20 text-center text-2xl font-bold text-navy">
        기능 비교
      </h2>
      <div className="mt-8 overflow-x-auto">
        <table className="w-full text-left text-sm">
          <thead>
            <tr className="border-b border-gray-200 text-gray-500">
              <th className="py-3 pr-4 font-medium">기능</th>
              <th className="w-24 py-3 text-center font-medium">Free</th>
              <th className="w-24 py-3 text-center font-medium">Pro</th>
            </tr>
          </thead>
          <tbody>
            {comparison.map((row) => (
              <tr
                key={row.feature}
                className="border-b border-gray-100 hover:bg-gray-50"
              >
                <td className="py-3 pr-4 text-gray-700">{row.feature}</td>
                <td className="py-3 text-center">
                  {row.free ? (
                    <span className="text-navy">&#10003;</span>
                  ) : (
                    <span className="text-gray-300">&#8212;</span>
                  )}
                </td>
                <td className="py-3 text-center">
                  {row.pro ? (
                    <span className="text-navy">&#10003;</span>
                  ) : (
                    <span className="text-gray-300">&#8212;</span>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
