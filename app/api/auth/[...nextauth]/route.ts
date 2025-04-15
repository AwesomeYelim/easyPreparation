import NextAuth from "next-auth";
import CredentialsProvider from "next-auth/providers/credentials";

export const handler = NextAuth({
  providers: [
    CredentialsProvider({
      name: "Credentials",
      credentials: {
        username: { label: "Username", type: "text" },
        password: { label: "Password", type: "password" },
      },
      async authorize(credentials) {
        const user = { username: "user1", password: "password123" };

        if (
          credentials?.username === user.username &&
          credentials?.password === user.password
        ) {
          return user; // 로그인 성공 시 사용자 객체 반환
        } else {
          return null; // 로그인 실패 시 null 반환
        }
      },
    }),
  ],
  session: {
    strategy: "jwt", // JWT 기반 세션 관리
  },
  pages: {
    signIn: "/login", // 로그인 페이지 경로 설정
  },
});

export { handler as GET, handler as POST }; // GET, POST 메서드 처리
