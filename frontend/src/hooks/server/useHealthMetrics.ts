// hooks/useHealthMetrics.ts
import { useState, useEffect } from 'react';
import AuthService from '../../auth/auth';
import { useAppContext } from '../../context/AppContext';

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

interface HealthData {
  cpu: {
    usage_percent: number;
  };
  ram: {
    total_mb: number;
    used_mb: number;
    free_mb: number;
    usage_percent: number;
  };
  disk: {
    total_mb: number;
    used_mb: number;
    free_mb: number;
    usage_percent: number;
  };
  net: {
    name: string;
    bytes_sent_mb: number;
    bytes_recv_mb: number;
  };
  open_ports: Array<{
    protocol: string;
    port: number;
    process: string;
  }>;
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

  const fetchHealthData = async () => {
    if (!activeDevice) {
      setLoading(false);
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/server/health`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            host: activeDevice.ip
          })
        }
      );

      if (response.ok) {
        const data: HealthData = await response.json();
        setHealthData(data);
        
        // Process the data into metrics
        const processedMetrics: ProcessedMetrics = {
          cpu: Math.round(data.cpu.usage_percent * 100) / 100,
          ram: Math.round(data.ram.usage_percent * 100) / 100,
          disk: Math.round(data.disk.usage_percent * 100) / 100,
          network: Math.round((data.net.bytes_sent_mb + data.net.bytes_recv_mb) * 100) / 100
        };
        
        setMetrics(processedMetrics);
      } else {
        throw new Error(`Failed to fetch health data: ${response.status}`);
      }
    } catch (err) {
      console.error('Error fetching health data:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch health data');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchHealthData();
    
    // Set up polling every 30 seconds
    const interval = setInterval(fetchHealthData, 30000);
    
    return () => clearInterval(interval);
  }, [activeDevice]);

  const refreshMetrics = () => {
    fetchHealthData();
  };

  return {
    healthData,
    metrics,
    loading,
    error,
    refreshMetrics
  };
};
