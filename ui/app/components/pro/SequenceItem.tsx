"use client";
import { WorshipOrderItem } from "@/types";
import SequenceStatusBadge from "./SequenceStatusBadge";

interface Props {
  item: WorshipOrderItem;
  index: number;
  isActive: boolean;
  onClick: () => void;
}

export default function SequenceItem({ item, index, isActive, onClick }: Props) {
  return (
    <button
      onClick={onClick}
      className={`w-full flex items-start gap-2 px-3 py-2.5 text-left transition-all border-l-2 ${
        isActive
          ? "bg-pro-accent/10 border-l-pro-accent"
          : "border-l-transparent hover:bg-pro-hover"
      }`}
    >
      {/* 인덱스 번호 */}
      <span className="text-pro-text-dim text-[10px] font-mono w-5 flex-shrink-0 mt-0.5 text-right">
        {String(index + 1).padStart(2, "0")}
      </span>
      {/* 항목 정보 */}
      <div className="flex-1 min-w-0">
        <div
          className={`text-xs font-semibold truncate ${
            isActive ? "text-pro-accent" : "text-pro-text"
          }`}
        >
          {item.title}
        </div>
        {item.obj && (
          <div className="text-[10px] text-pro-text-muted truncate mt-0.5">{item.obj}</div>
        )}
      </div>
      {/* 상태 배지 */}
      <SequenceStatusBadge status={isActive ? "live" : "program"} />
    </button>
  );
}
