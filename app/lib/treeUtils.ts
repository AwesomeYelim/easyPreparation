import { WorshipOrderItem } from "@/types";

export const deleteNode = (items: WorshipOrderItem[], key: string): WorshipOrderItem[] =>
  items
    .map((item) => {
      if (item.key === key) return null;
      return item.children ? { ...item, children: deleteNode(item.children, key) } : item;
    })
    .filter(Boolean) as WorshipOrderItem[];

export const insertSiblingNode = (items: WorshipOrderItem[], newItem: WorshipOrderItem): WorshipOrderItem[] => {
  const keys = newItem.key.split(".");
  const lastKey = parseInt(keys[keys.length - 1], 10);
  const beforeKey = `${keys.slice(0, -1).join(".")}.${lastKey - 1}`;

  return items.flatMap((item) => {
    if (item.key === beforeKey) {
      return [item, newItem];
    } else if (item.children) {
      return [{ ...item, children: insertSiblingNode(item.children, newItem) }];
    }
    return [item];
  });
};

export const findNode = (items: WorshipOrderItem[] | undefined, key: string): WorshipOrderItem | null => {
  if (!items) return null;
  for (const item of items) {
    if (item.key === key) return item;
    if (item.children) {
      const found = findNode(item.children, key);
      if (found) return found;
    }
  }
  return null;
};
