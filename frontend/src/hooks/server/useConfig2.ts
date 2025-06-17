// hooks/server/useConfig2.ts
import { useState, useEffect } from 'react';
import AuthService from '../../auth/auth';
import { useAppContext } from '../../context/AppContext';
import { useNotification } from '../../context/NotificationContext';

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

interface FirewallChain {
  name: string;
  policy: string;
  rules: FirewallRule[];
}

interface FirewallData {
  type: string;
  chains: FirewallChain[];
  active: boolean;
}

interface FirewallRule {
  chain: string;
  number: number;
  target: string;
  protocol: string;
  source: string;
  destination: string;
  options: string;
  state?: string;
}

// Updated RouteEntry interface to match your API response
interface RouteEntry {
  destination: string;
  gateway: string;
  genmask: string;
  flags: string;
  metric: string;
  ref: string;
  use: string;
  iface: string;
}

export const useConfig2 = () => {
  const [networkBasics, setNetworkBasics] = useState<NetworkBasics | null>(null);
  const [routeTable, setRouteTable] = useState<RouteEntry[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const { activeDevice } = useAppContext();
  const { addNotification } = useNotification();
  const [firewallData, setFirewallData] = useState<FirewallData | null>(null);

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
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch network basics';
      console.error('Error fetching network basics:', err);
      setError(errorMessage);
      addNotification({
        title: 'Network Fetch Error',
        message: errorMessage,
        type: 'error',
        duration: 5000
      });
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
        if (Array.isArray(data)) {
          setRouteTable(data);
        } else {
          setRouteTable([]);
        }
      } else {
        throw new Error(`Failed to fetch route table: ${response.status}`);
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch route table';
      console.error('Error fetching route table:', err);
      setError(errorMessage);
      addNotification({
        title: 'Route Table Fetch Error',
        message: errorMessage,
        type: 'error',
        duration: 5000
      });
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
    if (!activeDevice) {
      addNotification({
        title: 'Network Update Error',
        message: 'No active device selected',
        type: 'error',
        duration: 5000
      });
      return false;
    }

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
          addNotification({
            title: 'Network Updated',
            message: 'Network configuration has been updated successfully',
            type: 'success',
            duration: 3000
          });
          return true;
        } else {
          throw new Error(data.message || 'Network update failed');
        }
      } else {
        throw new Error(`Failed to update network: ${response.status}`);
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to update network';
      console.error('Error updating network:', err);
      setError(errorMessage);
      addNotification({
        title: 'Network Update Failed',
        message: errorMessage,
        type: 'error',
        duration: 5000
      });
      return false;
    } finally {
      setLoading(false);
    }
  };

  // Update Interface
  const updateInterface = async (interfaceName: string, status: string): Promise<boolean> => {
    if (!activeDevice) {
      addNotification({
        title: 'Interface Update Error',
        message: 'No active device selected',
        type: 'error',
        duration: 5000
      });
      return false;
    }

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
          addNotification({
            title: 'Interface Updated',
            message: `Interface ${interfaceName} has been ${status}d successfully`,
            type: 'success',
            duration: 3000
          });
          return true;
        } else {
          throw new Error(data.message || `Failed to ${status} interface`);
        }
      } else {
        throw new Error(`Failed to update interface: ${response.status}`);
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : `Failed to ${status} interface`;
      console.error('Error updating interface:', err);
      setError(errorMessage);
      addNotification({
        title: 'Interface Update Failed',
        message: errorMessage,
        type: 'error',
        duration: 5000
      });
      return false;
    } finally {
      setLoading(false);
    }
  };

  // Restart Interface
  const restartInterface = async (): Promise<boolean> => {
    if (!activeDevice) {
      addNotification({
        title: 'Interface Restart Error',
        message: 'No active device selected',
        type: 'error',
        duration: 5000
      });
      return false;
    }

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
          addNotification({
            title: 'Interface Restarted',
            message: 'Network interface has been restarted successfully',
            type: 'success',
            duration: 3000
          });
          return true;
        } else {
          throw new Error(data.message || 'Failed to restart interface');
        }
      } else {
        throw new Error(`Failed to restart interface: ${response.status}`);
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to restart interface';
      console.error('Error restarting interface:', err);
      setError(errorMessage);
      addNotification({
        title: 'Interface Restart Failed',
        message: errorMessage,
        type: 'error',
        duration: 5000
      });
      return false;
    } finally {
      setLoading(false);
    }
  };


  // Add these functions to your existing hook

