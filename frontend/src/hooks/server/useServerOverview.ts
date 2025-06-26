// hooks/server/useServerOverview.ts
import { useState, useEffect } from "react";
import AuthService from "../../auth/auth";
import { useAppContext } from "../../context/AppContext";
import { useNotification } from "../../context/NotificationContext";

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

interface ServerOverviewData {
  status: string;
  uptime: string;
}

// ✅ Standardized error response interface
interface ErrorResponse {
  status: string;
  message: string;
}

export const useServerOverview = () => {
  const [data, setData] = useState<ServerOverviewData | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const { activeDevice } = useAppContext();
  const { addNotification } = useNotification();

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
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ host: activeDevice.ip }),
        },
      );

      if (response.ok) {
        const responseData = await response.json();

        // ✅ Check for standardized error response
        if (responseData.status === "failed") {
          throw new Error(
            responseData.message || "Failed to fetch server overview",
          );
        }

        setData({
          status: responseData.status,
          uptime: responseData.uptime,
        });
      } else {
        // ✅ Handle HTTP error responses
        const errorData = (await response
          .json()
          .catch(() => ({}))) as ErrorResponse;
        throw new Error(
          errorData.message ||
            `Failed to fetch server overview: ${response.status}`,
        );
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to fetch server overview";
      console.error("Error fetching server overview:", err);
      setError(errorMessage);

      // ✅ Show error notification only for non-network errors
      if (!(err instanceof Error && err.message.includes("Failed to reach"))) {
        addNotification({
          title: "Server Overview Error",
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
      fetchOverview();

      // Set up polling every 30 seconds
      const interval = setInterval(fetchOverview, 30000);

      return () => clearInterval(interval);
    } else {
      // Clear data when no device is selected
      setData(null);
      setError(null);
      setLoading(false);
    }
  }, [activeDevice]);

  const refresh = () => {
    fetchOverview();
  };

  return {
    data,
    loading,
    error,
    refresh,
  };
};
