import React from "react";
import { WorshipOrderItem } from "../page";

interface EditChildNewsProps {
  handleValueChange: (key: string, { newObj, newLead }: { newObj: string; newLead?: string }) => void;
  selectedChild: WorshipOrderItem;
  setSelectedChild: React.Dispatch<React.SetStateAction<WorshipOrderItem>>;
}
export default function EditChildNews({
  selectedChild,
  setSelectedChild,
  handleValueChange, // handleValueChange를 부모에서 전달받음
}: EditChildNewsProps) {
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
          if (handleValueChange) {
            handleValueChange(selectedChild.key, { newObj }); // 부모에서 받은 handleValueChange 호출
          }
        }}
        placeholder="Enter news content"
      />
    </div>
  );
}
