const base = {
  width: 26,
  height: 26,
  viewBox: "0 0 24 24",
  fill: "none",
  stroke: "#7fb3e6",
  strokeWidth: 1.7,
  strokeLinecap: "round" as const,
  strokeLinejoin: "round" as const,
  "aria-hidden": true,
};

export const IconDoc = () => (
  <svg {...base}>
    <path d="M6 3h9l4 4v14H6z" />
    <path d="M14 3v5h5" />
    <path d="M9 13h7M9 17h5" />
  </svg>
);

export const IconMusic = () => (
  <svg {...base}>
    <path d="M9 18V6l10-2v12" />
    <circle cx="6.5" cy="18" r="2.5" />
    <circle cx="16.5" cy="16" r="2.5" />
  </svg>
);

export const IconBook = () => (
  <svg {...base}>
    <path d="M4 5h6a2 2 0 0 1 2 2v13a2 2 0 0 0-2-2H4z" />
    <path d="M20 5h-6a2 2 0 0 0-2 2v13a2 2 0 0 1 2-2h6z" />
  </svg>
);

export const IconScreen = () => (
  <svg {...base}>
    <rect x="3" y="4" width="18" height="12" rx="2" />
    <path d="M8 20h8M12 16v4" />
  </svg>
);

export const IconLines = () => (
  <svg {...base}>
    <path d="M4 6h16M4 12h16M4 18h10" />
  </svg>
);

export const IconGear = () => (
  <svg {...base}>
    <circle cx="12" cy="12" r="3" />
    <path d="M12 3v3M12 18v3M3 12h3M18 12h3M5.6 5.6l2.1 2.1M16.3 16.3l2.1 2.1M5.6 18.4l2.1-2.1M16.3 7.7l2.1-2.1" />
  </svg>
);

const os = { ...base, width: 40, height: 40, strokeWidth: 1.6 };

export const IconApple = () => (
  <svg {...os}>
    <path d="M16 3c-1 2-3 2-4 1 1-2 3-2 4-1z" />
    <path d="M12 7c-4 0-6 3-6 7 0 3 2 7 4 7 1 0 2-1 3-1s2 1 3 1c2 0 4-4 4-7 0-2-1-4-3-5-2 0-3 1-5 1z" />
  </svg>
);

export const IconWindows = () => (
  <svg {...os}>
    <path d="M3 5.5l8-1.2v7.2H3zM12 4.2l9-1.2v8.5h-9zM3 12.5h8v7.2l-8-1.2zM12 12.5h9V21l-9-1.2z" />
  </svg>
);

export const IconLinux = () => (
  <svg {...os}>
    <rect x="4" y="4" width="16" height="16" rx="2" />
    <path d="M8 17l3-6 2 4 1-2 2 4" />
  </svg>
);
