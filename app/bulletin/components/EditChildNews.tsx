import React, { useEffect } from "react";
import { WorshipOrderItem } from "@/types";
import { findNode } from "@/lib/treeUtils";

interface EditChildNewsProps {
  selectedDetail: WorshipOrderItem;
  selectedChild: WorshipOrderItem;
  setSelectedChild: React.Dispatch<React.SetStateAction<WorshipOrderItem>>;
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
    <div className="form-group">
      <label htmlFor="obj" className="form-label">
        Content
      </label>
      <textarea
        id="obj"
        className="form-textarea"
        value={selectedChild.obj}
        onChange={(e) => {
          const newObj = e.target.value;
          setSelectedChild((prev) => ({ ...prev, obj: newObj }));
          handleValueChange(selectedChild.key, { newObj });
        }}
        placeholder="Enter news content"
      />
    </div>
  );
}
