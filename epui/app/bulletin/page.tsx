"use client";

import { WorshipOrder, WorshipOrderItem } from "./components/WorshipOrder";
import SelectedOrder from "./components/SelectedOrder";
import ChurchNews from "./components/ChurchNews";
import { useState } from "react";

export default function Bulletin() {
  const [selectedItems, setSelectedItems] = useState<WorshipOrderItem[]>([]);

  return (
    <div>
      <WorshipOrder setSelectedItems={setSelectedItems} />
      <SelectedOrder
        selectedItems={selectedItems}
        setSelectedItems={setSelectedItems}
      />
      <ChurchNews />
    </div>
  );
}
