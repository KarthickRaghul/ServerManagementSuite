// pages/settings/Types.ts
export interface User {
  id: number;
  username: string;
  role: 'admin' | 'viewer';
  email: string;
  created_at?: string;
  last_login?: string;
}
