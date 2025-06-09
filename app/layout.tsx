import RecoilProvider from "@/components/RecoilProvider";
import NavBar from "@/components/NavBar";
import AuthProvider from "./lib/next-auth";
import { WebSocketProvider } from "./components/WebSocketProvider";
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
          <RecoilProvider>
            <AuthProvider>
              <WebSocketProvider>
                <NavBar />
                {children}
              </WebSocketProvider>
            </AuthProvider>
          </RecoilProvider>
        </div>
      </body>
    </html>
  );
}
