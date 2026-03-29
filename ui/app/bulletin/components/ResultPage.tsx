import { WorshipOrderItem } from "@/types";
import { formatBibleReference } from "@/lib/bibleUtils";
import s from "../bulletin.module.scss";

export function ResultPart({
  selectedItems,
}: {
  selectedItems: WorshipOrderItem[];
}) {
  return (
    <div className={s.card}>
      <h2>생성된 예배 내용</h2>
      <div className={s.contents}>
        {(() => {
          const result = [];

          for (const el of selectedItems) {
            if (el.title !== "말씀내용" && el.title !== "행사") {
              result.push(
                <div className={s.row} key={el.title + el.obj}>
                  <div className={s.title}>{el.title}</div>
                  <div
                    className={s.obj}
                    title={
                      el.info === "b_edit"
                        ? formatBibleReference(el.obj)
                        : el.obj
                    }
                  >
                    {el.info === "b_edit"
                      ? formatBibleReference(el.obj)
                      : el.obj}
                  </div>
                  <div className={s.lead}>{el.lead}</div>
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
