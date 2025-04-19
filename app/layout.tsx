import RecoilProvider from "@/components/recoil-provider";
import NavBar from "@/components/nav-bar";
import AuthProvider from "./lib/next-auth";
import "@/globals.css";
import "@/styles.scss";

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="ko">
      <body>
        <div className="entire-wrap">
          <AuthProvider>
            <NavBar />
            <RecoilProvider>{children}</RecoilProvider>
          </AuthProvider>
        </div>
      </body>
    </html>
  );
}
