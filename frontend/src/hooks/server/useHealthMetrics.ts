// hooks/server/useHealthMetrics.ts
import { useState, useEffect } from "react";
import AuthService from "../../auth/auth";
import { useAppContext } from "../../context/AppContext";
import { useNotification } from "../../context/NotificationContext";

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

interface HealthData {
  cpu: {
    usage_percent: number;
  };
  ram: {
    free_mb: number;
    total_mb: number;
    used_mb: number;
    usage_percent: number;
  };
  disk: {
    free_mb: number;
    total_mb: number;
    used_mb: number;
    usage_percent: number;
  };
  net: {
    bytes_recv_mb: number;
    bytes_sent_mb: number;
    name: string;
  };
  open_ports: Array<{
    port: number;
    process: string;
    protocol: string;
  }>;
}

// ✅ SHOULD ADD: Standardized response interface
interface HealthResponse {
  status: string;
  message?: string;
  data?: HealthData;
  // Direct properties for backward compatibility
  cpu?: HealthData["cpu"];
  ram?: HealthData["ram"];
  disk?: HealthData["disk"];
  net?: HealthData["net"];
  open_ports?: HealthData["open_ports"];
}

interface ErrorResponse {
  status: string;
  message: string;
}

interface ProcessedMetrics {
  cpu: number;
  ram: number;
  disk: number;
  network: number;
}

export const useHealthMetrics = () => {
  const [healthData, setHealthData] = useState<HealthData | null>(null);
  const [metrics, setMetrics] = useState<ProcessedMetrics | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const { activeDevice } = useAppContext();
  const { addNotification } = useNotification();

  const fetchHealthData = async () => {
    if (!activeDevice) {
      setLoading(false);
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/server/health`, // ✅ SHOULD UPDATE: Standardized endpoint
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            host: activeDevice.ip,
          }),
        },
      );

      if (response.ok) {
        const responseData: HealthResponse = await response.json();

        // ✅ SHOULD ADD: Standardized error checking
        if (
          responseData.status === "failed" ||
          responseData.status === "error"
        ) {
          throw new Error(
            responseData.message || "Failed to fetch health data",
          );
        }

        // ✅ SHOULD ADD: Handle both wrapped and direct response formats
        const data: HealthData = responseData.data || {
          cpu: responseData.cpu!,
          ram: responseData.ram!,
          disk: responseData.disk!,
          net: responseData.net!,
          open_ports: responseData.open_ports!,
        };

        setHealthData(data);

        // Process the data into metrics
        const processedMetrics: ProcessedMetrics = {
          cpu: Math.round(data.cpu.usage_percent * 100) / 100,
          ram: Math.round(data.ram.usage_percent * 100) / 100,
          disk: Math.round(data.disk.usage_percent * 100) / 100,
          network:
            Math.round(
              (data.net.bytes_sent_mb + data.net.bytes_recv_mb) * 100,
            ) / 100,
        };

        setMetrics(processedMetrics);
      } else {
        // ✅ SHOULD ADD: Enhanced HTTP error handling
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
        err instanceof Error ? err.message : "Failed to fetch health data";
      console.error("Error fetching health data:", err);
      setError(errorMessage);

      // Only show notification for non-network errors to avoid spam
      if (!(err instanceof Error && err.message.includes("Failed to reach"))) {
        addNotification({
          title: "Health Data Error",
          message: errorMessage,
          type: "error",
          duration: 5000,
        });
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (activeDevice) {
      fetchHealthData();
      // Set up polling every 30 seconds
      const interval = setInterval(fetchHealthData, 30000);
      return () => clearInterval(interval);
    } else {
      // ✅ SHOULD ADD: Clear data when no device selected
      setHealthData(null);
      setMetrics(null);
      setError(null);
      setLoading(false);
    }
  }, [activeDevice]);

  const refreshMetrics = async () => {
    await fetchHealthData();
  };

  return {
    healthData,
    metrics,
    loading,
    error,
    refreshMetrics,
  };
};

export default useHealthMetrics;
