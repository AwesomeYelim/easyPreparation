import { WorshipOrderItem } from "@/types";
import { formatBibleReference } from "@/lib/bibleUtils";

export function ResultPart({
  selectedItems,
}: {
  selectedItems: WorshipOrderItem[];
}) {
  return (
    <div className="bg-pro-surface rounded-lg border border-pro-border p-6 sticky top-0">
      <h2 className="text-xs font-black uppercase tracking-[0.2em] text-pro-text-muted mb-4">
        생성된 예배 내용
      </h2>
      <div className="flex flex-col gap-0">
        {(() => {
          const result = [];

          for (const el of selectedItems) {
            if (el.title !== "말씀내용" && el.title !== "행사") {
              result.push(
                <div
                  key={el.title + el.obj}
                  className="flex items-center justify-between py-2.5 border-b border-pro-border last:border-0 gap-2"
                >
                  <div className="text-sm font-bold text-pro-text min-w-0 flex-shrink-0 max-w-[100px] truncate">
                    {el.title}
                  </div>
                  <div
                    className="text-xs text-pro-text-muted flex-1 min-w-0 truncate text-center"
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
                  <div className="text-xs text-pro-text-muted flex-shrink-0 max-w-[80px] truncate text-right">
                    {el.lead}
                  </div>
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
