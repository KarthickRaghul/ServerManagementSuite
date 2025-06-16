// hooks/useRole.ts
import { useState, useEffect } from 'react';
import AuthService from '../../auth/auth';

export type UserRole = 'admin' | 'viewer' | null;

export const useRole = () => {
  const [role, setRole] = useState<UserRole>(null);
  const [userInfo, setUserInfo] = useState<{ username?: string; role?: UserRole }>({});
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const getUserRole = () => {
      setLoading(true);
      try {
        const userRole = AuthService.getUserRole();
        const userDetails = AuthService.getUserInfo();
        
        setRole(userRole);
        setUserInfo(userDetails || {});
      } catch (error) {
        console.error('Error getting user role:', error);
        setRole(null);
        setUserInfo({});
      } finally {
        setLoading(false);
      }
    };

    getUserRole();
  }, []);

  const isAdmin = role === 'admin';
  const isViewer = role === 'viewer';
  const canAccessRoute = (path: string) => AuthService.canAccessRoute(path);

  return {
    role,
    userInfo,
    loading,
    isAdmin,
    isViewer,
    canAccessRoute
  };
};
