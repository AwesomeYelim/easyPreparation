import { selectedDetailState } from "@/recoilState";
import { useRecoilValue } from "recoil";
import { WorshipOrderItem } from "./WorshipOrder";

export default function Detail() {
  const selectedDetail: WorshipOrderItem | null = useRecoilValue(selectedDetailState);

  return (
    <section className="card">
      <h2>{selectedDetail?.title?.split("_")[1]}</h2>
      <div key={selectedDetail?.title} className="detail-card">
        <p>
          <strong>Obj:</strong> {selectedDetail?.obj}
        </p>
        <p>
          <strong>Info:</strong> {selectedDetail?.info}
        </p>
        <p>
          <strong>Lead:</strong> {selectedDetail?.lead}
        </p>
      </div>
    </section>
  );
}
