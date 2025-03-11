'use client'

import { useState } from 'react'

const newsCategories = [
  '교회 절기 및 행사',
  '예배 참여 안내',
  '선교회 소식',
  '교회 봉사',
  '노회 소식',
  '일부 교회 행사',
  '단임 목사 활동',
]

export default function ChurchNews() {
  const [selectedNews, setSelectedNews] = useState<string[]>([
    '교회 절기 및 행사',
    '예배 참여 안내',
  ])

  return (
    <section className="border p-4 rounded-md mt-4 bg-blue-900 text-white">
      <h2 className="text-lg font-semibold">교회 소식</h2>
      <div className="flex flex-wrap gap-2">
        {newsCategories.map((news) => (
          <span key={news} className="px-3 py-1 bg-white text-black rounded">
            {news}
          </span>
        ))}
      </div>
    </section>
  )
}
