// types/app.ts
export interface Device {
  id: string;
  tag: string;
  ip: string;
  os: string;
  created_at: string;
}

export type ModeType = 'server' | 'network';

export interface AppContextType {
  activeMode: ModeType;
  updateActiveMode: (mode: ModeType) => void;
  activeDevice: Device | null;
  updateActiveDevice: (device: Device) => void;
  devices: Device[];
  devicesLoading: boolean;
  devicesError: string | null;
  refreshDevices: () => Promise<Device[]>;
}
