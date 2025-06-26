// hooks/server/useResourceOptimization.ts
import { useState, useEffect } from "react";
import AuthService from "../../auth/auth";
import { useAppContext } from "../../context/AppContext";
import { useNotification } from "../../context/NotificationContext";

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

interface CleanupInfo {
  failed: string | null;
  folders: string[];
  sizes: {
    [folder: string]: number;
  };
}

interface Service {
  pid: number;
  user: string;
  name: string;
  cmdline: string;
}

// ✅ Standardized response interfaces
interface CleanupInfoResponse {
  status: string;
  message?: string;
  data?: CleanupInfo;
  failed?: string | null;
  folders?: string[];
  sizes?: { [folder: string]: number };
}

interface ServicesResponse {
  status: string;
  message?: string;
  data?: Service[];
  services?: Service[];
  timestamp?: string;
}

interface OptimizeResponse {
  status: "success" | "partial" | "failed";
  message: string;
  details?: string;
}

interface RestartServiceResponse {
  status: string;
  message: string;
  service?: string;
  timestamp?: string;
}

// ✅ Standardized error response interface
interface ErrorResponse {
  status: string;
  message: string;
}

export const useResourceOptimization = () => {
  const [cleanupInfo, setCleanupInfo] = useState<CleanupInfo | null>(null);
  const [services, setServices] = useState<Service[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [optimizing, setOptimizing] = useState<boolean>(false);
  const [restartingService, setRestartingService] = useState<string | null>(
    null,
  );
  const { activeDevice } = useAppContext();
  const { addNotification } = useNotification();

  // ✅ Enhanced fetchCleanupInfo with standardized error handling
  const fetchCleanupInfo = async () => {
    if (!activeDevice) {
      setLoading(false);
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/resource/cleaninfo`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ host: activeDevice.ip }),
        },
      );

      if (response.ok) {
        const data: CleanupInfoResponse = await response.json();

        // ✅ Enhanced status checking - handle more response types
        if (data.status === "failed" || data.status === "error") {
          throw new Error(data.message || "Failed to fetch cleanup info");
        }

        // ✅ Handle multiple success status variations
        if (data.status === "success" || data.status === "ok" || !data.status) {
          // Handle both wrapped and direct response formats
          const cleanupData: CleanupInfo = data.data || {
            failed: data.failed || null,
            folders: data.folders || [],
            sizes: data.sizes || {},
          };
          setCleanupInfo(cleanupData);
        } else {
          // ✅ Log the actual response for debugging
          console.warn("Unexpected response status:", data.status, data);
          throw new Error(`Unexpected response status: ${data.status}`);
        }
      } else {
        // ✅ Enhanced HTTP error handling
        const errorData = (await response
          .json()
          .catch(() => ({}))) as ErrorResponse;
        throw new Error(
          errorData.message ||
            `HTTP ${response.status}: ${response.statusText}`,
        );
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to fetch cleanup info";
      console.error("Error fetching cleanup info:", err);
      setError(errorMessage);

      // ✅ Only show notification for non-network errors to avoid spam
      if (!(err instanceof Error && err.message.includes("Failed to reach"))) {
        addNotification({
          title: "Cleanup Info Error",
          message: errorMessage,
          type: "error",
          duration: 5000,
        });
      }
    } finally {
      setLoading(false);
    }
  }; // ✅ Enhanced fetchServices with standardized error handling
  const fetchServices = async () => {
    if (!activeDevice) {
      setLoading(false);
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/resource/service`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ host: activeDevice.ip }),
        },
      );

      if (response.ok) {
        const data: ServicesResponse = await response.json();

        // ✅ Check for standardized error response
        if (data.status === "failed") {
          throw new Error(data.message || "Failed to fetch services");
        }

        if (data.status === "success") {
          // ✅ Handle both wrapped and direct response formats
          const servicesData = data.data || data.services || [];
          setServices(servicesData);
        } else {
          throw new Error("Invalid response status from server");
        }
      } else {
        // ✅ Enhanced HTTP error handling
        const errorData = (await response
          .json()
          .catch(() => ({}))) as ErrorResponse;
        throw new Error(
          errorData.message || `Failed to fetch services: ${response.status}`,
        );
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to fetch services";
      console.error("Error fetching services:", err);
      setError(errorMessage);

      // ✅ Only show notification for non-network errors to avoid spam
      if (!(err instanceof Error && err.message.includes("Failed to reach"))) {
        addNotification({
          title: "Services Error",
          message: errorMessage,
          type: "error",
          duration: 5000,
        });
      }
    } finally {
      setLoading(false);
    }
  };

  // ✅ Enhanced optimizeSystem with standardized error handling
  const optimizeSystem = async (): Promise<boolean> => {
    if (!activeDevice) return false;

    setOptimizing(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/resource/optimize`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ host: activeDevice.ip }),
        },
      );

      if (response.ok) {
        const data: OptimizeResponse = await response.json();

        // ✅ Handle different response statuses including 'failed'
        switch (data.status) {
          case "success":
            addNotification({
              title: "System Optimized",
              message: data.message,
              type: "success",
              duration: 4000,
            });
            await fetchCleanupInfo(); // Refresh cleanup info
            return true;

          case "partial":
            addNotification({
              title: "Optimization Partially Completed",
              message:
                data.message +
                (data.details ? ` Details: ${data.details}` : ""),
              type: "warning",
              duration: 6000,
            });
            await fetchCleanupInfo(); // Refresh cleanup info
            return true; // Still consider it successful

          case "failed":
            throw new Error(data.message || "System optimization failed");

          default:
            throw new Error("Unknown response status from server");
        }
      } else {
        // ✅ Enhanced HTTP error handling
        const errorData = (await response
          .json()
          .catch(() => ({}))) as ErrorResponse;
        throw new Error(
          errorData.message || `Failed to optimize system: ${response.status}`,
        );
      }
    } catch (err) {
      console.error("Error optimizing system:", err);
      const errorMessage =
        err instanceof Error ? err.message : "Failed to optimize system";
      setError(errorMessage);
      addNotification({
        title: "Optimization Failed",
        message: errorMessage,
        type: "error",
        duration: 5000,
      });
      return false;
    } finally {
      setOptimizing(false);
    }
  };

  // ✅ Enhanced restartService with standardized error handling
  const restartService = async (serviceName: string): Promise<boolean> => {
    if (!activeDevice) return false;

    setRestartingService(serviceName);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/resource/restartservice`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            host: activeDevice.ip,
            service: serviceName,
          }),
        },
      );

      if (response.ok) {
        const data: RestartServiceResponse = await response.json();

        // ✅ Check for standardized error response
        if (data.status === "failed") {
          throw new Error(data.message || "Service restart failed");
        }

        if (data.status === "success") {
          addNotification({
            title: "Service Restarted",
            message:
              data.message || `Service ${serviceName} restarted successfully`,
            type: "success",
            duration: 4000,
          });
          await fetchServices(); // Refresh services
          return true;
        } else {
          throw new Error(
            `Service restart failed: ${data.message || "Unknown error"}`,
          );
        }
      } else {
        // ✅ Enhanced HTTP error handling
        const errorData = (await response
          .json()
          .catch(() => ({}))) as ErrorResponse;
        throw new Error(
          errorData.message || `Failed to restart service: ${response.status}`,
        );
      }
    } catch (err) {
      console.error("Error restarting service:", err);
      const errorMessage =
        err instanceof Error ? err.message : "Failed to restart service";
      setError(errorMessage);
      addNotification({
        title: "Restart Failed",
        message: `Failed to restart ${serviceName}: ${errorMessage}`,
        type: "error",
        duration: 5000,
      });
      return false;
    } finally {
      setRestartingService(null);
    }
  };

  // ✅ Enhanced refresh function with better error handling
  const refreshData = async () => {
    try {
      await Promise.all([fetchCleanupInfo(), fetchServices()]);
      addNotification({
        title: "Data Refreshed",
        message: "Resource optimization data has been refreshed",
        type: "success",
        duration: 3000,
      });
    } catch (err) {
      // Individual functions handle their own errors
      console.error("Error refreshing data:", err);
    }
  };

  useEffect(() => {
    if (activeDevice) {
      refreshData();
    } else {
      // Clear data when no device is selected
      setCleanupInfo(null);
      setServices([]);
      setError(null);
    }
  }, [activeDevice]);

  return {
    cleanupInfo,
    services,
    loading,
    error,
    optimizing,
    restartingService,
    fetchCleanupInfo,
    fetchServices,
    optimizeSystem,
    restartService,
    refreshData,
  };
};
