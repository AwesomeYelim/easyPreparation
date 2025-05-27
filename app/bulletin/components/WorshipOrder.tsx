import { useRecoilValue } from "recoil";
import { worshipOrderState } from "@/recoilState";
import { WorshipOrderItem } from "../page";
import fixData from "@/data/fix_data.json";

export function WorshipOrder({
  selectedItems,
  setSelectedItems,
}: {
  selectedItems: WorshipOrderItem[];
  setSelectedItems: React.Dispatch<React.SetStateAction<WorshipOrderItem[]>>;
}) {
  // const worshipOrder = useRecoilValue(worshipOrderState);

  const handleSelectItem = (item: Partial<WorshipOrderItem>) => {
    setSelectedItems((prevItems) => [
      ...prevItems,
      {
        ...(item as WorshipOrderItem),
        key: String(selectedItems.length),
        lead: "",
      },
    ]);
  };

  return (
    <div className="card">
      <h2>예배 순서 선택하기</h2>
      <div>
        {fixData.map((item) => (
          <span
            key={item.title}
            className="fix tag"
            onClick={() => handleSelectItem(item)}
          >
            {item.title}
          </span>
        ))}
      </div>
    </div>
  );
}
