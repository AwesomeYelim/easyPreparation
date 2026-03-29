import { WorshipOrderItem } from "@/types";
import fixData from "@/data/fix_data.json";
import s from "../bulletin.module.scss";

export function WorshipOrder({
  selectedItems,
  setSelectedItems,
}: {
  selectedItems: WorshipOrderItem[];
  setSelectedItems: React.Dispatch<React.SetStateAction<WorshipOrderItem[]>>;
}) {
  const handleSelectItem = (item: Partial<WorshipOrderItem>) => {
    setSelectedItems((prevItems) => [
      ...prevItems,
      {
        ...(item as WorshipOrderItem),
        key: `add_${Date.now()}_${prevItems.length}`,
        obj: (item as WorshipOrderItem).obj || "",
        lead: "",
      },
    ]);
  };

  return (
    <div className={s.card}>
      <h2>예배 순서 선택하기</h2>
      <div>
        {fixData.map((item: Partial<WorshipOrderItem>) => (
          <span key={item.title} className={`${s.tag} ${s.fix}`} onClick={() => handleSelectItem(item)}>
            {item.title}
          </span>
        ))}
      </div>
    </div>
  );
}
