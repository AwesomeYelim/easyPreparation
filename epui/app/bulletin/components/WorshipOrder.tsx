import { useRecoilState } from "recoil";
import { worshipOrderState } from "@/recoilState";

export type WorshipOrderItem = {
  title: string;
  obj?: string;
  info?: string;
  lead?: string;
};

type WorshipOrderProps = {
  setSelectedItems: React.Dispatch<React.SetStateAction<WorshipOrderItem[]>>;
};

export function WorshipOrder({ setSelectedItems }: WorshipOrderProps) {
  const [worshipOrder, setWorshipOrder] = useRecoilState(worshipOrderState);

  const handleSelectItem = (item: WorshipOrderItem) => {
    setWorshipOrder(worshipOrder.filter((i) => i.title !== item.title));
    setSelectedItems((prevItems) => [...prevItems, item]);
  };

  return (
    <div className="card">
      <h2>예배 순서 선택하기</h2>
      <div>
        {worshipOrder.map((item) => (
          <span
            key={item.title}
            className="tag"
            onClick={() => handleSelectItem(item)} // Handle item click
          >
            {item.title.split("_")[1]}
          </span>
        ))}
      </div>
    </div>
  );
}
