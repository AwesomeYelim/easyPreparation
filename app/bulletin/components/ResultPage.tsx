import { WorshipOrderItem } from "../page";

export function ResultPart({
  selectedItems,
  setSelectedItems,
}: {
  selectedItems: WorshipOrderItem[];
  setSelectedItems: React.Dispatch<React.SetStateAction<WorshipOrderItem[]>>;
}) {
  console.log(selectedItems);

  return (
    <div className="card">
      <h2>생성된 예배 내용</h2>
      <div className="contents">
        {selectedItems.map((el) => {
          return (
            <div className="row">
              <div className="title">{el.title}</div>
              <div className="obj">{el.obj}</div>
              <div className="lead">{el.lead}</div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
