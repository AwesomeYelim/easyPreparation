"use client";

import { useState } from "react";

const worshipItems = [
  "전주",
  "예배의 부름",
  "개회기도",
  "찬송",
  "성시교독",
  "신앙고백",
  "기도",
  "성경봉독",
  "특송",
  "참회의 기도",
  "말씀",
  "헌금봉헌",
  "교회소식",
  "축도",
];

export default function WorshipOrder() {
  const [selectedItems, setSelectedItems] = useState<string[]>([]);

  const toggleItem = (item: string) => {
    setSelectedItems((prev) =>
      prev.includes(item) ? prev.filter((i) => i !== item) : [...prev, item]
    );
  };

  return (
    <section className="card">
      <h2>예배 순서 선택하기</h2>
      <div>
        {worshipItems.map((item) => (
          <button key={item} className="tag" onClick={() => toggleItem(item)}>
            {item}
          </button>
        ))}
      </div>
    </section>
  );
}
