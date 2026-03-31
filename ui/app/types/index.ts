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
  figmaInfo: {
    key: string;
    token: string;
  };
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
  created_at: string;
};
