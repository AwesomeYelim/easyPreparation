import Beams from "./components/live/Beams";
import Playground from "./components/live/Playground";
import ProCarousel from "./components/live/ProCarousel";
import Stories from "./components/live/Stories";
import { CountUp, Reveal } from "./components/motion";
import {
  IconBook,
  IconDoc,
  IconGear,
  IconLines,
  IconMusic,
  IconScreen,
} from "./components/icons";
import { Band, Button, Card, SectionHead, Tag } from "./components/ui";

const stats = [
  { n: 645, suffix: "곡", label: "새찬송가 악보 자동 표시" },
  { n: 7, suffix: "개", label: "성경 번역판 · 비교 모드" },
  { n: 8, suffix: "종", label: "주보 시안 템플릿" },
  { n: 3, suffix: " OS", label: "macOS · Windows · Linux" },
];

const features = [
  {
    icon: <IconDoc />,
    title: "주보 PDF · 시안 8종",
    desc: "인쇄용 A4와 프레젠테이션 PDF를 한 번에. 교회소식은 계층형으로 편집하고, 시안 템플릿은 브라우저에서 바로 인쇄합니다.",
  },
  {
    icon: <IconMusic />,
    title: "찬송가 · 성시교독 자동 표시",
    desc: "번호만 입력하면 645곡 악보와 교독문이 자동으로 내려와 화면에 뜹니다. 한 번 받은 파일은 로컬에 캐시됩니다.",
  },
  {
    icon: <IconBook />,
    title: "성경 7개 번역판 · 비교 모드",
    desc: "책·장·절을 고르면 본문이 슬라이드로. 두 번역판을 나란히 비교해 보여 줄 수도 있습니다.",
  },
  {
    icon: <IconScreen />,
    title: "예배 화면 3종",
    desc: "프로젝터용 전체 화면, 방송용 자막 오버레이, 다음 순서와 타이머가 보이는 무대 모니터를 각각 URL 하나로.",
  },
  {
    icon: <IconLines />,
    title: "가사 PPT · 커스텀 곡",
    desc: "곡 제목만 넣으면 가사를 찾아 슬라이드로 묶어 ZIP으로. 우리 교회 곡은 직접 등록해 재사용합니다.",
  },
  {
    icon: <IconGear />,
    title: "방송 자동화",
    desc: "항목마다 OBS 씬 자동 전환, 예배 시간 카운트다운 후 송출 시작, YouTube 방송·썸네일 자동 생성, PTZ 카메라 프리셋 이동.",
    pro: true,
  },
];

const roles = [
  {
    title: "미디어팀 봉사자",
    desc: "처음 온 주일에도 됩니다. 순서는 이미 만들어져 있고, 할 일은 다음 버튼을 누르는 것뿐입니다.",
  },
  {
    title: "찬양 인도자",
    desc: "찬송 번호와 곡 제목만 넘기면 악보와 가사 슬라이드가 준비됩니다. 다시 타이핑할 일이 없습니다.",
  },
  {
    title: "목회자",
    desc: "본문과 설교 제목을 넣으면 주보·화면·YouTube 설명란까지 같은 내용으로 맞춰집니다.",
  },
  {
    title: "방송팀",
    desc: "OBS·YouTube·PTZ 카메라가 순서에 맞춰 움직입니다. 주일 아침에 손댈 게 하나 줄어듭니다.",
  },
];

