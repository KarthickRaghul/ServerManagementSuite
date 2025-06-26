// hooks/server/useAlerts.ts
import { useState, useEffect } from "react";
import AuthService from "../../auth/auth";
import { useAppContext } from "../../context/AppContext";
import { useNotification } from "../../context/NotificationContext";

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

interface Alert {
  id: number;
  host: string;
  severity: "warning" | "critical" | "info";
  content: string;
  status: "notseen" | "seen";
  time: string;
}

interface AlertsResponse {
  status: string;
  alerts: Alert[];
  count: number;
  message?: string;
}

interface MarkSeenResponse {
  status: string;
  message: string;
  count?: number;
}

// ✅ Standardized error response interface
interface ErrorResponse {
  status: string;
  message: string;
}

export const useAlerts = () => {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [markingAsSeen, setMarkingAsSeen] = useState<number[]>([]);
  const [resolving, setResolving] = useState<number[]>([]);
  const { activeDevice } = useAppContext();
  const { addNotification } = useNotification();

  // ✅ Enhanced fetchAlerts with standardized error handling
  const fetchAlerts = async (onlyUnseen = false, limit = 50) => {
    if (!activeDevice) {
      setLoading(false);
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const requestBody: {
        host: string;
        limit: number;
        only_unseen?: boolean;
      } = {
        host: activeDevice.ip,
        limit,
      };

      if (onlyUnseen) {
        requestBody.only_unseen = true;
      }

      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/server/alerts`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(requestBody),
        },
      );

      if (response.ok) {
        const data: AlertsResponse = await response.json();

        // ✅ Check for standardized error response
        if (data.status === "failed") {
          throw new Error(data.message || "Failed to fetch alerts");
        }

        if (data.status === "success") {
          setAlerts(data.alerts || []);
        } else {
          throw new Error("Invalid response status from server");
        }
      } else {
        // ✅ Enhanced HTTP error handling
        const errorData = (await response
          .json()
          .catch(() => ({}))) as ErrorResponse;
        throw new Error(
          errorData.message || `Failed to fetch alerts: ${response.status}`,
        );
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to fetch alerts";
      console.error("Error fetching alerts:", err);
      setError(errorMessage);

      // ✅ Only show notification for non-network errors to avoid spam
      if (!(err instanceof Error && err.message.includes("Failed to reach"))) {
        addNotification({
          title: "Alert Fetch Error",
          message: errorMessage,
          type: "error",
          duration: 5000,
        });
      }
    } finally {
      setLoading(false);
    }
  };

  // ✅ Enhanced markAlertsAsSeen with standardized error handling
  const markAlertsAsSeen = async (alertIds: number[]): Promise<boolean> => {
    setMarkingAsSeen(alertIds);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/server/alerts/markseen`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ alert_ids: alertIds }),
        },
      );

      if (response.ok) {
        const data: MarkSeenResponse = await response.json();

        // ✅ Check for standardized error response
        if (data.status === "failed") {
          throw new Error(data.message || "Failed to mark alerts as seen");
        }

        if (data.status === "success") {
          // Update local state
          setAlerts((prev) =>
            prev.map((alert) =>
              alertIds.includes(alert.id)
                ? { ...alert, status: "seen" as const }
                : alert,
            ),
          );

          addNotification({
            title: "Alerts Acknowledged",
            message: `${alertIds.length} alert(s) have been acknowledged`,
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
            `Failed to mark alerts as seen: ${response.status}`,
        );
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to mark alerts as seen";
      console.error("Error marking alerts as seen:", err);
      setError(errorMessage);

      addNotification({
        title: "Acknowledge Failed",
        message: errorMessage,
        type: "error",
        duration: 5000,
      });

      return false;
    } finally {
      setMarkingAsSeen([]);
    }
  };

  // ✅ Enhanced markSingleAlertAsSeen with standardized error handling
  const markSingleAlertAsSeen = async (alertId: number): Promise<boolean> => {
    setMarkingAsSeen([alertId]);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/server/alerts/marksingleseen?id=${alertId}`,
        {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
          },
        },
      );

      if (response.ok) {
        const data: MarkSeenResponse = await response.json();

        // ✅ Check for standardized error response
        if (data.status === "failed") {
          throw new Error(data.message || "Failed to mark alert as seen");
        }

        if (data.status === "success") {
          setAlerts((prev) =>
            prev.map((alert) =>
              alert.id === alertId
                ? { ...alert, status: "seen" as const }
                : alert,
            ),
          );

          addNotification({
            title: "Alert Acknowledged",
            message: "Alert has been marked as seen",
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
            `Failed to mark alert as seen: ${response.status}`,
        );
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to mark alert as seen";
      console.error("Error marking alert as seen:", err);
      setError(errorMessage);

      addNotification({
        title: "Acknowledge Failed",
        message: errorMessage,
        type: "error",
        duration: 5000,
      });

      return false;
    } finally {
      setMarkingAsSeen([]);
    }
  };

  // ✅ Enhanced resolveAlerts with standardized error handling
  const resolveAlerts = async (alertIds: number[]): Promise<boolean> => {
    setResolving(alertIds);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/server/alerts/delete`,
        {
          method: "DELETE",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ alert_ids: alertIds }),
        },
      );

      if (response.ok) {
        const data: MarkSeenResponse = await response.json();

        // ✅ Check for standardized error response
        if (data.status === "failed") {
          throw new Error(data.message || "Failed to resolve alerts");
        }

        if (data.status === "success") {
          // Remove resolved alerts from local state
          setAlerts((prev) =>
            prev.filter((alert) => !alertIds.includes(alert.id)),
          );

          addNotification({
            title: "Alerts Resolved",
            message: `${alertIds.length} alert(s) have been resolved and deleted`,
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
          errorData.message || `Failed to resolve alerts: ${response.status}`,
        );
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to resolve alerts";
      console.error("Error resolving alerts:", err);
      setError(errorMessage);

      addNotification({
        title: "Resolve Failed",
        message: errorMessage,
        type: "error",
        duration: 5000,
      });

      return false;
    } finally {
      setResolving([]);
    }
  };

  // ✅ Enhanced refresh function
  const refreshAlerts = async () => {
    await fetchAlerts();
  };

  useEffect(() => {
    if (activeDevice) {
      fetchAlerts();
    } else {
      // Clear alerts when no device is selected
      setAlerts([]);
      setError(null);
    }
  }, [activeDevice]);

  return {
    alerts,
    loading,
    error,
    markingAsSeen,
    resolving,
    fetchAlerts,
    markAlertsAsSeen,
    markSingleAlertAsSeen,
    resolveAlerts,
    refreshAlerts, // ✅ Added refresh function
  };
};
