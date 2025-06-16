// hooks/useServerConfiguration.ts
import { useState, useEffect } from 'react';
import AuthService from '../../auth/auth';
import { useAppContext } from '../../context/AppContext';

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

interface ServerConfigData {
  hostname: string;
  timezone: string;
}

interface UpdateConfigData {
  host: string;
  hostname: string;
  timezone: string;
}

export const useServerConfiguration = () => {
  const [data, setData] = useState<ServerConfigData | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [updating, setUpdating] = useState<boolean>(false);
  const { activeDevice } = useAppContext();

  const fetchConfiguration = async () => {
    if (!activeDevice) {
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config1/basic`,
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
          hostname: responseData.hostname?.trim() || '',
          timezone: responseData.timezone || ''
        });
      } else {
        throw new Error(`Failed to fetch server configuration: ${response.status}`);
      }
    } catch (err) {
      console.error('Error fetching server configuration:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch server configuration');
    } finally {
      setLoading(false);
    }
  };

  const updateConfiguration = async (configData: Omit<UpdateConfigData, 'host'>): Promise<boolean> => {
    if (!activeDevice) {
      throw new Error('No active device selected');
    }

    setUpdating(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config1/basic_update`,
        {
          method: 'POST', // or POST if your API expects POST for updates
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            host: activeDevice.ip,
            hostname: configData.hostname,
            timezone: configData.timezone
          })
        }
      );

      if (response.ok) {
        const responseData = await response.json();
        if (responseData.status === 'success') {
          // Refresh the data after successful update
          await fetchConfiguration();
          return true;
        } else {
          throw new Error('Update failed: Invalid response status');
        }
      } else {
        throw new Error(`Failed to update server configuration: ${response.status}`);
      }
    } catch (err) {
      console.error('Error updating server configuration:', err);
      setError(err instanceof Error ? err.message : 'Failed to update server configuration');
      return false;
    } finally {
      setUpdating(false);
    }
  };

  useEffect(() => {
    fetchConfiguration();
  }, [activeDevice]);

  return {
    data,
    loading,
    error,
    updating,
    fetchConfiguration,
    updateConfiguration
  };
};
