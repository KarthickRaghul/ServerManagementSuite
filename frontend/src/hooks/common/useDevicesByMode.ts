// hooks/useDevicesByMode.ts
import { useState, useEffect } from 'react';
import type { Device } from '../../types/app';
import AuthService from '../../auth/auth';

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

export const useDevicesByMode = (mode: 'server' | 'network') => {
  const [devices, setDevices] = useState<Device[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchDevices = async () => {
      setLoading(true);
      setError(null);
      try {
        const endpoint = mode === 'server' ? '/api/admin/server/config1/device' : '/api/admin/network/config1/device';
        
        // Get auth headers from AuthService
        const authHeaders = AuthService.getAuthHeader();
        
        const response = await fetch(`${BACKEND_URL}${endpoint}`, {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
            ...authHeaders, // Spread the auth headers
          },
          credentials: 'include',
        });

        if (response.ok) {
          const data = await response.json();
          if (data.status === 'success' && data.devices) {
            setDevices(data.devices);
          } else {
            setError('Failed to fetch devices');
          }
        } else if (response.status === 401) {
          // Handle unauthorized - token might be expired
          console.log('Unauthorized request, attempting to refresh token...');
          const refreshed = await AuthService.authorized();
          if (refreshed) {
            // Retry the request with new token
            const newAuthHeaders = AuthService.getAuthHeader();
            const retryResponse = await fetch(`${BACKEND_URL}${endpoint}`, {
              method: 'GET',
              headers: {
                'Content-Type': 'application/json',
                ...newAuthHeaders,
              },
              credentials: 'include',
            });
            
            if (retryResponse.ok) {
              const retryData = await retryResponse.json();
              if (retryData.status === 'success' && retryData.devices) {
                setDevices(retryData.devices);
              } else {
                setError('Failed to fetch devices after retry');
              }
            } else {
              setError(`Failed to fetch devices after retry: ${retryResponse.status}`);
            }
          } else {
            setError('Authentication failed');
          }
        } else {
          setError(`Failed to fetch devices: ${response.status}`);
        }
      } catch (err) {
        console.error('Error fetching devices:', err);
        setError('Network error while fetching devices');
      } finally {
        setLoading(false);
      }
    };

    fetchDevices();
  }, [mode]);

  return { devices, loading, error };
};
