import { useRecoilState } from "recoil";
import { worshipOrderState } from "../../recoilState";

export default function WorshipOrder() {
  const [worshipOrder, setWorshipOrder] = useRecoilState(worshipOrderState);

  return (
    <div className="card">
      <h2>예배 순서 선택하기</h2>
      <div>
        {worshipOrder.map((item) => (
          <span key={item.title} className="tag">
            {item.title.split("_")[1]}
          </span>
        ))}
      </div>
    </div>
  );
}
