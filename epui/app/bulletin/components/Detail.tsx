import { selectedDetailState, churchNewsState } from "@/recoilState";
import { useRecoilValue } from "recoil";
import { useState } from "react";
import BibleSelect from "./BibleSelect";
import { WorshipOrderItem } from "../page";

export default function Detail({
  setSelectedItems,
}: {
  setSelectedItems: React.Dispatch<React.SetStateAction<WorshipOrderItem[]>>;
}) {
  const selectedDetail = useRecoilValue(selectedDetailState);
  const churchNews = useRecoilValue(churchNewsState);

  const [expandedKeys, setExpandedKeys] = useState<string[]>([]);

  const toggleExpand = (key: string) => {
    setExpandedKeys((prev) =>
      prev.includes(key) ? prev.filter((k) => k !== key) : [...prev, key]
    );
  };

  // 트리 노드 수정 핸들러
  const handleEdit = (key: string, newValue: string) => {
    setSelectedItems((prev) =>
      prev.map((item) => (item.key === key ? { ...item, obj: newValue } : item))
    );
  };

  // 트리 노드 삭제 핸들러
  const handleDelete = (key: string) => {
    const deleteItem = (items: WorshipOrderItem[]): WorshipOrderItem[] => {
      return items
        .filter((item) => item.key !== key) // 현재 요소 삭제
        .map((item) => ({
          ...item,
          children: item.children ? deleteItem(item.children) : [], // 자식 요소 재귀적으로 삭제
        }));
    };

    setSelectedItems((prev) => deleteItem(prev));
  };

  const renderTree = (items: WorshipOrderItem[]) => {
    return (
      <ul className="tree-list">
        {items.map((item) => {
          const hasChildren = item.children && item.children.length > 0;
          const isExpanded = expandedKeys.includes(item.key);

          return (
            <li key={item.key} className="tree-item">
              {hasChildren && (
                <button
                  onClick={() => toggleExpand(item.key)}
                  className="toggle-btn"
                >
                  {isExpanded ? "▼" : "▶️"}
                </button>
              )}
              <strong>{item.title}</strong> -{/* 🔽 편집 가능한 input 추가 */}
              <input
                type="text"
                value={item.obj}
                onChange={(e) => handleEdit(item.key, e.target.value)}
                className="edit-input"
              />
              {/* 삭제 버튼 추가 */}
              <button
                onClick={() => handleDelete(item.key)}
                className="delete-btn"
              >
                ❌
              </button>
              {hasChildren && isExpanded && (
                <div className="children">{renderTree(item.children)}</div>
              )}
            </li>
          );
        })}
      </ul>
    );
  };

  return (
    <section className="card">
      <h2>{selectedDetail?.title}</h2>

      {/* 편집 가능한 경우 */}
      {selectedDetail?.info.includes("edit") && (
        <div key={selectedDetail?.key} className="detail-card">
          <p>
            <strong>
              Object<span>center</span>
            </strong>
            {selectedDetail.info.includes("b_") ? (
              <BibleSelect
                handleValueChange={(key, newObj) =>
                  setSelectedItems((prev) =>
                    prev.map((item) =>
                      item.key === key ? { ...item, obj: newObj } : item
                    )
                  )
                }
                parentKey={selectedDetail?.key || ""}
              />
            ) : (
              <input
                type="text"
                onChange={(e) =>
                  setSelectedItems((prev) =>
                    prev.map((item) =>
                      item.key === selectedDetail.key
                        ? { ...item, obj: e.target.value }
                        : item
                    )
                  )
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
              onChange={(e) =>
                setSelectedItems((prev) =>
                  prev.map((item) =>
                    item.key === selectedDetail.key
                      ? { ...item, lead: e.target.value }
                      : item
                  )
                )
              }
              placeholder={selectedDetail?.lead || "새로 입력하세요"}
            />
          </p>
        </div>
      )}

      {/* 교회 소식 (트리 구조 편집 가능 + 삭제 기능) */}
      {selectedDetail?.info.includes("notice") && (
        <div className="church-news">{renderTree(churchNews.children)}</div>
      )}

      {!selectedDetail?.info.includes("edit") &&
        !selectedDetail?.info.includes("notice") && <>is not editable</>}
    </section>
  );
}
