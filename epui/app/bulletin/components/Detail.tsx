import { selectedDetailState } from "@/recoilState";
import { useRecoilValue } from "recoil";
import BibleSelect from "./BibleSelect";
import { WorshipOrderItem } from "../page";

export default function Detail({
  setSelectedItems,
}: {
  setSelectedItems: React.Dispatch<React.SetStateAction<WorshipOrderItem[]>>;
}) {
  const selectedDetail = useRecoilValue(selectedDetailState);
  // const setWorshipOrder = useSetRecoilState(worshipOrderState);

  const handleValueChange = (key: number, newObj: string) => {
    const updateData = (items: WorshipOrderItem[]): WorshipOrderItem[] => {
      return items.map((item) => {
        if (item.key === key) {
          switch (item.info) {
            case "b_edit":
              return { ...item, obj: newObj };
            case "c_edit":
              return { ...item, obj: newObj };
            case "r_edit":
              return { ...item, lead: newObj };
            case "edit":
              return { ...item, obj: newObj };
          }
          if (item.children) {
            return {
              ...item,
              children: updateData(item.children),
            };
          }
        }
        return item;
      });
    };

    setSelectedItems((prevData) => updateData(prevData));
  };

  return (
    <section className="card">
      <h2>{selectedDetail?.title}</h2>{" "}
      {selectedDetail?.info.includes("edit") ? (
        <div key={selectedDetail?.key} className="detail-card">
          <p>
            <strong>
              Object<span>center</span>
            </strong>
            {(selectedDetail.info.includes("b_") && (
              <BibleSelect
                handleValueChange={handleValueChange}
                parentKey={selectedDetail?.key || ""}
              />
            )) || (
              <input
                type="text"
                onChange={(e) =>
                  handleValueChange(selectedDetail.key, e.target.value)
                }
                placeholder={selectedDetail?.title}
              />
            )}
          </p>
          {/* <p>
          <strong>
            Information <span>center</span>
          </strong>
          <input
            type="text"
            onChange={(e) => handleValueChange(key, e.target.value)}
            placeholder={selectedDetail?.info}
          />
        </p> */}
          <p>
            <strong>
              Lead<span>right</span>
            </strong>
            <input
              type="text"
              onChange={(e) =>
                handleValueChange(selectedDetail.key, e.target.value)
              }
              placeholder={selectedDetail?.lead || "새로 입력하세요"}
            />
          </p>
        </div>
      ) : (
        <>is not editable</>
      )}
    </section>
  );
}
