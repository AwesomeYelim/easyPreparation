"use client";

type Status = "live" | "program" | "draft" | "preview";

interface Props {
  status: Status;
}

const BADGE_CONFIG: Record<Status, { label: string; className: string }> = {
  live:    { label: "LIVE",  className: "bg-pro-live/20 text-pro-live border border-pro-live/40" },
  program: { label: "PGM",   className: "bg-pro-program/10 text-pro-program border border-pro-program/20" },
  preview: { label: "PREV",  className: "bg-pro-preview/20 text-pro-preview border border-pro-preview/40" },
  draft:   { label: "DRAFT", className: "bg-pro-border/30 text-pro-text-dim border border-pro-border" },
};

export default function SequenceStatusBadge({ status }: Props) {
  const { label, className } = BADGE_CONFIG[status];
  return (
    <span className={`flex-shrink-0 text-[8px] font-bold px-1 py-0.5 rounded leading-none ${className}`}>
      {label}
    </span>
  );
}
