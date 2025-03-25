import RecoilProvider from "@/components/recoil-provider";
import NavBar from "@/components/nav-bar";
import "@/globals.css";
import "@/styles.scss";

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ko">
      <body>
        <div className="entire-wrap">
          <NavBar />
          <RecoilProvider>{children}</RecoilProvider>
        </div>
      </body>
    </html>
  );
}
