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

// rekeyChildren — 배열 순서에 맞춰 key를 "{parentKey}.{index}"로 재부여 (하위 자손 key까지 재귀적으로 갱신)
// 순서 변경 후 key가 위치와 어긋나면 handleModifyChild의 "다음 key = 마지막 key + 1" 계산이
// 기존 항목과 충돌할 수 있어, 재정렬 시 항상 key를 위치에 맞게 다시 매겨야 함
export const rekeyChildren = (parentKey: string, children: WorshipOrderItem[]): WorshipOrderItem[] =>
  children.map((child, idx) => {
    const newKey = `${parentKey}.${idx}`;
    return child.children
      ? { ...child, key: newKey, children: rekeyChildren(newKey, child.children) }
      : { ...child, key: newKey };
  });

// replaceChildren — key가 일치하는 노드의 children을 통째로 교체 (트리 어디에 있든 재귀 탐색)
export const replaceChildren = (
  items: WorshipOrderItem[],
  parentKey: string,
  newChildren: WorshipOrderItem[]
): WorshipOrderItem[] =>
  items.map((item) => {
    if (item.key === parentKey) return { ...item, children: newChildren };
    if (item.children) return { ...item, children: replaceChildren(item.children, parentKey, newChildren) };
    return item;
  });

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
