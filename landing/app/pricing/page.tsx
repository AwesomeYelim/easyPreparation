import type { Metadata } from "next";
import { Button, Card, CheckList } from "../components/ui";
import { ISSUES } from "../lib/links";

export const metadata: Metadata = {
  title: "요금제 — easyPreparation",
  description:
    "주보·악보·성경·예배 화면은 Free로 끝까지. 방송 자동화만 Pro. 카드 등록 없이 시작합니다.",
};

const freeFeatures = [
  "주보 PDF 생성 (인쇄용 + 프레젠테이션) · 시안 템플릿 8종",
  "찬송가 645곡 악보 · 성시교독 자동 표시",
  "성경 7개 번역판 조회 · 비교 모드",
  "예배 화면 3종 (프로젝터 · 방송 오버레이 · 무대 모니터)",
  "가사 PPT(ZIP) 생성 · 커스텀 곡 관리",
  "모바일 리모컨 (PWA) · 자동 업데이트",
];

const proFeatures = [
  "OBS 씬 자동 전환 · 스트리밍 시작/중지 제어",
  "자동 스케줄러 — 예배 시간 카운트다운 후 순서 로드 · 송출 시작",
  "YouTube 라이브 방송 자동 생성 · 스트림 키 OBS 자동 설정",
  "썸네일 자동 생성 · 업로드 (설교 제목 · 본문 자동 반영)",
  "PTZ 카메라 프리셋 자동 이동",
];

const faqs = [
  {
    q: "Free는 정말 무료인가요?",
    a: "네. 기간·워터마크·기능 잠금 없이 계속 쓸 수 있습니다. Pro는 방송 자동화 기능에만 필요합니다.",
  },
  {
    q: "결제는 어디서 하나요?",
    a: "앱 안의 라이선스 화면에서 결제하면 라이선스 키가 발급됩니다. 키는 PC 한 대에 등록되고, 오프라인에서도 30일간 유지됩니다.",
  },
  {
    q: "OBS는 어떤 버전이 필요하나요?",
    a: "WebSocket 서버가 내장된 OBS 28 이상을 권장합니다. 별도 플러그인은 필요 없습니다.",
  },
  {
    q: "설치 후 바로 방송할 수 있나요?",
    a: "Display 화면은 Free에서 바로 OBS Browser Source로 송출됩니다. 씬 자동 전환·YouTube 방송 생성 같은 자동화만 Pro에서 켜집니다.",
  },
];

export default function PricingPage() {
  return (
    <section className="mx-auto flex max-w-[1024px] flex-col gap-14 px-6 pb-24 pt-20">
      <div className="flex flex-col items-center gap-3.5 text-center">
        <h1 className="text-3xl font-extrabold tracking-[-0.02em] text-white md:text-[44px] md:leading-tight">
          교회 예산에 맞는 요금. 기본은 전부 무료.
        </h1>
        <p className="max-w-[620px] text-base leading-[1.7] text-muted md:text-[17px]">
          주보·악보·성경·예배 화면은 Free로 끝까지 쓸 수 있습니다. 방송을
          자동화하고 싶을 때만 Pro를 켜세요. 카드 등록 없이 시작합니다.
        </p>
      </div>

      <div className="grid items-stretch gap-7 md:grid-cols-2">
        <Card className="flex flex-col gap-5 p-9">
          <div className="flex flex-col gap-1.5">
            <h2 className="text-lg font-bold text-accent-light">Free</h2>
            <p className="num text-[34px] font-extrabold text-white">무료</p>
            <p className="text-sm text-dim">기간 제한 없음 · 워터마크 없음</p>
          </div>
          <CheckList items={freeFeatures} />
          <Button href="/download" variant="ghost" size="lg" className="mt-auto">
            무료 다운로드
          </Button>
        </Card>

        <Card highlight className="relative flex flex-col gap-5 p-9">
          <span className="absolute -top-[13px] left-9 inline-flex h-[26px] items-center rounded bg-accent px-3 text-[11px] font-extrabold text-bg">
            방송하는 교회에 추천
          </span>
          <div className="flex flex-col gap-1.5">
            <h2 className="text-lg font-bold text-accent-light">Pro</h2>
            <p className="num text-[34px] font-extrabold text-white">월 9,900원</p>
            <p className="text-sm text-dim">
              연 99,000원 (2개월 무료) · 앱 안에서 결제 · 언제든 해지
            </p>
          </div>
          <p className="text-sm font-bold text-accent-light">
            Free 기능 전부, 그리고
          </p>
          <CheckList items={proFeatures} />
          <Button href="/download" size="lg" className="mt-auto">
            앱 설치 후 Pro 시작하기
          </Button>
        </Card>
      </div>

      <div className="flex flex-col gap-6 rounded-card border border-line bg-band px-8 py-7 md:flex-row md:items-center md:justify-between">
        <div className="flex flex-col gap-1.5">
          <h3 className="text-lg font-bold text-white">
            Enterprise — 여러 캠퍼스·노회 단위
          </h3>
          <p className="text-sm text-muted">
            다수 교회 라이선스와 도입 지원이 필요하면 따로 상담합니다.
          </p>
        </div>
        <Button href={ISSUES} variant="ghost" external className="whitespace-nowrap">
          문의하기
        </Button>
      </div>

      <div className="flex flex-col gap-6">
        <h2 className="text-center text-[26px] font-extrabold text-white">
          자주 묻는 질문
        </h2>
        <div className="grid gap-5 md:grid-cols-2">
          {faqs.map((f) => (
            <Card key={f.q} className="flex flex-col gap-2 p-6">
              <h3 className="text-base font-bold text-white">{f.q}</h3>
              <p className="text-sm leading-[1.65] text-muted">{f.a}</p>
            </Card>
          ))}
        </div>
      </div>
    </section>
  );
}
