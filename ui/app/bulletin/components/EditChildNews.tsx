import React, { useEffect } from "react";
import { WorshipOrderItem } from "@/types";
import { findNode } from "@/lib/treeUtils";

interface EditChildNewsProps {
  selectedDetail: WorshipOrderItem;
  selectedChild: WorshipOrderItem;
  setSelectedChild: React.Dispatch<React.SetStateAction<WorshipOrderItem | null>>;
  handleValueChange: (
    key: string,
    { newObj, newLead }: { newObj: string; newLead?: string }
  ) => void;
}

export default function EditChildNews({
  selectedDetail,
  selectedChild,
  setSelectedChild,
  handleValueChange,
}: EditChildNewsProps) {
  useEffect(() => {
    const matched = findNode(selectedDetail.children, selectedChild.key);
    if (matched) setSelectedChild(matched);
  }, [selectedDetail, selectedChild.key]); // key 기준으로 찾기

  return (
    <div className="flex flex-col gap-1.5 mt-4">
      <label htmlFor="obj" className="text-xs font-bold text-pro-text-dim uppercase tracking-wider">
        Content
      </label>
      <textarea
        id="obj"
        value={selectedChild.obj}
        onChange={(e) => {
          const newObj = e.target.value;
          setSelectedChild((prev) => prev ? { ...prev, obj: newObj } : prev);
          handleValueChange(selectedChild.key, { newObj });
        }}
        placeholder="Enter news content"
        className="w-full px-3 py-2.5 border border-pro-border bg-pro-elevated text-pro-text rounded-lg text-sm resize-none min-h-[120px] focus:outline-none focus:border-pro-accent focus:ring-2 focus:ring-pro-accent/20 transition-all"
      />
    </div>
  );
}