export default function Home() {
  return (
    <>
      {/* Hero: centered headline over light beams, playable app below */}
      <section className="relative overflow-hidden px-6 pb-24 pt-20 md:pt-[112px]">
        <Beams />
        <div className="relative mx-auto flex max-w-site flex-col items-center gap-14">
          <div className="live-fade flex max-w-[760px] flex-col items-center gap-6 text-center">
            <span className="inline-flex h-7 items-center gap-2 rounded-full border border-accent/35 bg-accent-deep/[0.12] pl-2 pr-3 text-xs font-semibold tracking-wide text-accent-light">
              <span className="h-1.5 w-1.5 rounded-full bg-accent" />
              교회 미디어팀을 위한 예배 준비 소프트웨어
            </span>
            <h1 className="text-4xl font-extrabold leading-[1.12] tracking-[-0.025em] text-white [text-shadow:0_4px_40px_rgba(0,0,0,0.5)] sm:text-5xl md:text-[64px]">
              주일 아침,{" "}
              <br className="hidden sm:block" />
              누구나 자신 있게 돌리는 예배
            </h1>
            <p className="max-w-[560px] text-lg leading-[1.7] text-muted">
              주보 PDF, 찬송 악보, 성경 슬라이드, OBS 송출까지 한 화면에서.
              아래 화면은 직접 눌러 볼 수 있습니다.
            </p>
            <div className="flex flex-wrap items-center justify-center gap-3">
              <Button href="/download" size="lg">
                무료 다운로드
              </Button>
              <Button href="/pricing" variant="ghost" size="lg">
                요금제 보기
              </Button>
            </div>
            <p className="text-[13px] text-dim">
              카드 등록 없음 · macOS / Windows / Linux · 자동 업데이트
            </p>
          </div>
          <div className="w-full">
            <Playground />
          </div>
        </div>
      </section>

      {/* Trust line */}
      <Band>
        <div className="mx-auto grid max-w-site grid-cols-2 gap-6 px-6 py-6 md:grid-cols-4">
          {stats.map((s) => (
            <div key={s.label} className="flex flex-col items-center gap-1">
              <span className="num text-[30px] font-extrabold text-white">
                <CountUp to={s.n} suffix={s.suffix} />
              </span>
              <span className="text-[13px] text-dim">{s.label}</span>
            </div>
          ))}
        </div>
      </Band>

      {/* Workflow: Cursor-style story rows */}
      <section className="mx-auto flex max-w-site flex-col gap-16 px-6 pb-24 pt-24 md:pt-[120px]">
        <div className="flex flex-col items-center text-center text-[22px] font-bold leading-[1.4] md:text-[28px]">
          <h2 className="text-white">토요일에 준비하면, 주일은 넘기기만.</h2>
          <p className="text-dim">한 번 입력한 순서가 주보 · 화면 · 방송까지 이어집니다.</p>
        </div>
        <Stories />
      </section>

      {/* Features */}
      <Band>
        <section
          id="features"
          className="mx-auto flex max-w-site scroll-mt-24 flex-col gap-12 px-6 py-24"
        >
          <SectionHead
            eyebrow="Features"
            title="필요한 건 전부, 복잡함은 없이"
            desc="기본 기능은 전부 무료입니다. 방송 자동화가 필요할 때만 Pro로 올리면 됩니다."
          />
          <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
            {features.map((f, i) => (
              <Reveal key={f.title} delay={(i % 3) * 90} className="h-full">
              <Card
                highlight={f.pro}
                className="flex h-full flex-col gap-3.5 p-7 transition duration-300 hover:-translate-y-1 hover:border-accent/40"
              >
                <div className="flex items-center justify-between">
                  {f.icon}
                  {f.pro && <Tag>PRO</Tag>}
                </div>
                <h3 className="text-lg font-bold text-white">{f.title}</h3>
                <p className="text-sm leading-[1.65] text-muted">{f.desc}</p>
              </Card>
              </Reveal>
            ))}
          </div>
        </section>
      </Band>

      {/* Pro automation carousel */}
      <section className="mx-auto flex max-w-site flex-col gap-12 px-6 py-24 md:py-[104px]">
        <div className="flex flex-col items-center text-center text-[22px] font-bold leading-[1.4] md:text-[28px]">
          <h2 className="text-white">방송하는 교회라면, Pro.</h2>
          <p className="text-dim">예배 순서가 OBS · YouTube · 카메라를 움직입니다.</p>
        </div>
        <ProCarousel />
      </section>


      <section className="mx-auto flex max-w-site flex-col gap-12 px-6 py-24 md:py-[104px]">
        <SectionHead
          eyebrow="For every role"
          title="팀의 모든 역할을 위해"
          desc="한 사람에게 모든 걸 맡기지 않아도 됩니다. 각자 자기 몫만 하면 주일이 굴러갑니다."
        />
        <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
          {roles.map((r, i) => (
            <Reveal key={r.title} delay={i * 90} className="h-full">
            <Card className="flex h-full flex-col gap-3 p-7 transition duration-300 hover:-translate-y-1 hover:border-accent/40">
              <span className="h-1 w-8 rounded-full bg-accent" />
              <h3 className="text-lg font-bold text-white">{r.title}</h3>
              <p className="text-sm leading-[1.65] text-muted">{r.desc}</p>
            </Card>
            </Reveal>
          ))}
        </div>
      </section>

      {/* Pricing teaser */}
      <Band>
        <section className="mx-auto flex max-w-[960px] flex-col gap-11 px-6 py-24">
          <SectionHead
            center
            eyebrow="Pricing"
            title="교회 예산에 맞는 요금"
            desc="기본 기능은 전부 무료. 방송 자동화만 Pro입니다."
          />
          <div className="grid gap-5 md:grid-cols-2">
            <Card className="flex flex-col gap-3.5 p-8">
              <h3 className="text-lg font-bold text-muted">Free</h3>
              <p className="num text-[34px] font-extrabold text-white">무료</p>
              <p className="text-sm leading-[1.65] text-muted">
                주보 PDF · 찬송/교독 악보 · 성경 · 예배 화면 3종 · 가사 PPT ·
                모바일 리모컨 · 자동 업데이트
              </p>
              <Button href="/download" variant="ghost" className="mt-auto">
                무료 다운로드
              </Button>
            </Card>
            <Card highlight className="relative flex flex-col gap-3.5 p-8">
              <span className="absolute -top-3 left-8 inline-flex h-6 items-center rounded bg-accent px-2.5 text-[11px] font-extrabold tracking-wide text-bg">
                방송하는 교회에 추천
              </span>
              <h3 className="text-lg font-bold text-accent-light">Pro</h3>
              <p className="num text-[34px] font-extrabold text-white">
                월 9,900원{" "}
                <span className="text-sm font-medium text-dim">
                  · 연 99,000원 (2개월 무료)
                </span>
              </p>
              <p className="text-sm leading-[1.65] text-muted">
                Free 기능 전부, 그리고 OBS 씬 자동 전환 · 자동 스케줄러 ·
                YouTube 방송 자동 생성 · 썸네일 · PTZ 카메라
              </p>
              <Button href="/pricing" className="mt-auto">
                Pro 자세히 보기
              </Button>
            </Card>
          </div>
        </section>
      </Band>

      {/* Final CTA */}
      <section className="flex flex-col items-center gap-5 bg-[radial-gradient(800px_300px_at_50%_120%,rgba(79,140,201,0.25),transparent_70%)] px-6 py-24 text-center">
        <h2 className="text-3xl font-extrabold tracking-[-0.02em] text-white md:text-[40px]">
          덜 긴장하고, 더 좋은 주일을.
        </h2>
        <p className="text-base text-muted">
          설치 한 번, 카드 등록 없이 오늘 바로 시작할 수 있습니다.
        </p>
        <Button href="/download" size="lg" className="px-7">
          무료 다운로드
        </Button>
      </section>
    </>
  );
}
