import { WorshipOrderItem, UserSettings, ScheduleConfig, ThumbnailConfig, LicenseStatus, OBSSourceItem, OBSDevice } from "@/types";

const BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL
  || (typeof window !== 'undefined' ? window.location.origin : 'http://localhost:8080');

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
  // 예배 순서 API (Go 서버 마스터)
  getWorshipOrder: (type: string) =>
    fetch(`${BASE_URL}/api/worship-order?type=${type}`)
      .then((r) => r.json()) as Promise<WorshipOrderItem[]>,

  saveWorshipOrder: (type: string, items: WorshipOrderItem[]) =>
    fetch(`${BASE_URL}/api/worship-order`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ type, items }),
    }),

  submitBulletin: (payload: { mark: string; targetInfo: WorshipOrderItem[]; target: string; email?: string }) =>
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

  submitLyrics: (payload: { mark: string; songs: SongItem[]; email?: string }) =>
    fetch(`${BASE_URL}/submitLyrics`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    }),

  startDisplay: (order: WorshipOrderItem[], churchName?: string, email?: string, preprocessed?: boolean) =>
    fetch(`${BASE_URL}/display/order`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ items: order, churchName: churchName || "", email: email || "", ...(preprocessed ? { preprocessed: true } : {}) }),
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

  appendToDisplay: (items: any[], source?: string, afterIdx?: number) =>
    fetch(`${BASE_URL}/display/append`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        items,
        ...(source ? { source } : {}),
        ...(afterIdx !== undefined ? { afterIdx } : {}),
      }),
    }),

  removeFromDisplay: (index: number) =>
    fetch(`${BASE_URL}/display/remove`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ index }),
    }),

  reorderDisplay: (from: number, to: number) =>
    fetch(`${BASE_URL}/display/reorder`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ from, to }),
    }),

  downloadFile: (fileName: string) => {
    const url = `${BASE_URL}/download?target=${fileName}`;
    // Wails WebView2 환경: blob 다운로드가 안 되므로 시스템 브라우저로 열기
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const wails = (window as any)?.go?.main?.App;
    if (wails?.OpenURL) {
      wails.OpenURL(url);
      return;
    }
    // 일반 브라우저 환경: fetch+blob
    fetch(url)
      .then((r) => {
        if (!r.ok) throw new Error(`download failed: ${r.status}`);
        return r.blob();
      })
      .then((blob) => {
        const blobUrl = URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = blobUrl;
        a.download = `${fileName}.zip`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(blobUrl);
      })
      .catch((e) => console.error("downloadFile error:", e));
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

  // 스케줄러 API
  getSchedule: () =>
    fetch(`${BASE_URL}/api/schedule`).then((r) => r.json()),

  saveSchedule: (config: ScheduleConfig) =>
    fetch(`${BASE_URL}/api/schedule`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(config),
    }).then((r) => r.json()),

  streamControl: (action: "start" | "stop" | "status") =>
    fetch(`${BASE_URL}/api/schedule/stream`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ action }),
    }).then((r) => r.json()),

  scheduleTest: (action: "countdown" | "trigger", worshipType: string) =>
    fetch(`${BASE_URL}/api/schedule/test`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ action, worshipType }),
    }).then((r) => r.json()),

  // 썸네일 API
  getThumbnailConfig: () =>
    fetch(`${BASE_URL}/api/thumbnail/config`).then((r) => r.json()),

  saveThumbnailConfig: (config: ThumbnailConfig) =>
    fetch(`${BASE_URL}/api/thumbnail/config`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(config),
    }).then((r) => r.json()),

  generateThumbnail: (worshipType: string, date?: string, upload?: boolean) =>
    fetch(`${BASE_URL}/api/thumbnail/generate`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ worshipType, date, upload }),
    }).then((r) => r.json()),

  getThumbnailPreviewUrl: (worshipType: string, date?: string) =>
    `${BASE_URL}/api/thumbnail/preview?worshipType=${worshipType}${date ? `&date=${date}` : ""}`,

  uploadThumbnailBg: (file: File, target?: string) => {
    const fd = new FormData();
    fd.append("image", file);
    if (target) fd.append("target", target);
    return fetch(`${BASE_URL}/api/thumbnail/upload`, { method: "POST", body: fd })
      .then((r) => r.json());
  },

  getThumbnailImageUrl: (path: string) =>
    `${BASE_URL}/api/thumbnail/image?path=${encodeURIComponent(path)}`,

  // 버전 + 업데이트 API
  getVersion: () =>
    fetch(`${BASE_URL}/api/version`).then((r) => r.json()) as Promise<{
      version: string;
      commit: string;
      buildTime: string;
    }>,

  checkUpdate: () =>
    fetch(`${BASE_URL}/api/update/check`).then((r) => r.json()) as Promise<{
      ok: boolean;
      current: string;
      latest?: string;
      updateUrl?: string;
      notes?: string;
      hasUpdate?: boolean;
      error?: string;
    }>,

  startUpdateDownload: async () => {
    const res = await fetch(`${BASE_URL}/api/update/download`, { method: 'POST' });
    return res.json() as Promise<{ ok: boolean; version?: string; error?: string }>;
  },

  applyUpdate: async () => {
    const res = await fetch(`${BASE_URL}/api/update/apply`, { method: 'POST' });
    return res.json() as Promise<{ ok: boolean; restartRequired?: boolean; error?: string }>;
  },

  getUpdateStatus: async () => {
    const res = await fetch(`${BASE_URL}/api/update/status`);
    return res.json() as Promise<{
      state: 'idle' | 'checking' | 'downloading' | 'downloaded' | 'applying' | 'restart_required' | 'error';
      percent: number;
      totalBytes: number;
      downloadedBytes: number;
      version: string;
      error?: string;
    }>;
  },

  cancelUpdateDownload: async () => {
    const res = await fetch(`${BASE_URL}/api/update/cancel`, { method: 'POST' });
    return res.json() as Promise<{ ok: boolean }>;
  },

  // 라이선스 API
  getLicenseStatus: async (): Promise<LicenseStatus> => {
    const res = await fetch(`${BASE_URL}/api/license`);
    return res.json();
  },

  activateLicense: async (licenseKey: string) => {
    const res = await fetch(`${BASE_URL}/api/license/activate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ license_key: licenseKey }),
    });
    return res.json();
  },

  deactivateLicense: async () => {
    const res = await fetch(`${BASE_URL}/api/license/deactivate`, { method: 'POST' });
    return res.json();
  },

  verifyLicense: async () => {
    const res = await fetch(`${BASE_URL}/api/license/verify`, { method: 'POST' });
    return res.json();
  },

  // 결제 API
  createCheckoutSession: async (plan: 'pro_monthly' | 'pro_annual') => {
    const res = await fetch(`${BASE_URL}/api/license/checkout`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ plan }),
    });
    return res.json() as Promise<{ checkoutUrl: string; sessionId: string }>;
  },

  pollActivation: async (sessionId: string) => {
    const res = await fetch(`${BASE_URL}/api/license/callback`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ sessionId }),
    });
    return res.json() as Promise<{ status: 'pending' | 'completed'; plan?: string; licenseKey?: string }>;
  },

  getPortalUrl: async () => {
    const res = await fetch(`${BASE_URL}/api/license/portal`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
    });
    return res.json() as Promise<{ portalUrl: string }>;
  },

  // 배경 템플릿 API
  getTemplates: (category: string) =>
    fetch(`${BASE_URL}/api/templates?category=${category}`)
      .then((r) => r.json()) as Promise<{ files: { name: string; url: string; size: number }[] }>,

  uploadTemplate: (file: File, category: string, name?: string) => {
    const fd = new FormData();
    fd.append("image", file);
    fd.append("category", category);
    if (name) fd.append("name", name);
    return fetch(`${BASE_URL}/api/templates/upload`, { method: "POST", body: fd })
      .then((r) => r.json()) as Promise<{ ok: boolean; name: string; url: string }>;
  },

  deleteTemplate: (category: string, filename: string) =>
    fetch(`${BASE_URL}/api/templates/${category}/${encodeURIComponent(filename)}`, { method: "DELETE" })
      .then((r) => r.json()) as Promise<{ ok: boolean }>,

  getTemplateUrl: (category: string, filename: string) =>
    `${BASE_URL}/api/templates/${category}/${encodeURIComponent(filename)}`,

  // YouTube API
  getYoutubeStatus: () =>
    fetch(`${BASE_URL}/api/youtube/status`).then((r) => r.json()),

  getYoutubeAuthUrl: () => `${BASE_URL}/api/youtube/auth`,

  setupOBSStream: (worshipType?: string) =>
    fetch(`${BASE_URL}/api/youtube/setup-obs`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ worshipType: worshipType || "main_worship" }),
    }).then((r) => r.json()),

  // OBS 소스 관리 API
  getOBSScenes: () =>
    fetch(`${BASE_URL}/api/obs/scenes`).then((r) => r.json()) as Promise<{
      connected: boolean; scenes: string[]; currentScene?: string; error?: string;
    }>,

  getOBSSources: (scene: string) =>
    fetch(`${BASE_URL}/api/obs/sources?scene=${encodeURIComponent(scene)}`)
      .then((r) => r.json()) as Promise<{ items: OBSSourceItem[]; error?: string }>,

  uploadOBSLogo: (file: File) => {
    const fd = new FormData();
    fd.append("image", file);
    return fetch(`${BASE_URL}/api/obs/logo/upload`, { method: "POST", body: fd })
      .then((r) => r.json()) as Promise<{ ok: boolean; path?: string }>;
  },

  applyOBSLogo: (scene: string, position: string, scale: number, x?: number, y?: number) =>
    fetch(`${BASE_URL}/api/obs/logo/apply`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ scene, position, scale, x: x ?? 0, y: y ?? 0 }),
    }).then((r) => r.json()) as Promise<{ ok: boolean; sceneItemId?: number; error?: string }>,

  getOBSCameraDevices: () =>
    fetch(`${BASE_URL}/api/obs/camera/devices`).then((r) => r.json()) as Promise<{
      devices: OBSDevice[]; error?: string;
    }>,

  addOBSCamera: (scene: string, deviceId: string, inputName: string) =>
    fetch(`${BASE_URL}/api/obs/camera/add`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ scene, deviceId, inputName }),
    }).then((r) => r.json()) as Promise<{ ok: boolean; sceneItemId?: number; error?: string }>,

  toggleOBSSource: (scene: string, sceneItemId: number, enabled: boolean) =>
    fetch(`${BASE_URL}/api/obs/sources/toggle`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ scene, sceneItemId, enabled }),
    }).then((r) => r.json()) as Promise<{ ok: boolean; error?: string }>,

  removeOBSSource: (inputName: string) =>
    fetch(`${BASE_URL}/api/obs/sources/remove`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ inputName }),
    }).then((r) => r.json()) as Promise<{ ok: boolean; error?: string }>,

  setupOBSDisplay: (scene?: string, url?: string) =>
    fetch(`${BASE_URL}/api/obs/setup-display`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ scene: scene ?? "", url: url ?? "" }),
    }).then((r) => r.json()) as Promise<{ ok: boolean; sceneItemId?: number; inputName?: string; scene?: string; url?: string; error?: string }>,
};
