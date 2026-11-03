"use client";

import { useRef } from "react";
import Stage from "./Stage";
import { useTicker } from "./hooks";
import { DockMock, PhoneMock } from "./mocks";
import { Canvas, Row, Window } from "./Stories";

// "어디서든 진행" — the detachable sequence panel (desktop) and the phone remote.

function DockDemo() {
  const ref = useRef<HTMLDivElement>(null);
  const t = useTicker(ref);
  return (
    <div ref={ref}>
      <Window title="easyPreparation — 데스크톱 앱">
        <Stage w={1280} h={720}>
          <DockMock t={t} />
        </Stage>
      </Window>
    </div>
  );
}

const remoteFeatures = [
  "미리보기 · LIVE 버튼 위쪽 고정",
  "탭해서 점프 · 섹션 점프",
  "손잡이로 순서 바꾸기",
  "좌우 스와이프로 넘기기",
  "자동 타이머 · 예약 카운트다운",
  "홈 화면에 추가하면 앱처럼",
];

function QrCard() {
  // decorative QR-like grid (not a real code)
  const cells = Array.from({ length: 81 }, (_, i) => ((i * 37 + (i % 7) * 11) % 5 < 2 ? 1 : 0));
  return (
    <div className="flex flex-col items-center gap-2 rounded-2xl bg-white p-3 shadow-[0_20px_40px_rgba(0,0,0,0.45)]">
      <div className="grid grid-cols-9 gap-[2px]">
        {cells.map((c, i) => (
          <span key={i} className={`h-2 w-2 ${c ? "bg-[#0b1220]" : "bg-transparent"}`} />
        ))}
      </div>
      <span className="text-[10px] font-semibold text-[#43474e]">제어판 QR</span>
    </div>
  );
}

function RemoteDemo() {
  const ref = useRef<HTMLDivElement>(null);
  const t = useTicker(ref);
  const index = 3 + Math.floor(t / 2);
  const live = t % 12 >= 6;
  return (
    <div ref={ref} className="flex items-center justify-center gap-5 sm:gap-8">
      <div className="hidden flex-col items-center gap-3 sm:flex">
        <QrCard />
        <span className="live-wave text-[22px] text-[#7fb3e6]">)))</span>
        <span className="text-[12px] text-[#cbd5e1]">같은 Wi-Fi</span>
      </div>
      <div className="w-[230px] shrink-0 overflow-hidden rounded-[30px] border-[6px] border-[#1c2740] shadow-[0_24px_48px_rgba(0,0,0,0.6)] sm:w-[260px]">
        <Stage w={390} h={844}>
          <PhoneMock index={index} live={live} />
        </Stage>
      </div>
    </div>
  );
}

export default function Anywhere() {
  return (
    <div className="flex flex-col gap-24 md:gap-32">
      <Row
        title="순서 패널, 떼어서 원하는 모니터로."
        sub="OBS 독처럼 제목줄을 잡고 끌면 예배 순서가 별도 창으로 떨어집니다. 미리보기가 함께 붙어 있어 보조 모니터 하나로 진행이 끝나고, 다시 끌어다 놓으면 제자리로 돌아갑니다. 위치는 다음 실행에도 그대로."
      >
        <DockDemo />
      </Row>

      <Row
        flip
        title="폰이 리모컨이 됩니다. 앱 설치 없이."
        sub="제어판의 QR을 찍으면 끝. 같은 Wi-Fi라면 누구의 폰이든 미리보기를 보면서 넘기고, 방송을 켜고, 순서를 바꿉니다. 강대상 옆에서도, 방송석 뒤에서도."
      >
        <Canvas art="morning">
          <div className="flex flex-col gap-6">
            <RemoteDemo />
            <ul className="mx-auto grid max-w-[460px] grid-cols-1 gap-x-6 gap-y-2 text-[14px] text-white sm:grid-cols-2">
              {remoteFeatures.map((f) => (
                <li key={f} className="flex items-center gap-2 [text-shadow:0_1px_8px_rgba(0,0,0,0.6)]">
                  <span className="text-[#6ee7b7]">✓</span>
                  {f}
                </li>
              ))}
            </ul>
          </div>
        </Canvas>
      </Row>
    </div>
  );
}
