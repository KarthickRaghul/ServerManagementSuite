// hooks/server/useServerConfiguration.ts
import { useState, useEffect } from "react";
import AuthService from "../../auth/auth";
import { useAppContext } from "../../context/AppContext";
import { useNotification } from "../../context/NotificationContext";

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

interface ServerConfigData {
  hostname: string;
  timezone: string;
}

interface UpdateConfigData {
  hostname: string;
  timezone: string;
}

// ✅ Standardized error response interface
interface ErrorResponse {
  status: string;
  message: string;
}

export const useServerConfiguration = () => {
  const [data, setData] = useState<ServerConfigData | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [updating, setUpdating] = useState<boolean>(false);
  const { activeDevice } = useAppContext();
  const { addNotification } = useNotification();

  // ✅ Enhanced fetchConfiguration with standardized error handling
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
            responseData.message || "Failed to fetch server configuration",
          );
        }

        setData({
          hostname: responseData.hostname?.trim() || "",
          timezone: responseData.timezone || "",
        });
      } else {
        // ✅ Handle HTTP error responses
        const errorData = (await response
          .json()
          .catch(() => ({}))) as ErrorResponse;
        throw new Error(
          errorData.message ||
            `Failed to fetch server configuration: ${response.status}`,
        );
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error
          ? err.message
          : "Failed to fetch server configuration";
      console.error("Error fetching server configuration:", err);
      setError(errorMessage);

      // ✅ Show error notification for fetch errors
      addNotification({
        title: "Configuration Fetch Error",
        message: errorMessage,
        type: "error",
        duration: 5000,
      });
    } finally {
      setLoading(false);
    }
  };

  // ✅ Enhanced updateConfiguration with standardized error handling
  const updateConfiguration = async (
    configData: UpdateConfigData,
  ): Promise<boolean> => {
    if (!activeDevice) {
      throw new Error("No active device selected");
    }

    setUpdating(true);
    setError(null);

    try {
      console.log("🔍 Sending config update:", {
        host: activeDevice.ip,
        ...configData,
      });

      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config1/basic_update`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            host: activeDevice.ip,
            hostname: configData.hostname,
            timezone: configData.timezone,
          }),
        },
      );

      const responseData = await response.json();
      console.log("🔍 Config update response:", responseData);

      if (response.ok) {
        // ✅ Check for standardized error response
        if (responseData.status === "failed") {
          throw new Error(
            responseData.message || "Configuration update failed",
          );
        }

        if (responseData.status === "success") {
          // Refresh the data after successful update
          await fetchConfiguration();
          return true;
        } else {
          throw new Error("Invalid response status from server");
        }
      } else {
        // ✅ Handle HTTP error responses
        const errorData = (await response
          .json()
          .catch(() => ({}))) as ErrorResponse;
        throw new Error(
          errorData.message || `Server error: ${response.status}`,
        );
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error
          ? err.message
          : "Failed to update server configuration";
      console.error("Error updating server configuration:", err);
      setError(errorMessage);

      // ✅ Re-throw error for component to handle notifications
      throw err;
    } finally {
      setUpdating(false);
    }
  };

  useEffect(() => {
    if (activeDevice) {
      fetchConfiguration();
    } else {
      // Clear data when no device is selected
      setData(null);
      setError(null);
    }
  }, [activeDevice]);

  return {
    data,
    loading,
    error,
    updating,
    fetchConfiguration,
    updateConfiguration,
  };
};
