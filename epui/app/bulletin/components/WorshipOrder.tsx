import { useRecoilState } from "recoil";
import { worshipOrderState } from "@/recoilState";

export type WorshipOrderItem = {
  title: string;
  obj: string;
  info: string;
  lead?: string;
  children?: WorshipOrderItem[];
};
export function WorshipOrder({
  selectedItems,
  setSelectedItems,
}: {
  selectedItems: WorshipOrderItem[];
  setSelectedItems: React.Dispatch<React.SetStateAction<WorshipOrderItem[]>>;
}) {
  const [worshipOrder, setWorshipOrder] = useRecoilState(worshipOrderState);

  const handleSelectItem = (item: WorshipOrderItem) => {
    // setWorshipOrder(worshipOrder.filter((i) => i.title !== item.title));
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
              {item.title.split("_")[1]}
            </span>
          ))}
      </div>
    </div>
  );
}
