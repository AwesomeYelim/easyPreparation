import type { Config } from "tailwindcss";

// Direction A "studio dark" tokens — mirrors landing/design-mock/*.dc.html
const config: Config = {
  content: ["./app/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        bg: "#0b1220",
        band: "#0d1526",
        surface: "#0f172a",
        raised: "#111b2e",
        line: "rgba(255,255,255,0.08)",
        "line-strong": "rgba(255,255,255,0.16)",
        ink: "#e6edf7",
        muted: "#a7b4c8",
        dim: "#7c8ba3",
        faint: "#55627a",
        list: "#c5cfdd",
        accent: { DEFAULT: "#7fb3e6", light: "#a9cdf0", deep: "#4f8cc9" },
      },
      fontFamily: {
        sans: [
          "var(--font-noto)",
          "Apple SD Gothic Neo",
          "Malgun Gothic",
          "맑은 고딕",
          "system-ui",
          "sans-serif",
        ],
      },
      borderRadius: {
        btn: "6px",
        card: "10px",
      },
      maxWidth: {
        site: "1152px",
      },
    },
  },
  plugins: [],
};

export default config;
