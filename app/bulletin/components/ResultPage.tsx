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
        {(() => {
          const result = [];

          for (const el of selectedItems) {
            if (el.title !== "말씀내용") {
              result.push(
                <div className="row" key={el.title + el.obj}>
                  <div className="title">{el.title}</div>
                  <div className="obj">{el.obj}</div>
                  <div className="lead">{el.lead}</div>
                </div>
              );
            }

            if (el.title === "축도") break;
          }

          return result;
        })()}
      </div>
    </div>
  );
}
