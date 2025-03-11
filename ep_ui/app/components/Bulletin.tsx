import WorshipOrder from '@/components/WorshipOrder'
import SelectedOrder from '@/components/SelectedOrder'
import ChurchNews from '@/components/ChurchNews'

export default function Bulletin() {
  return (
    <main className="p-4">
      <h1 className="text-2xl font-bold text-center">Bulletin Lyrics</h1>
      <WorshipOrder />
      <SelectedOrder />
      <ChurchNews />
    </main>
  )
}
