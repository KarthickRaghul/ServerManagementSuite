// hooks/server/useResourceOptimization.ts
import { useState, useEffect } from 'react';
import AuthService from '../../auth/auth';
import { useAppContext } from '../../context/AppContext';

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

interface ServicesResponse {
  status: string;
  message: string;
  services: Service[];
  timestamp: string;
}

interface OptimizeResponse {
  message: string;
  status: string;
}

interface RestartServiceResponse {
  service: string;
  status: string;
  message: string;
  timestamp: string;
}

export const useResourceOptimization = () => {
  const [cleanupInfo, setCleanupInfo] = useState<CleanupInfo | null>(null);
  const [services, setServices] = useState<Service[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [optimizing, setOptimizing] = useState<boolean>(false);
  const [restartingService, setRestartingService] = useState<string | null>(null);
  const { activeDevice } = useAppContext();

  const fetchCleanupInfo = async () => {
    if (!activeDevice) return;

    setLoading(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/resource/cleaninfo`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ host: activeDevice.ip })
        }
      );

      if (response.ok) {
        const data: CleanupInfo = await response.json();
        setCleanupInfo(data);
      } else {
        throw new Error(`Failed to fetch cleanup info: ${response.status}`);
      }
    } catch (err) {
      console.error('Error fetching cleanup info:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch cleanup info');
    } finally {
      setLoading(false);
    }
  };

  const fetchServices = async () => {
    if (!activeDevice) return;

    setLoading(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/resource/service`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ host: activeDevice.ip })
        }
      );

      if (response.ok) {
        const data: ServicesResponse = await response.json();
        if (data.status === 'success') {
          setServices(data.services);
        } else {
          throw new Error('Failed to fetch services: Invalid response status');
        }
      } else {
        throw new Error(`Failed to fetch services: ${response.status}`);
      }
    } catch (err) {
      console.error('Error fetching services:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch services');
    } finally {
      setLoading(false);
    }
  };

  const optimizeSystem = async (): Promise<boolean> => {
    if (!activeDevice) return false;

    setOptimizing(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/resource/optimize`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ host: activeDevice.ip })
        }
      );

      if (response.ok) {
        const data: OptimizeResponse = await response.json();
        if (data.status === 'success') {
          // Refresh cleanup info after optimization
          await fetchCleanupInfo();
          return true;
        } else {
          throw new Error('System optimization failed: Invalid response status');
        }
      } else {
        throw new Error(`Failed to optimize system: ${response.status}`);
      }
    } catch (err) {
      console.error('Error optimizing system:', err);
      setError(err instanceof Error ? err.message : 'Failed to optimize system');
      return false;
    } finally {
      setOptimizing(false);
    }
  };

  const restartService = async (serviceName: string): Promise<boolean> => {
    if (!activeDevice) return false;

    setRestartingService(serviceName);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/resource/restart`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ 
            host: activeDevice.ip,
            service: serviceName
          })
        }
      );

      if (response.ok) {
        const data: RestartServiceResponse = await response.json();
        if (data.status === 'success') {
          // Refresh services after restart
          await fetchServices();
          return true;
        } else {
          throw new Error(`Service restart failed: ${data.message}`);
        }
      } else {
        throw new Error(`Failed to restart service: ${response.status}`);
      }
    } catch (err) {
      console.error('Error restarting service:', err);
      setError(err instanceof Error ? err.message : 'Failed to restart service');
      return false;
    } finally {
      setRestartingService(null);
    }
  };

  const refreshData = async () => {
    await Promise.all([fetchCleanupInfo(), fetchServices()]);
  };

  useEffect(() => {
    if (activeDevice) {
      refreshData();
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
    refreshData
  };
};
