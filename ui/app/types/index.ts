export type WorshipOrderItem = {
  key: string;
  title: string;
  obj: string;
  info: string;
  lead?: string;
  contents?: string;
  children?: WorshipOrderItem[];
  bpm?: number;
  pages?: string[];
  lyricsMap?: string[];
  sections?: { label: string; startPage: number; text: string }[];
  source?: string;
  versionId?: number;
  ptzPreset?: number; // PTZ 카메라 프리셋 (0=없음, 1~9)
};

export type PTZConfig = {
  ip: string;
  port: number;
  username: string;
  password: string;
  profileToken: string;
  streamPath: string;
  presetCount: number;
  presetNames?: Record<string, string>; // {"1":"강대상","2":"피아노"}
  enabled: boolean;
};

export type PTZPreset = {
  token: string; // "1", "2", ...
  name: string;  // 카메라에 저장된 이름 (e.g. "강대상", "피아노")
};

export type OBSStatus = { connected: boolean; currentScene: string };

export type SongBlock = {
  title: string;
  lyrics: string;
  bpm: number;
  expanded: boolean;
};

export type UserChurchInfo = {
  id: number;
  name: string;
  english_name: string;
  title: string;
  content: string;
  email: string;
};

export type Hymn = {
  id: number;
  hymnbook: string;
  number: number;
  title: string;
  first_line?: string;
  category?: string;
  lyrics?: string;
  has_pdf: boolean;
};

export type UserSettings = {
  preferred_bible_version: number;
  theme: string;
  font_size: number;
  default_bpm: number;
  display_layout: string;
};

export type GenerationHistory = {
  id: number;
  type: string;
  filename?: string;
  status: string;
  metadata?: Record<string, any>;
  order_data?: WorshipOrderItem[] | { title: string; lyrics: string; bpm?: number }[];
  created_at: string;
};

export type ScheduleEntry = {
  worshipType: string;
  label: string;
  weekday: number;
  hour: number;
  minute: number;
  enabled: boolean;
};

export type ScheduleConfig = {
  entries: ScheduleEntry[];
  autoStream: boolean;
  countdownMinutes: number;
};

export type StreamStatus = {
  active: boolean;
  reconnecting: boolean;
  timecode: string;
  bytesSent: number;
};

export type DefaultTheme = {
  background: string;
  titleFormat: string;
};

export type SpecialDate = {
  date: string;
  label: string;
  background: string;
  titleOverride: string;
};

export type ThumbnailConfig = {
  defaults: Record<string, DefaultTheme>;
  specials: SpecialDate[];
  fontName?: string;
  logoPosition?: string;     // "bottom-right" | "bottom-left" | "top-right" | "top-left"
  logoSizePercent?: number;  // 0~30 (0=없음)
};

export type YouTubeStatus = {
  connected: boolean;
};

export type OBSSourceItem = {
  sceneItemId: number;
  sourceName: string;
  inputKind: string;
  enabled: boolean;
  positionX: number;
  positionY: number;
  scaleX: number;
  scaleY: number;
};

export type OBSDevice = {
  name: string;
  value: string;
};

export type LicensePlan = 'free' | 'pro' | 'enterprise';

export type LicenseFeature =
  | 'obs_control'
  | 'auto_scheduler'
  | 'youtube_integration'
  | 'thumbnail'
  | 'multi_worship'
  | 'cloud_backup';

export interface LicenseStatus {
  plan: LicensePlan;
  features: LicenseFeature[];
  expires_at: string | null;
  days_remaining: number;
  device_id: string;
  grace_period: boolean;
  is_active: boolean;
  dev_mode?: boolean;
}

export type OBSInitialSetupResult = {
  success: boolean;
  scenes_created: string[];
  sources_created: string[];
  warnings: string[];
};

declare global {
  interface Window {
    __resetTour?: () => void;
  }
}
