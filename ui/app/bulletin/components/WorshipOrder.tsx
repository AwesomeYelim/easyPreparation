import { WorshipOrderItem } from "@/types";
import fixData from "@/data/fix_data.json";

export function WorshipOrder({
  selectedItems,
  setSelectedItems,
}: {
  selectedItems: WorshipOrderItem[];
  setSelectedItems: React.Dispatch<React.SetStateAction<WorshipOrderItem[]>>;
}) {
  const handleSelectItem = (item: Partial<WorshipOrderItem>) => {
    setSelectedItems((prevItems) => [
      ...prevItems,
      {
        ...(item as WorshipOrderItem),
        key: `add_${Date.now()}_${prevItems.length}`,
        obj: (item as WorshipOrderItem).obj || "",
        lead: "",
      },
    ]);
  };

  return (
    <div className="bg-white rounded-2xl border border-slate-100 shadow-sm p-6">
      <h2 className="text-xs font-black uppercase tracking-[0.2em] text-on-surface-variant mb-4">
        예배 순서 선택하기
      </h2>
      <div className="flex flex-wrap gap-2">
        {fixData.map((item: Partial<WorshipOrderItem>) => (
          <span
            key={item.title}
            className="inline-flex items-center px-3 py-1.5 rounded-xl text-sm font-semibold cursor-pointer select-none border border-slate-200 bg-surface-low text-navy-dark hover:border-electric-blue hover:bg-electric-blue/5 hover:text-electric-blue transition-all"
            onClick={() => handleSelectItem(item)}
          >
            {item.title}
          </span>
        ))}
      </div>
    </div>
  );
}
