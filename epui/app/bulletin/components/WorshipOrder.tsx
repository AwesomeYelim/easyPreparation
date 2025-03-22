import { useRecoilValue } from "recoil";
import { worshipOrderState } from "@/recoilState";
import { WorshipOrderItem } from "../page";

export function WorshipOrder({
  selectedItems,
  setSelectedItems,
}: {
  selectedItems: WorshipOrderItem[];
  setSelectedItems: React.Dispatch<React.SetStateAction<WorshipOrderItem[]>>;
}) {
  const worshipOrder = useRecoilValue(worshipOrderState);

  const handleSelectItem = (item: WorshipOrderItem) => {
    setSelectedItems((prevItems) => [...prevItems, item]);
  };

  return (
    <div className="card">
      <h2>예배 순서 선택하기</h2>
      <div>
        {worshipOrder
          .filter(
            (el) => !selectedItems.map((el) => el.title).includes(el.title)
          )
          .map((item) => (
            <span
              key={item.title}
              className="tag"
              onClick={() => handleSelectItem(item)}
            >
              {item.title}
            </span>
          ))}
      </div>
    </div>
  );
}
