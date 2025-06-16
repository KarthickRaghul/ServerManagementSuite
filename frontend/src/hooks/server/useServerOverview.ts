// hooks/useServerOverview.ts
import { useState, useEffect } from 'react';
import AuthService from '../../auth/auth';
import { useAppContext } from '../../context/AppContext';

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

interface ServerOverviewData {
  status: string;
  uptime: string;
}

export const useServerOverview = () => {
  const [data, setData] = useState<ServerOverviewData | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const { activeDevice } = useAppContext();

  const fetchOverview = async () => {
    if (!activeDevice) {
      setLoading(false);
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config1/overview`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ host: activeDevice.ip })
        }
      );

      if (response.ok) {
        const responseData = await response.json();
        setData({
          status: responseData.status,
          uptime: responseData.uptime
        });
      } else {
        throw new Error(`Failed to fetch server overview: ${response.status}`);
      }
    } catch (err) {
      console.error('Error fetching server overview:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch server overview');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchOverview();
    
    // Set up polling every 30 seconds
    const interval = setInterval(fetchOverview, 30000);
    
    return () => clearInterval(interval);
  }, [activeDevice]);

  const refresh = () => {
    fetchOverview();
  };

  return { 
    data, 
    loading, 
    error, 
    refresh 
  };
};
