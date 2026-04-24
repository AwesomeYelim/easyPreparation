import React from "react";
import { useRecoilState } from "recoil";
import { selectedDetailState } from "@/recoilState";
import { WorshipOrderItem } from "@/types";
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
    let updatedDetail: WorshipOrderItem | null = null;

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

        if (item.key == key) {
          if (["b_edit", "c_edit", "c-edit", "edit"].includes(item.info)) {
            updatedItem.obj = newObj;
            if (newLead !== undefined) {
              updatedItem.lead = newLead;
            }
          } else if (item.info === "r_edit" && newLead !== undefined) {
            updatedItem.lead = newLead;
          }
        }

        if (item.key === selectedDetail.key) {
          updatedDetail = updatedItem;
        }
        return updatedItem;
      });
    };

    setSelectedItems((prevData) => updateData(prevData));
    if (updatedDetail) {
      setSelectedDetail(updatedDetail);
    }
  };

  return (
    <section className="bg-pro-surface rounded-lg border border-pro-border p-6">
      <h2 className="text-xs font-black uppercase tracking-[0.2em] text-pro-text-muted mb-4">
        {selectedDetail?.title || "상세 편집"}
      </h2>

      {selectedDetail?.info.includes("edit") && (
        <div key={selectedDetail?.key} className="flex flex-col gap-4">
          {/* Object 필드 */}
          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-bold text-pro-text uppercase tracking-wider flex items-center gap-1">
              Object
              <span className="text-electric-blue text-[10px] font-black">* center</span>
            </label>
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
                className="w-full px-3 py-2.5 border border-pro-border rounded-lg text-sm bg-pro-elevated text-pro-text resize-none min-h-[100px] focus:outline-none focus:border-electric-blue focus:ring-2 focus:ring-electric-blue/20 transition-all"
              />
            )}
          </div>

          {/* Lead 필드 */}
          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-bold text-pro-text uppercase tracking-wider flex items-center gap-1">
              Lead
              <span className="text-pro-text-muted text-[10px] font-semibold">right</span>
            </label>
            <input
              type="text"
              value={selectedDetail.lead}
              onChange={(e) =>
                handleValueChange(selectedDetail.key, {
                  newObj: selectedDetail.obj,
                  newLead: e.target.value,
                })
              }
              placeholder={selectedDetail?.lead ? "" : "새로 입력하세요"}
              className="w-full px-3 py-2.5 border border-pro-border rounded-lg text-sm bg-pro-elevated text-pro-text focus:outline-none focus:border-electric-blue focus:ring-2 focus:ring-electric-blue/20 transition-all"
            />
          </div>
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
        !selectedDetail?.info.includes("notice") && (
          <p className="text-pro-text-muted text-sm py-3">
            이 항목은 자동으로 처리됩니다
          </p>
        )}
    </section>
  );
}
