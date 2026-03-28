export type WorshipOrderItem = {
  key: string;
  title: string;
  obj: string;
  info: string;
  lead?: string;
  contents?: string;
  children?: WorshipOrderItem[];
};

export type OBSStatus = { connected: boolean; currentScene: string };

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
