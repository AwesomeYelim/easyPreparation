"use client";

import { useState } from "react";

export default function SelectedOrder() {
  const [selectedItems, setSelectedItems] = useState<string[]>([
    "전주",
    "예배의 부름",
    "개회기도",
    "찬송",
    "성시교독",
  ]);

  return (
    <section className="card">
      <h2>선택된 예배 순서</h2>
      <div>
        {selectedItems.map((item) => (
          <span key={item} className="tag">
            {item}
          </span>
        ))}
      </div>
    </section>
  );
}
