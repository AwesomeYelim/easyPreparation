import { selectedDetailState, worshipOrderState } from "@/recoilState";
import { useRecoilState, useRecoilValue, useSetRecoilState } from "recoil";
import { WorshipOrderItem } from "./WorshipOrder";
import BibleSelect from "./BibleSelect";

export default function Detail() {
  const selectedDetail = useRecoilValue(selectedDetailState);
  const setWorshipOrder = useSetRecoilState(worshipOrderState);

  const handleValueChange = (key: string, newObj: string) => {
    const updateData = (
      items: WorshipOrderItem[],
      keyParts: string[]
    ): WorshipOrderItem[] => {
      const [currentIndex, ...restKeyParts] = keyParts;
      if (!currentIndex) return items;
      return items.map((item, index) => {
        if (index === parseInt(currentIndex)) {
          if (restKeyParts.length === 0) {
            switch (item.info) {
              case "b_edit":
                return { ...item, obj: newObj };
              case "c_edit":
                return { ...item, obj: newObj };
              case "r_edit":
                return { ...item, lead: newObj };
            }
          }
          if (item.children) {
            return {
              ...item,
              children: updateData(item.children, restKeyParts),
            };
          }
        }
        return item;
      });
    };

    setWorshipOrder((prevData) => updateData(prevData, key.split("-")));
  };

  return (
    <section className="card">
      <h2>{selectedDetail?.title?.split("_")[1]}</h2>
      <div key={selectedDetail?.title} className="detail-card">
        <p>
          <strong>Obj</strong>
          {(selectedDetail?.info.includes("edit") &&
            selectedDetail.info.includes("b_") && (
              <BibleSelect
                handleValueChange={handleValueChange}
                parentKey={selectedDetail?.title || ""}
              />
            )) ||
            selectedDetail?.title}
        </p>
        <p>
          <strong>Info</strong> {selectedDetail?.info}
        </p>
        <p>
          <strong>Lead</strong> {selectedDetail?.lead}
        </p>
      </div>
    </section>
  );
}
