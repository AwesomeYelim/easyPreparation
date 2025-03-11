'use client'

import { useState } from 'react'

export default function SelectedOrder() {
  const [selectedItems, setSelectedItems] = useState<string[]>([
    '전주',
    '예배의 부름',
    '개회기도',
    '찬송',
    '성시교독',
  ])

  return (
    <section className="border p-4 rounded-md mt-4">
      <h2 className="text-lg font-semibold">선택된 예배 순서</h2>
      <div className="flex flex-wrap gap-2">
        {selectedItems.map((item) => (
          <span key={item} className="px-3 py-1 bg-blue-500 text-white rounded">
            {item}
          </span>
        ))}
      </div>
    </section>
  )
}
