import React, { useEffect } from "react";
import { WorshipOrderItem } from "../page";

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
    const findMatchingChild = (
      items: WorshipOrderItem[] | undefined,
      key: string
    ): WorshipOrderItem | null => {
      if (!items) return null;
      for (const item of items) {
        if (item.key === key) return item;
        if (item.children) {
          const found = findMatchingChild(item.children, key);
          if (found) return found;
        }
      }
      return null;
    };

    const matched = findMatchingChild(
      selectedDetail.children,
      selectedChild.key
    );
    if (matched) {
      setSelectedChild(matched);
    }
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
