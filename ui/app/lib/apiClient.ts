import { WorshipOrderItem } from "@/types";

const BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL;

type FigmaInfo = { key: string; token: string };
type SongItem = { title: string; lyrics: string };

export const apiClient = {
  saveBulletin: (target: string, targetInfo: WorshipOrderItem[]) =>
    fetch("/api/saveBulletin", {
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

  downloadFile: (fileName: string) => {
    const link = document.createElement("a");
    link.href = `${BASE_URL}/download?target=${fileName}`;
    link.download = fileName;
    link.target = "_self";
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  },
};
