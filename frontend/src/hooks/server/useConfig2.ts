// hooks/server/useConfig2.ts
import { useState, useEffect } from "react";
import AuthService from "../../auth/auth";
import { useAppContext } from "../../context/AppContext";
import { useNotification } from "../../context/NotificationContext";

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

// ✅ Updated interface to match actual backend response
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
      power: string;
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
  type: "iptables";
  chains: FirewallChain[];
  active: boolean;
}

interface WindowsFirewallRule {
  Name: string;
  DisplayName: string;
  Direction: "Inbound" | "Outbound";
  Action: "Allow" | "Block";
  Enabled: "True" | "False";
  Profile: "Public" | "Private" | "Domain" | "Any";
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
  direction?: "Inbound" | "Outbound";
  actionType?: "Allow" | "Block";
  enabled?: "True" | "False";
  profile?: "Public" | "Private" | "Domain" | "Any";
  protocol?: "TCP" | "UDP" | "Any";
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

// ✅ Standardized response interface
interface StandardResponse {
  status: string;
  message?: string;
  error?: string;
  details?: string;
}

// ✅ Enhanced error response interface
interface ErrorResponse {
  status: string;
  message: string;
}

export const useConfig2 = () => {
  const [networkBasics, setNetworkBasics] = useState<NetworkBasics | null>(
    null,
  );
  const [routeTable, setRouteTable] = useState<RouteEntry[]>([]);
  const [firewallData, setFirewallData] = useState<FirewallData | null>(null);
  const [loading, setLoading] = useState<LoadingStates>({
    networkBasics: false,
    routeTable: false,
    firewallData: false,
    updating: false,
  });
  const [error, setError] = useState<string | null>(null);
  const { activeDevice } = useAppContext();
  const { addNotification } = useNotification();

  // ✅ Enhanced Fetch Network Basics with better error handling
  const fetchNetworkBasics = async () => {
    if (!activeDevice) return;

    setLoading((prev) => ({ ...prev, networkBasics: true }));
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config2/getnetworkbasics`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ host: activeDevice.ip }),
        },
      );

      if (response.ok) {
        const data = await response.json();

        // ✅ Check for standardized error response
        if (data.status === "failed") {
          throw new Error(data.message || "Failed to fetch network basics");
        }

        // ✅ Handle direct data response (GET-like endpoint)
        setNetworkBasics(data as NetworkBasics);
      } else {
        // ✅ Enhanced HTTP error handling
        const errorData = (await response
          .json()
          .catch(() => ({}))) as ErrorResponse;
        throw new Error(
          errorData.message ||
            `Failed to fetch network basics: ${response.status}`,
        );
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to fetch network basics";
      console.error("Error fetching network basics:", err);
      setError(errorMessage);

      // ✅ Only show notification for non-network errors to avoid spam
      if (!(err instanceof Error && err.message.includes("Failed to reach"))) {
        addNotification({
          title: "Network Fetch Error",
          message: errorMessage,
          type: "error",
          duration: 5000,
        });
      }
    } finally {
      setLoading((prev) => ({ ...prev, networkBasics: false }));
    }
  };

  // ✅ Enhanced Fetch Route Table with better error handling
  const fetchRouteTable = async () => {
    if (!activeDevice) return;

    setLoading((prev) => ({ ...prev, routeTable: true }));
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config2/getroute`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ host: activeDevice.ip }),
        },
      );

      if (response.ok) {
        const data = await response.json();

        // ✅ Check for standardized error response
        if (data.status === "failed") {
          throw new Error(data.message || "Failed to fetch route table");
        }

        // ✅ Handle array response
        if (Array.isArray(data)) {
          setRouteTable(data);
        } else if (data.routes && Array.isArray(data.routes)) {
          setRouteTable(data.routes);
        } else {
          setRouteTable([]);
        }
      } else {
        // ✅ Enhanced HTTP error handling
        const errorData = (await response
          .json()
          .catch(() => ({}))) as ErrorResponse;
        throw new Error(
          errorData.message ||
            `Failed to fetch route table: ${response.status}`,
        );
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to fetch route table";
      console.error("Error fetching route table:", err);
      setError(errorMessage);

      // ✅ Only show notification for non-network errors
      if (!(err instanceof Error && err.message.includes("Failed to reach"))) {
        addNotification({
          title: "Route Table Fetch Error",
          message: errorMessage,
          type: "error",
          duration: 5000,
        });
      }
    } finally {
      setLoading((prev) => ({ ...prev, routeTable: false }));
    }
  };

  // ✅ Enhanced Fetch Firewall Rules with better error handling
  const fetchFirewallRules = async () => {
    if (!activeDevice) return;

    setLoading((prev) => ({ ...prev, firewallData: true }));
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config2/getfirewall`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ host: activeDevice.ip }),
        },
      );

      if (response.ok) {
        const data = await response.json();

        // ✅ Check for standardized error response
        if (data.status === "failed") {
          throw new Error(data.message || "Failed to fetch firewall rules");
        }

        setFirewallData(data);
      } else {
        // ✅ Enhanced HTTP error handling
        const errorData = (await response
          .json()
          .catch(() => ({}))) as ErrorResponse;
        throw new Error(
          errorData.message ||
            `Failed to fetch firewall rules: ${response.status}`,
        );
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to fetch firewall rules";
      console.error("Error fetching firewall rules:", err);
      setError(errorMessage);

      // ✅ Only show notification for non-network errors
      if (!(err instanceof Error && err.message.includes("Failed to reach"))) {
        addNotification({
          title: "Firewall Fetch Error",
          message: errorMessage,
          type: "error",
          duration: 5000,
        });
      }
    } finally {
      setLoading((prev) => ({ ...prev, firewallData: false }));
    }
  };

  // ✅ Enhanced Network Update with better error handling
  const updateNetwork = async (
    networkData: NetworkUpdateData,
  ): Promise<boolean> => {
    if (!activeDevice) {
      addNotification({
        title: "Network Update Error",
        message: "No active device selected",
        type: "error",
        duration: 5000,
      });
      return false;
    }

    setLoading((prev) => ({ ...prev, updating: true }));
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config2/postnetwork`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            host: activeDevice.ip,
            ...networkData,
          }),
        },
      );

      if (response.ok) {
        const data: StandardResponse = await response.json();

        // ✅ Check for standardized error response
        if (data.status === "failed") {
          throw new Error(data.message || "Network update failed");
        }

        if (data.status === "success") {
          await fetchNetworkBasics();
          addNotification({
            title: "Network Updated",
            message: "Network configuration has been updated successfully",
            type: "success",
            duration: 3000,
          });
          return true;
        } else {
          throw new Error("Invalid response status from server");
        }
      } else {
        // ✅ Enhanced HTTP error handling
        const errorData = (await response
          .json()
          .catch(() => ({}))) as ErrorResponse;
        throw new Error(
          errorData.message || `Failed to update network: ${response.status}`,
        );
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to update network";
      console.error("Error updating network:", err);
      setError(errorMessage);
      addNotification({
        title: "Network Update Failed",
        message: errorMessage,
        type: "error",
        duration: 5000,
      });
      return false;
    } finally {
      setLoading((prev) => ({ ...prev, updating: false }));
    }
  };

  // ✅ Enhanced Interface Update with better error handling
  const updateInterface = async (
    interfaceName: string,
    status: string,
  ): Promise<boolean> => {
    if (!activeDevice) {
      addNotification({
        title: "Interface Update Error",
        message: "No active device selected",
        type: "error",
        duration: 5000,
      });
      return false;
    }

    setLoading((prev) => ({ ...prev, updating: true }));
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config2/postinterface`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            host: activeDevice.ip,
            interface: interfaceName,
            status: status,
          }),
        },
      );

      if (response.ok) {
        const data: StandardResponse = await response.json();

        // ✅ Check for standardized error response
        if (data.status === "failed") {
          throw new Error(data.message || `Failed to ${status} interface`);
        }

        if (data.status === "success") {
          await fetchNetworkBasics();
          addNotification({
            title: "Interface Updated",
            message: `Interface ${interfaceName} has been ${status}d successfully`,
            type: "success",
            duration: 3000,
          });
          return true;
        } else {
          throw new Error("Invalid response status from server");
        }
      } else {
        // ✅ Enhanced HTTP error handling
        const errorData = (await response
          .json()
          .catch(() => ({}))) as ErrorResponse;
        throw new Error(
          errorData.message || `Failed to update interface: ${response.status}`,
        );
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : `Failed to ${status} interface`;
      console.error("Error updating interface:", err);
      setError(errorMessage);
      addNotification({
        title: "Interface Update Failed",
        message: errorMessage,
        type: "error",
        duration: 5000,
      });
      return false;
    } finally {
      setLoading((prev) => ({ ...prev, updating: false }));
    }
  };

  // ✅ Enhanced Restart Interface with better error handling
  const restartInterface = async (): Promise<boolean> => {
    if (!activeDevice) {
      addNotification({
        title: "Interface Restart Error",
        message: "No active device selected",
        type: "error",
        duration: 5000,
      });
      return false;
    }

    setLoading((prev) => ({ ...prev, updating: true }));
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config2/postrestartinterface`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ host: activeDevice.ip }),
        },
      );

      if (response.ok) {
        const data: StandardResponse = await response.json();

        // ✅ Check for standardized error response
        if (data.status === "failed") {
          throw new Error(data.message || "Failed to restart interface");
        }

        if (data.status === "success") {
          await fetchNetworkBasics();
          addNotification({
            title: "Interface Restarted",
            message: "Network interface has been restarted successfully",
            type: "success",
            duration: 3000,
          });
          return true;
        } else {
          throw new Error("Invalid response status from server");
        }
      } else {
        // ✅ Enhanced HTTP error handling
        const errorData = (await response
          .json()
          .catch(() => ({}))) as ErrorResponse;
        throw new Error(
          errorData.message ||
            `Failed to restart interface: ${response.status}`,
        );
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to restart interface";
      console.error("Error restarting interface:", err);
      setError(errorMessage);
      addNotification({
        title: "Interface Restart Failed",
        message: errorMessage,
        type: "error",
        duration: 5000,
      });
      return false;
    } finally {
      setLoading((prev) => ({ ...prev, updating: false }));
    }
  };

  // ✅ Enhanced Route Update with better error handling
  const updateRoute = async (routeData: RouteUpdateData): Promise<boolean> => {
    if (!activeDevice) {
      addNotification({
        title: "Route Update Error",
        message: "No active device selected",
        type: "error",
        duration: 5000,
      });
      return false;
    }

    setLoading((prev) => ({ ...prev, updating: true }));
    setError(null);

    try {
      console.log("🔍 Sending route update:", {
        host: activeDevice.ip,
        ...routeData,
      });

      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config2/postupdateroute`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            host: activeDevice.ip,
            ...routeData,
          }),
        },
      );

      const data: StandardResponse = await response.json();
      console.log("🔍 Route update response:", data);

      if (response.ok) {
        // ✅ Check for standardized error response
        if (data.status === "failed") {
          throw new Error(
            data.message || `Failed to ${routeData.action} route`,
          );
        }

        if (data.status === "success") {
          await fetchRouteTable();
          addNotification({
            title: "Route Updated",
            message: `Route has been ${routeData.action}ed successfully`,
            type: "success",
            duration: 3000,
          });
          return true;
        } else {
          throw new Error("Invalid response status from server");
        }
      } else {
        // ✅ Enhanced HTTP error handling
        throw new Error(data.message || `Server error: ${response.status}`);
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error
          ? err.message
          : `Failed to ${routeData.action} route`;
      console.error("Error updating route:", err);
      setError(errorMessage);
      addNotification({
        title: "Route Update Failed",
        message: errorMessage,
        type: "error",
        duration: 5000,
      });
      return false;
    } finally {
      setLoading((prev) => ({ ...prev, updating: false }));
    }
  };

  // ✅ Enhanced Firewall Rule Update with better error handling
  const updateFirewallRule = async (
    firewallUpdateData: FirewallUpdateData,
  ): Promise<boolean> => {
    if (!activeDevice) {
      addNotification({
        title: "Firewall Update Error",
        message: "No active device selected",
        type: "error",
        duration: 5000,
      });
      return false;
    }

    setLoading((prev) => ({ ...prev, updating: true }));
    setError(null);

    try {
      console.log("🔍 Sending firewall update:", {
        host: activeDevice.ip,
        ...firewallUpdateData,
      });

      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config2/postupdatefirewall`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            host: activeDevice.ip,
            ...firewallUpdateData,
          }),
        },
      );

      const data: StandardResponse = await response.json();
      console.log("🔍 Firewall update response:", data);

      if (response.ok) {
        // ✅ Check for standardized error response
        if (data.status === "failed") {
          throw new Error(
            data.message ||
              `Failed to ${firewallUpdateData.action} firewall rule`,
          );
        }

        if (data.status === "success") {
          await fetchFirewallRules();
          addNotification({
            title: "Firewall Rule Updated",
            message: `Firewall rule has been ${firewallUpdateData.action}ed successfully`,
            type: "success",
            duration: 3000,
          });
          return true;
        } else {
          throw new Error("Invalid response status from server");
        }
      } else {
        // ✅ Enhanced HTTP error handling
        throw new Error(data.message || `Server error: ${response.status}`);
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error
          ? err.message
          : `Failed to ${firewallUpdateData.action} firewall rule`;
      console.error("Error updating firewall rule:", err);
      setError(errorMessage);
      addNotification({
        title: "Firewall Update Failed",
        message: errorMessage,
        type: "error",
        duration: 5000,
      });
      return false;
    } finally {
      setLoading((prev) => ({ ...prev, updating: false }));
    }
  };

  // ✅ Enhanced data refresh function
  const refreshAllData = async () => {
    if (activeDevice) {
      await Promise.all([
        fetchNetworkBasics(),
        fetchRouteTable(),
        fetchFirewallRules(),
      ]);
    }
  };

  // ✅ Enhanced useEffect with cleanup
  useEffect(() => {
    if (activeDevice) {
      refreshAllData();
    } else {
      // Clear data when no device is selected
      setNetworkBasics(null);
      setRouteTable([]);
      setFirewallData(null);
      setError(null);
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
    updateFirewallRule,
    refreshAllData, // ✅ Added refresh function
  };
};
