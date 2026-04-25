"use client";

import { useState, useEffect } from "react";

interface TourStep {
  icon: string;
  title: string;
  description: string;
}

const TOUR_STEPS: TourStep[] = [
  {
    icon: "waving_hand",
    title: "easyPreparation에 오신 것을 환영합니다",
    description:
      "예배 슬라이드를 5분 안에 프로젝터에 띄울 수 있도록 안내해드릴게요. 간단한 투어를 시작해볼까요?",
  },
  {
    icon: "queue_music",
    title: "예배 순서 패널",
    description:
      "왼쪽 패널에서 예배 순서를 구성합니다. 찬양, 기도, 성경 봉독 등 각 항목을 추가하고 순서를 조정할 수 있어요.",
  },
  {
    icon: "edit_document",
    title: "주보 편집",
    description:
      "중앙에서 예배 정보를 입력합니다. 상단 탭에서 Bulletin 탭을 선택하면 교회 정보와 예배 순서 세부 내용을 편집할 수 있어요.",
  },
  {
    icon: "cast",
    title: "프로젝터 전송",
    description:
      "'프로젝터에 보내기' 버튼으로 화면에 표시합니다. OBS와 연동되어 실시간으로 슬라이드가 방송 화면에 나타나요.",
  },
  {
    icon: "keyboard",
    title: "슬라이드 조작",
    description:
      "Space / → 키로 다음 슬라이드, ← 키로 이전 슬라이드로 이동합니다. 예배 중 키보드 하나로 간편하게 제어할 수 있어요.",
  },
];

export default function WelcomeTour() {
  const [show, setShow] = useState(false);
  const [step, setStep] = useState(0);

  useEffect(() => {
    if (!localStorage.getItem("ep_tour_done")) {
      setShow(true);
    }

    window.__resetTour = () => {
      localStorage.removeItem("ep_tour_done");
      window.location.reload();
    };

    return () => {
      delete window.__resetTour;
    };
  }, []);

  function finish() {
    localStorage.setItem("ep_tour_done", "1");
    setShow(false);
  }

  function handleNext() {
    if (step < TOUR_STEPS.length - 1) {
      setStep((s) => s + 1);
    } else {
      finish();
    }
  }

  function handleSkip() {
    finish();
  }

  if (!show) return null;

  const current = TOUR_STEPS[step];
  const isLast = step === TOUR_STEPS.length - 1;

  return (
    <>
      {/* 배경 오버레이 — 클릭 불가 */}
      <div
        className="fixed inset-0 z-[49999] bg-black/50"
        aria-hidden="true"
      />

      {/* 투어 카드 */}
      <div
        role="dialog"
        aria-modal="true"
        aria-label="환영 투어"
        className="fixed bottom-8 left-1/2 -translate-x-1/2 z-[50000] w-[480px] max-w-[calc(100vw-2rem)]"
      >
        <div className="bg-[#1a1a2e] border border-[#3B82F6]/30 rounded-2xl shadow-2xl p-6 flex flex-col gap-5">
          {/* 스텝 닷 인디케이터 */}
          <div className="flex items-center justify-center gap-2">
            {TOUR_STEPS.map((_, i) => (
              <span
                key={i}
                className={
                  i === step
                    ? "w-2.5 h-2.5 rounded-full bg-[#3B82F6]"
                    : "w-2 h-2 rounded-full border border-[#3B82F6]/50"
                }
                aria-hidden="true"
              />
            ))}
          </div>

          {/* 아이콘 + 제목 + 설명 */}
          <div className="flex flex-col items-center gap-3 text-center">
            <span
              className="material-symbols-outlined text-[#3B82F6] select-none"
              style={{ fontSize: 48 }}
              aria-hidden="true"
            >
              {current.icon}
            </span>
            <h2 className="text-[#f1f5f9] font-bold text-lg leading-snug">
              {current.title}
            </h2>
            <p className="text-[#94a3b8] text-sm leading-relaxed">
              {current.description}
            </p>
          </div>

          {/* 버튼 행 */}
          <div className="flex items-center justify-between pt-1">
            <button
              onClick={handleSkip}
              className="text-[#94a3b8] text-sm hover:text-[#f1f5f9] transition-colors px-2 py-1 rounded-lg focus:outline-none focus:ring-2 focus:ring-[#3B82F6]/50"
            >
              건너뛰기
            </button>

            <button
              onClick={handleNext}
              className="flex items-center gap-1.5 bg-[#3B82F6] hover:bg-[#2563eb] active:bg-[#1d4ed8] text-white text-sm font-semibold px-5 py-2 rounded-xl transition-colors focus:outline-none focus:ring-2 focus:ring-[#3B82F6]/70"
            >
              {isLast ? (
                <>
                  시작하기
                  <span className="material-symbols-outlined text-base select-none" style={{ fontSize: 18 }} aria-hidden="true">
                    check
                  </span>
                </>
              ) : (
                <>
                  다음
                  <span className="material-symbols-outlined text-base select-none" style={{ fontSize: 18 }} aria-hidden="true">
                    arrow_forward
                  </span>
                </>
              )}
            </button>
          </div>
        </div>
      </div>
    </>
  );
}
