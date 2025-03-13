import "@/styles.scss";
import RecoilProvider from "@/components/RecoilProvider";

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="ko">
      <body>
        <RecoilProvider>{children}</RecoilProvider>
      </body>
    </html>
  );
}
