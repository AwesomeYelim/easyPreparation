import WorshipOrder from "@/components/WorshipOrder";
import SelectedOrder from "@/components/SelectedOrder";
import ChurchNews from "@/components/ChurchNews";

export default function Bulletin() {
  return (
    <div>
      <h1>Bulletin Lyrics</h1>
      <WorshipOrder />
      <SelectedOrder />
      <ChurchNews />
    </div>
  );
}
