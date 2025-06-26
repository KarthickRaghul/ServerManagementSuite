// hooks/server/useServerManagement.ts
import { useState, useEffect } from "react";
import AuthService from "../../auth/auth";
import { useNotification } from "../../context/NotificationContext";

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

interface ServerDevice {
  id: string;
  ip: string;
  tag: string;
  os: string;
  created_at: string;
  access_token?: string;
}

interface CreateServerRequest {
  ip: string;
  tag: string;
  os: string;
}

interface CreateServerResponse {
  access_token: string;
  created_by: string;
  device: ServerDevice;
  message: string;
  status: string;
}

interface DeleteServerResponse {
  deleted_by: string;
  deleted_ip: string;
  message: string;
  status: string;
}

// ✅ Standardized error response interface
interface ErrorResponse {
  status: string;
  message: string;
}

export const useServerManagement = () => {
  const [servers, setServers] = useState<ServerDevice[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [creating, setCreating] = useState<boolean>(false);
  const [deleting, setDeleting] = useState<string | null>(null);
  const { addNotification } = useNotification();

  // ✅ Enhanced fetchServers with standardized error handling
  const fetchServers = async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/server/config1/device`,
        {
          method: "GET",
        },
      );

      if (response.ok) {
        const data = await response.json();

        // ✅ Check for standardized error response
        if (data.status === "failed") {
          throw new Error(data.message || "Failed to fetch servers");
        }

        if (data.status === "success" && data.devices) {
          setServers(data.devices);
        } else {
          setServers([]);
        }
      } else {
        // ✅ Handle HTTP error responses
        const errorData = (await response
          .json()
          .catch(() => ({}))) as ErrorResponse;
        throw new Error(
          errorData.message || `Failed to fetch servers: ${response.status}`,
        );
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to fetch servers";
      console.error("Error fetching servers:", err);
      setError(errorMessage);

      // ✅ Show error notification
      addNotification({
        title: "Server Fetch Error",
        message: errorMessage,
        type: "error",
        duration: 5000,
      });
    } finally {
      setLoading(false);
    }
  };

  // ✅ Enhanced createServer with standardized error handling
  const createServer = async (
    serverData: CreateServerRequest,
  ): Promise<CreateServerResponse | boolean> => {
    setCreating(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config1/create`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(serverData),
        },
      );

      if (response.ok) {
        const data: CreateServerResponse = await response.json();

        // ✅ Check for standardized error response
        if (data.status === "failed") {
          throw new Error(data.message || "Failed to create server");
        }

        if (data.status === "success") {
          // Refresh the server list
          await fetchServers();
          return data; // Return the full response including access_token
        } else {
          throw new Error("Invalid response status from server");
        }
      } else {
        // ✅ Handle HTTP error responses
        const errorData = (await response
          .json()
          .catch(() => ({}))) as ErrorResponse;
        throw new Error(
          errorData.message || `Failed to create server: ${response.status}`,
        );
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to create server";
      console.error("Error creating server:", err);
      setError(errorMessage);

      // ✅ Don't show notification here - let component handle it
      throw err; // Re-throw for component to handle
    } finally {
      setCreating(false);
    }
  };

  // ✅ Enhanced deleteServer with standardized error handling
  const deleteServer = async (ip: string): Promise<boolean> => {
    setDeleting(ip);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config1/delete`,
        {
          method: "DELETE",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ ip }),
        },
      );

      if (response.ok) {
        const data: DeleteServerResponse = await response.json();

        // ✅ Check for standardized error response
        if (data.status === "failed") {
          throw new Error(data.message || "Failed to delete server");
        }

        if (data.status === "success") {
          // Refresh the server list
          await fetchServers();
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
          errorData.message || `Failed to delete server: ${response.status}`,
        );
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to delete server";
      console.error("Error deleting server:", err);
      setError(errorMessage);

      // ✅ Don't show notification here - let component handle it
      throw err; // Re-throw for component to handle
    } finally {
      setDeleting(null);
    }
  };

  useEffect(() => {
    fetchServers();
  }, []);

  return {
    servers,
    loading,
    error,
    creating,
    deleting,
    fetchServers,
    createServer,
    deleteServer,
  };
};
