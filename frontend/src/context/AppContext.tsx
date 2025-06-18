import React, { createContext, useContext, useState, useEffect } from 'react';
import type { ReactNode } from 'react';
import type { Device, AppContextType } from '../types/app';
import AuthService from '../auth/auth';

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

const AppContext = createContext<AppContextType | undefined>(undefined);

export const AppProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [activeDevice, setActiveDevice] = useState<Device | null>(null);
  const [devices, setDevices] = useState<Device[]>([]);
  const [devicesLoading, setDevicesLoading] = useState<boolean>(false);
  const [devicesError, setDevicesError] = useState<string | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);
  const [authChecked, setAuthChecked] = useState<boolean>(false);

  // Check authentication on mount
  useEffect(() => {
    const checkAuth = async () => {
      try {
        // Skip auth check if on login page
        if (window.location.pathname === '/login') {
          setIsAuthenticated(false);
          setAuthChecked(true);
          return;
        }

        const authorized = await AuthService.authorized();
        setIsAuthenticated(authorized);
      } catch (error) {
        console.error('Auth check failed:', error);
        setIsAuthenticated(false);
      } finally {
        setAuthChecked(true);
      }
    };

    checkAuth();
  }, []);

  // Fetch devices only when authenticated
  useEffect(() => {
    // Don't fetch if auth not checked yet, not authenticated, or on login page
    if (!authChecked || !isAuthenticated || window.location.pathname === '/login') {
      return;
    }

    const fetchDevices = async () => {
      setDevicesLoading(true);
      setDevicesError(null);
      
      try {
        const response = await AuthService.makeAuthenticatedRequest(
          `${BACKEND_URL}/api/admin/server/config1/device`,
          { method: 'GET' },
          false // Don't retry to avoid loops
        );

        if (response.ok) {
          const data = await response.json();
          
          if (data.status === 'success' && data.devices) {
            setDevices(data.devices);
            
            // Auto-select device using functional update
            setActiveDevice(currentDevice => {
              if (!currentDevice && data.devices.length > 0) {
                const storedDeviceId = localStorage.getItem('active_device_id');
                const deviceToSet = storedDeviceId 
                  ? data.devices.find((d: Device) => d.id === storedDeviceId) || data.devices[0]
                  : data.devices[0];
                
                localStorage.setItem('active_device_id', deviceToSet.id);
                return deviceToSet;
              }
              return currentDevice;
            });
          } else {
            setDevices([]);
            setDevicesError('No devices found');
          }
        } else {
          throw new Error(`Failed to fetch devices: ${response.status}`);
        }
      } catch (err) {
        console.error('Error fetching devices:', err);
        setDevices([]);
        setDevicesError(err instanceof Error ? err.message : 'Network error while fetching devices');
        
        // If unauthorized, update auth state
        if (err instanceof Error && err.message.includes('Not authorized')) {
          setIsAuthenticated(false);
        }
      } finally {
        setDevicesLoading(false);
      }
    };

    fetchDevices();
  }, [isAuthenticated, authChecked]);

  const updateActiveDevice = (device: Device) => {
    setActiveDevice(device);
    localStorage.setItem('active_device_id', device.id);
  };

  const refreshDevices = async () => {
    if (!isAuthenticated) {
      console.log('Not authenticated, cannot refresh devices');
      return [];
    }
      
    try {
      setDevicesLoading(true);
      setDevicesError(null);
      
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config1/device`,
        { method: 'GET' },
        false
      );

      if (response.ok) {
        const data = await response.json();
        if (data.status === 'success' && data.devices) {
          setDevices(data.devices);
          
          setActiveDevice(currentDevice => {
            if (currentDevice) {
              const deviceStillExists = data.devices.find((d: Device) => d.id === currentDevice.id);
              if (deviceStillExists) {
                return currentDevice;
              }
            }
            
            if (data.devices.length > 0) {
              const newDevice = data.devices[0];
              localStorage.setItem('active_device_id', newDevice.id);
              return newDevice;
            }
            
            return null;
          });
          
          return data.devices;
        }
      } else {
        throw new Error(`Failed to refresh devices: ${response.status}`);
      }
    } catch (err) {
      console.error('Error refreshing devices:', err);
      setDevicesError(err instanceof Error ? err.message : 'Failed to refresh devices');
    } finally {
      setDevicesLoading(false);
    }
    return [];
  };

  return (
    <AppContext.Provider value={{
      activeDevice,
      updateActiveDevice,
      devices,
      devicesLoading,
      devicesError,
      refreshDevices
    }}>
      {children}
    </AppContext.Provider>
  );
};

export const useAppContext = () => {
  const context = useContext(AppContext);
  if (context === undefined) {
    throw new Error('useAppContext must be used within an AppProvider');
  }
  return context;
};
