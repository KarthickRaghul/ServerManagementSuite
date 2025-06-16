// hooks/server/useAlerts.ts
import { useState, useEffect } from 'react';
import AuthService from '../../auth/auth';
import { useAppContext } from '../../context/AppContext';

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

interface Alert {
  id: number;
  host: string;
  severity: 'warning' | 'critical' | 'info';
  content: string;
  status: 'notseen' | 'seen';
  time: string;
}

interface AlertsResponse {
  status: string;
  alerts: Alert[];
  count: number;
}

interface MarkSeenResponse {
  status: string;
  message: string;
  count: number;
}

export const useAlerts = () => {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [markingAsSeen, setMarkingAsSeen] = useState<number[]>([]);
  const [resolving, setResolving] = useState<number[]>([]);
  const { activeDevice } = useAppContext();

  const fetchAlerts = async (onlyUnseen = false, limit = 50) => {
    if (!activeDevice) return;

    setLoading(true);
    setError(null);

    try {
      const requestBody: { host: string; limit: number; only_unseen?: boolean } = {
        host: activeDevice.ip,
        limit
      };

      if (onlyUnseen) {
        requestBody.only_unseen = true;
      }

      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/server/alerts`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(requestBody)
        }
      );

      if (response.ok) {
        const data: AlertsResponse = await response.json();
        if (data.status === 'success') {
          setAlerts(data.alerts);
        } else {
          throw new Error('Failed to fetch alerts: Invalid response status');
        }
      } else {
        throw new Error(`Failed to fetch alerts: ${response.status}`);
      }
    } catch (err) {
      console.error('Error fetching alerts:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch alerts');
    } finally {
      setLoading(false);
    }
  };

  const markAlertsAsSeen = async (alertIds: number[]): Promise<boolean> => {
    setMarkingAsSeen(alertIds);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/server/alerts/markseen`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ alert_ids: alertIds })
        }
      );

      if (response.ok) {
        const data: MarkSeenResponse = await response.json();
        if (data.status === 'success') {
          // Update local state
          setAlerts(prev => prev.map(alert => 
            alertIds.includes(alert.id) 
              ? { ...alert, status: 'seen' }
              : alert
          ));
          return true;
        } else {
          throw new Error('Failed to mark alerts as seen');
        }
      } else {
        throw new Error(`Failed to mark alerts as seen: ${response.status}`);
      }
    } catch (err) {
      console.error('Error marking alerts as seen:', err);
      setError(err instanceof Error ? err.message : 'Failed to mark alerts as seen');
      return false;
    } finally {
      setMarkingAsSeen([]);
    }
  };

  const markSingleAlertAsSeen = async (alertId: number): Promise<boolean> => {
    setMarkingAsSeen([alertId]);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/server/alerts/marksingleseen?id=${alertId}`,
        {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json',
          }
        }
      );

      if (response.ok) {
        const data: MarkSeenResponse = await response.json();
        if (data.status === 'success') {
          setAlerts(prev => prev.map(alert => 
            alert.id === alertId 
              ? { ...alert, status: 'seen' }
              : alert
          ));
          return true;
        } else {
          throw new Error('Failed to mark alert as seen');
        }
      } else {
        throw new Error(`Failed to mark alert as seen: ${response.status}`);
      }
    } catch (err) {
      console.error('Error marking alert as seen:', err);
      setError(err instanceof Error ? err.message : 'Failed to mark alert as seen');
      return false;
    } finally {
      setMarkingAsSeen([]);
    }
  };

  const resolveAlerts = async (alertIds: number[]): Promise<boolean> => {
    setResolving(alertIds);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/server/alerts/delete`,
        {
          method: 'DELETE',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ alert_ids: alertIds })
        }
      );

      if (response.ok) {
        const data: MarkSeenResponse = await response.json();
        if (data.status === 'success') {
          // Remove resolved alerts from local state
          setAlerts(prev => prev.filter(alert => !alertIds.includes(alert.id)));
          return true;
        } else {
          throw new Error('Failed to resolve alerts');
        }
      } else {
        throw new Error(`Failed to resolve alerts: ${response.status}`);
      }
    } catch (err) {
      console.error('Error resolving alerts:', err);
      setError(err instanceof Error ? err.message : 'Failed to resolve alerts');
      return false;
    } finally {
      setResolving([]);
    }
  };

  useEffect(() => {
    if (activeDevice) {
      fetchAlerts();
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
    resolveAlerts
  };
};
