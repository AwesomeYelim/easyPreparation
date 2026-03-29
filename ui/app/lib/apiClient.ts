import { WorshipOrderItem, UserSettings } from "@/types";

const BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL;

type FigmaInfo = { key: string; token: string };
type SongItem = { title: string; lyrics: string };

/** Display 창 참조 — 이미 열려있으면 reload 방지 */
let displayWindow: Window | null = null;

export function openDisplayWindow() {
  if (displayWindow && !displayWindow.closed) {
    displayWindow.focus();
    return;
  }
  displayWindow = window.open(`${BASE_URL}/display`, "display_window");
}

export const apiClient = {
  saveBulletin: (target: string, targetInfo: WorshipOrderItem[]) =>
    fetch(`${BASE_URL}/api/saveBulletin`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ target, targetInfo }),
    }),

  submitBulletin: (payload: { mark: string; targetInfo: WorshipOrderItem[]; target: string; figmaInfo: FigmaInfo }) =>
    fetch(`${BASE_URL}/submit`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    }),

  searchLyrics: (songs: SongItem[]) =>
    fetch(`${BASE_URL}/searchLyrics`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ songs }),
    }),

  submitLyrics: (payload: { figmaInfo: FigmaInfo; mark: string; songs: SongItem[] }) =>
    fetch(`${BASE_URL}/submitLyrics`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    }),

  startDisplay: (order: WorshipOrderItem[]) =>
    fetch(`${BASE_URL}/display/order`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(order),
    }),

  navigateDisplay: (direction: "next" | "prev") =>
    fetch(`${BASE_URL}/display/navigate`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ direction }),
    }),

  jumpDisplay: (index: number, subPageIdx?: number) =>
    fetch(`${BASE_URL}/display/jump`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ index, subPageIdx: subPageIdx ?? 0 }),
    }),

  getDisplayStatus: () =>
    fetch(`${BASE_URL}/display/status`).then((res) => {
      if (!res.ok) throw new Error(`display/status ${res.status}`);
      return res.json();
    }),

  timerControl: (action: string, factor?: number) =>
    fetch(`${BASE_URL}/display/timer`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ action, factor }),
    }),

  sendLyricsToDisplay: (songs: { title: string; lyrics: string; bpm: number }[]) =>
    fetch(`${BASE_URL}/display/lyrics-order`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ songs }),
    }),

  appendToDisplay: (items: any[]) =>
    fetch(`${BASE_URL}/display/append`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ items }),
    }),

  removeFromDisplay: (index: number) =>
    fetch(`${BASE_URL}/display/remove`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ index }),
    }),

  downloadFile: (fileName: string) => {
    const link = document.createElement("a");
    link.href = `${BASE_URL}/download?target=${fileName}`;
    link.download = fileName;
    link.target = "_self";
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  },

  // 찬송가 API
  getHymns: (page = 1, limit = 50, book?: string) =>
    fetch(`${BASE_URL}/api/hymns?page=${page}&limit=${limit}${book ? `&book=${book}` : ""}`)
      .then((r) => r.json()),

  searchHymns: (q: string, type?: string) =>
    fetch(`${BASE_URL}/api/hymns/search?q=${encodeURIComponent(q)}${type ? `&type=${type}` : ""}`)
      .then((r) => r.json()),

  getHymnDetail: (number: number, book = "new") =>
    fetch(`${BASE_URL}/api/hymns/detail?number=${number}&book=${book}`)
      .then((r) => r.json()),

  // 설정 API
  getSettings: (email: string) =>
    fetch(`${BASE_URL}/api/settings?email=${encodeURIComponent(email)}`)
      .then((r) => r.json()),

  saveSettings: (email: string, settings: Partial<UserSettings>) =>
    fetch(`${BASE_URL}/api/settings`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, ...settings }),
    }).then((r) => r.json()),

  saveLicense: (email: string, licenseKey: string, token: string) =>
    fetch(`${BASE_URL}/api/settings/license`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, license_key: licenseKey, token }),
    }).then((r) => r.json()),

  // 이력 API
  getHistory: (email: string, type?: string, page = 1) =>
    fetch(`${BASE_URL}/api/history?email=${encodeURIComponent(email)}${type ? `&type=${type}` : ""}&page=${page}`)
      .then((r) => r.json()),
};
