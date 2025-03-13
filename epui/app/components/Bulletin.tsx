"use client";

import WorshipOrder from "@/components/WorshipOrder";
import SelectedOrder from "@/components/SelectedOrder";
import ChurchNews from "@/components/ChurchNews";

export default function Bulletin() {
  return (
    <div>
      <div className="title">
        <h1>Bulletin</h1>
        <h1>Lyrics</h1>
      </div>
      <WorshipOrder />
      <SelectedOrder />
      <ChurchNews />
    </div>
  );
}
