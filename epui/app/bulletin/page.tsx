"use client";

import WorshipOrder from "./components/WorshipOrder";
import SelectedOrder from "./components/SelectedOrder";
import ChurchNews from "./components/ChurchNews";

export default function Bulletin() {
  return (
    <div>
      <WorshipOrder />
      <SelectedOrder />
      <ChurchNews />
    </div>
  );
}
