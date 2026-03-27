import NextAuth from "next-auth";
import GithubProvider from "next-auth/providers/github";
import GoogleProvider from "next-auth/providers/google";
import pool from "@/lib/db";

const handler = NextAuth({
  providers: [
    GithubProvider({
      clientId: process.env.GITHUB_ID as string,
      clientSecret: process.env.GITHUB_SECRET as string,
    }),
    GoogleProvider({
      clientId: process.env.GOOGLE_CLIENT_ID as string,
      clientSecret: process.env.GOOGLE_CLIENT_SECRET as string,
    }),
  ],

  events: {
    async signIn({ user }) {
      if (!user?.email) return;

      try {
        const { rows } = await pool.query(
          "SELECT 1 FROM churches WHERE email = $1 LIMIT 1",
          [user.email]
        );

        if (rows.length === 0) {
          await pool.query(
            "INSERT INTO churches (name, english_name, email) VALUES ($1, $2, $3)",
            ["", "", user.email]
          );
          console.log(`Inserted new user email: ${user.email}`);
        }
      } catch (error) {
        console.error("Failed to insert into churches:", error);
      }
    },
  },
});

export { handler as GET, handler as POST };
