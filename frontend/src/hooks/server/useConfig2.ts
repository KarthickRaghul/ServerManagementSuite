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

interface RouteEntry {
  destination: string;
  gateway: string;
  interface: string;
  metric: string;
}

export const useConfig2 = () => {
  const [networkBasics, setNetworkBasics] = useState<NetworkBasics | null>(null);
  const [routeTable, setRouteTable] = useState<RouteEntry[]>([]);
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
        `${BACKEND_URL}/api/admin/server/config2/getnetworkbasics`,
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

  // Fetch Route Table
  const fetchRouteTable = async () => {
    if (!activeDevice) return;

    setLoading(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config2/getroute`,
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
        setRouteTable(data.routes || []);
      } else {
        throw new Error(`Failed to fetch route table: ${response.status}`);
      }
    } catch (err) {
      console.error('Error fetching route table:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch route table');
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

  // Update Interface
  const updateInterface = async (interfaceName: string, status: string): Promise<boolean> => {
    if (!activeDevice) return false;

    setLoading(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config2/postinterface`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            host: activeDevice.ip,
            interface: interfaceName,
            status: status
          })
        }
      );

      if (response.ok) {
        const data = await response.json();
        if (data.status === 'success') {
          await fetchNetworkBasics();
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
        `${BACKEND_URL}/api/admin/server/config2/postrestartinterface`,
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
          await fetchNetworkBasics();
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

  // Update Route
  const updateRoute = async (routeData: {
    action: string;
    destination: string;
    gateway: string;
    interface?: string;
    metric?: string;
  }): Promise<boolean> => {
    if (!activeDevice) return false;

    setLoading(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config2/postupdateroute`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            host: activeDevice.ip,
            ...routeData
          })
        }
      );

      if (response.ok) {
        const data = await response.json();
        if (data.status === 'success') {
          await fetchRouteTable(); // Refresh route table
          return true;
        }
        return false;
      } else {
        throw new Error(`Failed to update route: ${response.status}`);
      }
    } catch (err) {
      console.error('Error updating route:', err);
      setError(err instanceof Error ? err.message : 'Failed to update route');
      return false;
    } finally {
      setLoading(false);
    }
  };

  // Fetch all data on mount
  useEffect(() => {
    if (activeDevice) {
      fetchNetworkBasics();
      fetchRouteTable();
    }
  }, [activeDevice]);

  return {
    networkBasics,
    routeTable,
    loading,
    error,
    fetchNetworkBasics,
    fetchRouteTable,
    updateInterface,
    updateNetwork,
    restartInterface,
    updateRoute
  };
};
