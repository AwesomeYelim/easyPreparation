"use client";

import { useState } from "react";

const newsCategories = [
  "교회 절기 및 행사",
  "예배 참여 안내",
  "선교회 소식",
  "교회 봉사",
  "노회 소식",
  "일부 교회 행사",
  "단임 목사 활동",
];

export default function ChurchNews() {
  const [selectedNews, setSelectedNews] = useState<string[]>([
    "교회 절기 및 행사",
    "예배 참여 안내",
  ]);

  return (
    <section className="card">
      <h2>교회 소식</h2>
      <div>
        {newsCategories.map((news) => (
          <span key={news} className="news-box">
            {news}
          </span>
        ))}
      </div>
    </section>
  );
}
