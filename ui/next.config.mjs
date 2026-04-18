const isDev = process.env.NODE_ENV !== "production";

const config = {
  reactStrictMode: true,
  trailingSlash: true,
  typescript: {},
};

if (!isDev) {
  config.output = "export";
}

if (isDev) {
  config.rewrites = async () => [
    { source: "/api/:path*", destination: "http://localhost:8080/api/:path*" },
    { source: "/display/:path*", destination: "http://localhost:8080/display/:path*" },
    { source: "/ws", destination: "http://localhost:8080/ws" },
    { source: "/submit", destination: "http://localhost:8080/submit" },
    { source: "/download", destination: "http://localhost:8080/download" },
    { source: "/searchLyrics", destination: "http://localhost:8080/searchLyrics" },
    { source: "/submitLyrics", destination: "http://localhost:8080/submitLyrics" },
    { source: "/mobile/:path*", destination: "http://localhost:8080/mobile/:path*" },
  ];
}

export default config;
