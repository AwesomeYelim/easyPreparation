import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./app/**/*.{js,ts,jsx,tsx}",
    "./components/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: ["class", "[data-theme='dark']"],
  theme: {
    extend: {
      colors: {
        surface: "#F8FAFC",
        "surface-low": "#F1F5F9",
        "surface-lowest": "#FFFFFF",
        "surface-container": "#e7eefe",
        "surface-high": "#e2e8f8",
        "surface-highest": "#dce2f3",
        "navy-dark": "#020617",
        primary: "#0F172A",
        "electric-blue": "#3B82F6",
        secondary: "#2563EB",
        "on-surface": "#1E293B",
        "on-surface-variant": "#64748B",
        outline: "#CBD5E1",
        error: "#EF4444",
        "accent-cyan": "#06B6D4",
        "accent-lime": "#84CC16",
        "inverse-surface": "#0F172A",
      },
      fontFamily: {
        sans: ["Inter", "ui-sans-serif", "system-ui", "sans-serif"],
        headline: ["Inter", "ui-sans-serif", "system-ui", "sans-serif"],
        body: ["Inter", "ui-sans-serif", "system-ui", "sans-serif"],
      },
      borderRadius: {
        DEFAULT: "0.375rem",
        lg: "0.75rem",
        xl: "1rem",
        "2xl": "1.5rem",
        "3xl": "2rem",
        full: "9999px",
      },
    },
  },
  plugins: [],
};

export default config;
