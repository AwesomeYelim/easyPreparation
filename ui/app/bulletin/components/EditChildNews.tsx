import React, { useEffect } from "react";
import { WorshipOrderItem } from "@/types";
import { findNode } from "@/lib/treeUtils";
import s from "../bulletin.module.scss";

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
    <div className={s.form_group}>
      <label htmlFor="obj">
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
      />
    </div>
  );
}
