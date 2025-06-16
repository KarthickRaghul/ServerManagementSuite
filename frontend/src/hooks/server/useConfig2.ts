// hooks/server/useConfig2.ts
import { useState, useEffect } from 'react';
import AuthService from '../../auth/auth';
import { useAppContext } from '../../context/AppContext';

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

// Updated Types based on your API response
interface NetworkBasics {
  ip_method: string;
  ip_address: string;
  gateway: string;
  subnet: string;
  dns: string;
  uptime: string;
  interface: {
    [key: string]: {
      mode: string;
      status: string;
    };
  };
}

export const useConfig2 = () => {
  const [networkBasics, setNetworkBasics] = useState<NetworkBasics | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const { activeDevice } = useAppContext();

  // Fetch Network Basics
  const fetchNetworkBasics = async () => {
    if (!activeDevice) return;

    setLoading(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/getnetworkbasics`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ host: activeDevice.ip })
        }
      );

      if (response.ok) {
        const data: NetworkBasics = await response.json();
        setNetworkBasics(data);
      } else {
        throw new Error(`Failed to fetch network basics: ${response.status}`);
      }
    } catch (err) {
      console.error('Error fetching network basics:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch network basics');
    } finally {
      setLoading(false);
    }
  };

  // Update Network Configuration
  const updateNetwork = async (networkData: {
    method: string;
    ip?: string;
    subnet?: string;
    gateway?: string;
    dns?: string;
  }): Promise<boolean> => {
    if (!activeDevice) return false;

    setLoading(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config2/postnetwork`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            host: activeDevice.ip,
            ...networkData
          })
        }
      );

      if (response.ok) {
        const data = await response.json();
        if (data.status === 'success') {
          // Refresh network basics after successful update
          await fetchNetworkBasics();
          return true;
        }
        return false;
      } else {
        throw new Error(`Failed to update network: ${response.status}`);
      }
    } catch (err) {
      console.error('Error updating network:', err);
      setError(err instanceof Error ? err.message : 'Failed to update network');
      return false;
    } finally {
      setLoading(false);
    }
  };

  // Update Interface - Fixed to use 'disable'/'enable' instead of 'down'/'up'
  const updateInterface = async (interfaceName: string, status: string): Promise<boolean> => {
    if (!activeDevice) return false;

    setLoading(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/postinterface`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            host: activeDevice.ip,
            interface: interfaceName,
            status: status // This should be 'disable' or 'enable'
          })
        }
      );

      if (response.ok) {
        const data = await response.json();
        if (data.status === 'success') {
          await fetchNetworkBasics(); // Refresh to get updated interface status
          return true;
        }
        return false;
      } else {
        throw new Error(`Failed to update interface: ${response.status}`);
      }
    } catch (err) {
      console.error('Error updating interface:', err);
      setError(err instanceof Error ? err.message : 'Failed to update interface');
      return false;
    } finally {
      setLoading(false);
    }
  };

  // Restart Interface
  const restartInterface = async (): Promise<boolean> => {
    if (!activeDevice) return false;

    setLoading(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/postrestartinterface`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ host: activeDevice.ip })
        }
      );

      if (response.ok) {
        const data = await response.json();
        if (data.status === 'success') {
          await fetchNetworkBasics(); // Refresh after restart
          return true;
        }
        return false;
      } else {
        throw new Error(`Failed to restart interface: ${response.status}`);
      }
    } catch (err) {
      console.error('Error restarting interface:', err);
      setError(err instanceof Error ? err.message : 'Failed to restart interface');
      return false;
    } finally {
      setLoading(false);
    }
  };

  // Fetch all data on mount
  useEffect(() => {
    if (activeDevice) {
      fetchNetworkBasics();
    }
  }, [activeDevice]);

  return {
    networkBasics,
    loading,
    error,
    fetchNetworkBasics,
    updateInterface,
    updateNetwork,
    restartInterface
  };
};
