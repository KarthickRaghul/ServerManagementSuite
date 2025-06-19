// types/app.ts
export interface Device {
  id: string;
  ip: string;
  tag: string;
  os: string;
  // Add other device properties as needed
}

export interface AppContextType {
  // Remove activeMode and updateActiveMode
  activeDevice: Device | null;
  updateActiveDevice: (device: Device) => void;
  devices: Device[];
  devicesLoading: boolean;
  devicesError: string | null;
  refreshDevices: () => Promise<Device[]>;
}

// Remove ModeType entirely if not used elsewhere
// export type ModeType = 'server' | 'network';