// Fetch Firewall Rules
const fetchFirewallRules = async () => {
  if (!activeDevice) return;

  setLoading(true);
  setError(null);

  try {
    const response = await AuthService.makeAuthenticatedRequest(
      `${BACKEND_URL}/api/admin/server/config2/getfirewall`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ host: activeDevice.ip })
      }
    );

    if (response.ok) {
      const data: FirewallData = await response.json();
      setFirewallData(data);
    } else {
      throw new Error(`Failed to fetch firewall rules: ${response.status}`);
    }
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : 'Failed to fetch firewall rules';
    console.error('Error fetching firewall rules:', err);
    setError(errorMessage);
    addNotification({
      title: 'Firewall Fetch Error',
      message: errorMessage,
      type: 'error',
      duration: 5000
    });
  } finally {
    setLoading(false);
  }
};

// Update Firewall Rule
const updateFirewallRule = async (firewallData: {
  action: string;
  rule: string;
  protocol: string;
  port: string;
  source?: string;
  destination?: string;
}): Promise<boolean> => {
  if (!activeDevice) {
    addNotification({
      title: 'Firewall Update Error',
      message: 'No active device selected',
      type: 'error',
      duration: 5000
    });
    return false;
  }

  setLoading(true);
  setError(null);

  try {
    const response = await AuthService.makeAuthenticatedRequest(
      `${BACKEND_URL}/api/admin/server/config2/postupdatefirewall`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          host: activeDevice.ip,
          ...firewallData
        })
      }
    );

    if (response.ok) {
      const data = await response.json();
      if (data.status === 'success') {
        await fetchFirewallRules();
        addNotification({
          title: 'Firewall Rule Updated',
          message: `Firewall rule has been ${firewallData.action}ed successfully`,
          type: 'success',
          duration: 3000
        });
        return true;
      } else {
        throw new Error(data.message || `Failed to ${firewallData.action} firewall rule`);
      }
    } else {
      throw new Error(`Failed to update firewall rule: ${response.status}`);
    }
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : `Failed to ${firewallData.action} firewall rule`;
    console.error('Error updating firewall rule:', err);
    setError(errorMessage);
    addNotification({
      title: 'Firewall Update Failed',
      message: errorMessage,
      type: 'error',
      duration: 5000
    });
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
    if (!activeDevice) {
      addNotification({
        title: 'Route Update Error',
        message: 'No active device selected',
        type: 'error',
        duration: 5000
      });
      return false;
    }

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
          await fetchRouteTable();
          addNotification({
            title: 'Route Updated',
            message: `Route has been ${routeData.action}ed successfully`,
            type: 'success',
            duration: 3000
          });
          return true;
        } else {
          throw new Error(data.message || `Failed to ${routeData.action} route`);
        }
      } else {
        throw new Error(`Failed to update route: ${response.status}`);
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : `Failed to ${routeData.action} route`;
      console.error('Error updating route:', err);
      setError(errorMessage);
      addNotification({
        title: 'Route Update Failed',
        message: errorMessage,
        type: 'error',
        duration: 5000
      });
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
      fetchFirewallRules();
    }
  }, [activeDevice]);

  return {
    networkBasics,
    routeTable,
    firewallData,
    loading,
    error,
    fetchNetworkBasics,
    fetchRouteTable,
    fetchFirewallRules,
    updateInterface,
    updateNetwork,
    restartInterface,
    updateRoute,
    updateFirewallRule
  };
};
