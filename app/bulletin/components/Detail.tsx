import React from "react";
import { useRecoilState } from "recoil";
import { selectedDetailState } from "@/recoilState";
import { WorshipOrderItem } from "../page";
import BibleSelect from "./BibleSelect";
import ChurchNews from "./ChurchNews";

export default function Detail({
  setSelectedItems,
}: {
  setSelectedItems: React.Dispatch<React.SetStateAction<WorshipOrderItem[]>>;
}) {
  const [selectedDetail, setSelectedDetail] =
    useRecoilState(selectedDetailState);

  const handleValueChange = (
    key: string,
    { newObj, newLead }: { newObj: string; newLead?: string }
  ) => {
    const updateData = (items: WorshipOrderItem[]): WorshipOrderItem[] => {
      return items.map((item) => {
        let updatedItem: WorshipOrderItem = { ...item };

        if (item.children) {
          const updatedChildren = updateData(item.children);
          updatedItem = {
            ...updatedItem,
            children: updatedChildren,
          };
        }

        if (item.key === key) {
          if (["b_edit", "c_edit", "edit"].includes(item.info)) {
            if (newObj) {
              updatedItem.obj = newObj;
            }
            if (newLead) {
              updatedItem.lead = newLead;
            }
          } else if (item.info === "r_edit" && newLead) {
            updatedItem.lead = newLead;
          }
        }

        // selectedDetail과 같은 key면, 전체를 업데이트
        if (item.key === selectedDetail.key) {
          setSelectedDetail(updatedItem);
        }

        return updatedItem;
      });
    };

    setSelectedItems((prevData) => updateData(prevData));
  };

  return (
    <section className="card">
      <h2>{selectedDetail?.title}</h2>
      {selectedDetail?.info.includes("edit") && (
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
              <textarea
                value={selectedDetail.obj}
                onChange={(e) =>
                  handleValueChange(selectedDetail.key, {
                    newObj: e.target.value,
                  })
                }
                placeholder={selectedDetail?.title}
              />
            )}
          </p>
          <p>
            <strong>
              Lead<span>right</span>
            </strong>
            <input
              type="text"
              value={selectedDetail.lead}
              onChange={(e) =>
                handleValueChange(selectedDetail.key, {
                  newObj: selectedDetail.obj,
                  newLead: e.target.value,
                })
              }
              placeholder={selectedDetail?.lead || "새로 입력하세요"}
            />
          </p>
        </div>
      )}
      {selectedDetail?.info.includes("notice") && (
        <ChurchNews
          handleValueChange={handleValueChange}
          selectedDetail={selectedDetail}
          setSelectedDetail={setSelectedDetail}
          setSelectedItems={setSelectedItems}
        />
      )}
      {!selectedDetail?.info.includes("edit") &&
        !selectedDetail?.info.includes("notice") && <>is not editable</>}
    </section>
  );
}
