import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./app/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        navy: { DEFAULT: "#1f3f62", light: "#4f8cc9" },
      },
      fontFamily: {
        sans: [
          "Apple SD Gothic Neo",
          "Malgun Gothic",
          "맑은 고딕",
          "sans-serif",
        ],
      },
    },
  },
  plugins: [],
};

export default config;
