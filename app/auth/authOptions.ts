import NextAuth from "next-auth";
import CredentialsProvider from "next-auth/providers/credentials";

export const authOptions = {
  providers: [
    CredentialsProvider({
      name: "Credentials",
      credentials: {
        username: { label: "Username", type: "text" },
        password: { label: "Password", type: "password" },
      },
      async authorize(credentials) {
        // 로그인 정보 처리
        if (
          credentials?.username === "admin" &&
          credentials?.password === "admin"
        ) {
          return { id: 1, name: "Admin", department: "IT" }; // 로그인 성공시 반환되는 정보
        } else {
          // 잘못된 로그인 정보 처리
          return null;
        }
      },
    }),
  ],

  secret: process.env.NEXTAUTH_SECRET,
};

export default NextAuth(authOptions);
