import type { ReactNode } from "react";

type ButtonProps = {
  href: string;
  children: ReactNode;
  variant?: "primary" | "ghost";
  size?: "md" | "lg";
  external?: boolean;
  className?: string;
};

export function Button({
  href,
  children,
  variant = "primary",
  size = "md",
  external,
  className = "",
}: ButtonProps) {
  const base =
    "inline-flex items-center justify-center rounded-btn font-bold transition";
  const sizes = size === "lg" ? "h-12 px-6 text-[15px]" : "h-11 px-5 text-sm";
  const variants =
    variant === "primary"
      ? "bg-white text-bg hover:bg-accent-light"
      : "border border-line-strong bg-white/[0.03] text-ink font-semibold hover:border-accent hover:text-white";
  return (
    <a
      href={href}
      className={`${base} ${sizes} ${variants} ${className}`}
      {...(external ? { target: "_blank", rel: "noopener noreferrer" } : {})}
    >
      {children}
    </a>
  );
}

export function Eyebrow({ children }: { children: ReactNode }) {
  return (
    <span className="text-xs font-bold uppercase tracking-[0.14em] text-accent">
      {children}
    </span>
  );
}

export function SectionHead({
  eyebrow,
  title,
  desc,
  center,
}: {
  eyebrow?: string;
  title: string;
  desc?: string;
  center?: boolean;
}) {
  return (
    <div
      className={`flex max-w-[720px] flex-col gap-3.5 ${
        center ? "items-center self-center text-center" : ""
      }`}
    >
      {eyebrow && <Eyebrow>{eyebrow}</Eyebrow>}
      <h2 className="text-3xl font-extrabold tracking-[-0.02em] text-white md:text-[38px] md:leading-tight">
        {title}
      </h2>
      {desc && (
        <p className="text-base leading-relaxed text-muted md:text-[17px]">
          {desc}
        </p>
      )}
    </div>
  );
}

export function Card({
  children,
  highlight,
  className = "",
}: {
  children: ReactNode;
  highlight?: boolean;
  className?: string;
}) {
  const tone = highlight
    ? "border-accent/50 bg-[linear-gradient(180deg,rgba(79,140,201,0.14),rgba(79,140,201,0.03))] bg-surface"
    : "border-line bg-surface";
  return (
    <div className={`rounded-card border ${tone} ${className}`}>{children}</div>
  );
}

export function Tag({ children }: { children: ReactNode }) {
  return (
    <span className="inline-flex h-[22px] items-center rounded px-2 text-[11px] font-extrabold tracking-[0.06em] text-bg bg-accent">
      {children}
    </span>
  );
}

export function Check() {
  return (
    <svg
      className="mt-0.5 h-[18px] w-[18px] flex-none"
      viewBox="0 0 24 24"
      fill="none"
      stroke="#7fb3e6"
      strokeWidth="2.2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <path d="M5 12l5 5L20 7" />
    </svg>
  );
}

export function CheckList({ items }: { items: string[] }) {
  return (
    <ul className="flex flex-col gap-3 text-[15px] text-list">
      {items.map((t) => (
        <li key={t} className="flex items-start gap-2.5">
          <Check />
          <span>{t}</span>
        </li>
      ))}
    </ul>
  );
}

export function Code({ children }: { children: ReactNode }) {
  return (
    <code className="rounded bg-raised px-1.5 py-0.5 font-mono text-[13px] text-accent-light">
      {children}
    </code>
  );
}

export function Band({
  children,
  className = "",
}: {
  children: ReactNode;
  className?: string;
}) {
  return (
    <div className={`border-y border-line bg-band ${className}`}>{children}</div>
  );
}
