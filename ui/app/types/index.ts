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
