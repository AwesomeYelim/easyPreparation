"use client";

import { useEffect, useRef, useState } from "react";
import Stage from "./Stage";
import { useTicker } from "./hooks";
import { BulletinMock, ConsoleMock, DisplayMock, RemoteMock } from "./mocks";

const STEP_SECONDS = 12;
const PERIOD = 3;
// Display tab walks through slides that show off each kind of output.
const displayOrder = [13, 1, 8, 3];

type Step = { n: string; title: string; desc: string };

export default function WorkflowTabs({ steps }: { steps: Step[] }) {
  const ref = useRef<HTMLDivElement>(null);
  const t = useTicker(ref);
  const [tab, setTab] = useState(0);
  const [since, setSince] = useState(0);
  const [manual, setManual] = useState(false);
  const local = Math.max(0, t - since);

  useEffect(() => {
    if (!manual && local >= STEP_SECONDS) {
      setTab((v) => (v + 1) % steps.length);
      setSince(t);
    }
  }, [local, manual, t, steps.length]);

  const pick = (i: number) => {
    setTab(i);
    setSince(t);
    setManual(true);
  };

  const beat = Math.floor(local / PERIOD);
  const panels = [
    <BulletinMock key="b" step={local * 2} />,
    <DisplayMock key="d" index={displayOrder[beat % displayOrder.length]} t={local} />,
    <ConsoleMock key="c" index={4 + beat} t={local} period={PERIOD} clockBase={11 * 3600 + 3 * 60} />,
    <RemoteMock key="r" index={9 + beat} />,
  ];

  return (
    <div ref={ref} className="grid items-start gap-6 lg:grid-cols-[340px_1fr]">
      <div role="tablist" aria-label="Workflow" className="grid grid-cols-2 gap-2 lg:grid-cols-1">
        {steps.map((s, i) => {
          const on = i === tab;
          return (
            <button
              key={s.n}
              role="tab"
              aria-selected={on}
              onClick={() => pick(i)}
              className={`relative overflow-hidden rounded-card border p-3.5 text-left transition-colors lg:p-5 ${on ? "border-accent/50 bg-accent-deep/[0.12]" : "border-line bg-surface hover:border-line-strong"}`}
            >
              <div className="flex items-center gap-2.5">
                <span className="num text-xs font-bold tracking-[0.1em] text-accent">{s.n}</span>
                <h3 className={`text-[15px] font-bold lg:text-[17px] ${on ? "text-white" : "text-muted"}`}>{s.title}</h3>
              </div>
              {on && <p className="mt-2 hidden text-sm leading-[1.65] text-muted lg:block">{s.desc}</p>}
              {on && !manual && (
                <span
                  key={`${tab}-${since}`}
                  className="live-progress absolute bottom-0 left-0 h-[2px] bg-accent"
                  style={{ animationDuration: `${STEP_SECONDS}s` }}
                />
              )}
            </button>
          );
        })}
      </div>
      <div className="overflow-hidden rounded-card border border-white/[0.12] bg-surface shadow-[0_30px_60px_rgba(0,0,0,0.45)]">
        <div className="flex h-[34px] items-center gap-1.5 border-b border-line bg-raised px-3">
          <span className="h-2.5 w-2.5 rounded-full bg-[#2a3a55]" />
          <span className="h-2.5 w-2.5 rounded-full bg-[#2a3a55]" />
          <span className="h-2.5 w-2.5 rounded-full bg-[#2a3a55]" />
          <span className="ml-3 text-xs text-dim">{steps[tab].title}</span>
        </div>
        <Stage w={1200} h={750}>
          <div key={tab} className="live-fade h-full">
            {panels[tab]}
          </div>
        </Stage>
      </div>
      <p key={tab} className="live-fade -mt-2 text-sm leading-[1.65] text-muted lg:hidden">
        {steps[tab].desc}
      </p>
    </div>
  );
}
