// hooks/server/useConfig2.ts
import { useState, useEffect } from 'react';
import AuthService from '../../auth/auth';
import { useAppContext } from '../../context/AppContext';
import { useNotification } from '../../context/NotificationContext';

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

// Proper TypeScript interfaces
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

interface FirewallChain {
  name: string;
  policy: string;
  rules: FirewallRule[];
}

interface LinuxFirewallData {
  type: 'iptables';
  chains: FirewallChain[];
  active: boolean;
}

interface WindowsFirewallRule {
  Name: string;
  DisplayName: string;
  Direction: 'Inbound' | 'Outbound';
  Action: 'Allow' | 'Block';
  Enabled: 'True' | 'False';
  Profile: 'Public' | 'Private' | 'Domain' | 'Any';
}

type FirewallData = LinuxFirewallData | WindowsFirewallRule[];

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

interface NetworkUpdateData {
  method: string;
  ip?: string;
  subnet?: string;
  gateway?: string;
  dns?: string;
}

interface RouteUpdateData {
  action: string;
  destination: string;
  gateway: string;
  interface?: string;
  metric?: string;
}

interface LinuxFirewallUpdateData {
  action: string;
  rule: string;
  protocol: string;
  port: string;
  source?: string;
  destination?: string;
}

interface WindowsFirewallUpdateData {
  action: string;
  name?: string;
  displayName?: string;
  direction?: 'Inbound' | 'Outbound';
  action?: 'Allow' | 'Block';
  enabled?: 'True' | 'False';
  profile?: 'Public' | 'Private' | 'Domain' | 'Any';
  protocol?: 'TCP' | 'UDP' | 'Any';
  localPort?: string;
  remotePort?: string;
  localAddress?: string;
  remoteAddress?: string;
  program?: string;
  service?: string;
}

type FirewallUpdateData = LinuxFirewallUpdateData | WindowsFirewallUpdateData;

interface LoadingStates {
  networkBasics: boolean;
  routeTable: boolean;
  firewallData: boolean;
  updating: boolean;
}

export const useConfig2 = () => {
  const [networkBasics, setNetworkBasics] = useState<NetworkBasics | null>(null);
  const [routeTable, setRouteTable] = useState<RouteEntry[]>([]);
  const [firewallData, setFirewallData] = useState<FirewallData | null>(null);
  const [loading, setLoading] = useState<LoadingStates>({
    networkBasics: false,
    routeTable: false,
    firewallData: false,
    updating: false
  });
  const [error, setError] = useState<string | null>(null);
  const { activeDevice } = useAppContext();
  const { addNotification } = useNotification();

  // Fetch Network Basics
  const fetchNetworkBasics = async () => {
    if (!activeDevice) return;

    setLoading(prev => ({ ...prev, networkBasics: true }));
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
      setLoading(prev => ({ ...prev, networkBasics: false }));
    }
  };

  // Fetch Route Table
  const fetchRouteTable = async () => {
    if (!activeDevice) return;

    setLoading(prev => ({ ...prev, routeTable: true }));
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
      setLoading(prev => ({ ...prev, routeTable: false }));
    }
  };

  // Fetch Firewall Rules
  const fetchFirewallRules = async () => {
    if (!activeDevice) return;

    setLoading(prev => ({ ...prev, firewallData: true }));
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
        const data = await response.json();
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
      setLoading(prev => ({ ...prev, firewallData: false }));
    }
  };

  // Update Network Configuration
  const updateNetwork = async (networkData: NetworkUpdateData): Promise<boolean> => {
    if (!activeDevice) {
      addNotification({
        title: 'Network Update Error',
        message: 'No active device selected',
        type: 'error',
        duration: 5000
      });
      return false;
    }

    setLoading(prev => ({ ...prev, updating: true }));
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
      setLoading(prev => ({ ...prev, updating: false }));
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

    setLoading(prev => ({ ...prev, updating: true }));
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
      setLoading(prev => ({ ...prev, updating: false }));
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

    setLoading(prev => ({ ...prev, updating: true }));
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
      setLoading(prev => ({ ...prev, updating: false }));
    }
  };

  // Update Route
  const updateRoute = async (routeData: RouteUpdateData): Promise<boolean> => {
    if (!activeDevice) {
      addNotification({
        title: 'Route Update Error',
        message: 'No active device selected',
        type: 'error',
        duration: 5000
      });
      return false;
    }

    setLoading(prev => ({ ...prev, updating: true }));
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
      setLoading(prev => ({ ...prev, updating: false }));
    }
  };

  // Update Firewall Rule
  const updateFirewallRule = async (firewallUpdateData: FirewallUpdateData): Promise<boolean> => {
    if (!activeDevice) {
      addNotification({
        title: 'Firewall Update Error',
        message: 'No active device selected',
        type: 'error',
        duration: 5000
      });
      return false;
    }

    setLoading(prev => ({ ...prev, updating: true }));
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
            ...firewallUpdateData
          })
        }
      );

      if (response.ok) {
        const data = await response.json();
        if (data.status === 'success') {
          await fetchFirewallRules();
          addNotification({
            title: 'Firewall Rule Updated',
            message: `Firewall rule has been ${firewallUpdateData.action}ed successfully`,
            type: 'success',
            duration: 3000
          });
          return true;
        } else {
          throw new Error(data.message || `Failed to ${firewallUpdateData.action} firewall rule`);
        }
      } else {
        throw new Error(`Failed to update firewall rule: ${response.status}`);
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : `Failed to ${firewallUpdateData.action} firewall rule`;
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
      setLoading(prev => ({ ...prev, updating: false }));
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
