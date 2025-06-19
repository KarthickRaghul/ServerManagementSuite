// context/ConnectionOverlayContext.tsx
import React, { createContext, useContext, useState, useEffect } from "react";
import { useAppContext } from "./AppContext";
import AuthService from "../auth/auth";

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

interface OverlayState {
  visible: boolean;
  loading: boolean;
  error: string | null;
  deviceIp: string | null;
  isInitialCheck: boolean;
  isConnected: boolean;
}

const initialState: OverlayState = {
  visible: false,
  loading: false,
  error: null,
  deviceIp: null,
  isInitialCheck: false,
  isConnected: false,
};

const ConnectionOverlayContext = createContext<{
  state: OverlayState;
  show: (ip: string, isInitial?: boolean) => void;
  hide: () => void;
  setError: (error: string) => void;
  setLoading: (loading: boolean) => void;
  checkConnection: (ip: string, isInitial?: boolean) => Promise<void>;
  isConnected: boolean;
}>({
  state: initialState,
  show: () => {},
  hide: () => {},
  setError: () => {},
  setLoading: () => {},
  checkConnection: async () => {},
  isConnected: false,
});

export const useConnectionOverlay = () => useContext(ConnectionOverlayContext);

export const ConnectionOverlayProvider: React.FC<{children: React.ReactNode}> = ({ children }) => {
  const [state, setState] = useState(initialState);
  const { activeDevice } = useAppContext();

  const show = (ip: string, isInitial = false) => 
    setState({ visible: true, loading: true, error: null, deviceIp: ip, isInitialCheck: isInitial, isConnected: false });
  
  const hide = () => setState(prev => ({ ...prev, visible: false }));
  
  const setError = (error: string) => 
    setState(s => ({ ...s, error, loading: false, isConnected: false }));
  
  const setLoading = (loading: boolean) => 
    setState(s => ({ ...s, loading }));

  const checkConnection = async (ip: string, isInitial = false) => {
    show(ip, isInitial);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/server/check`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ host: ip }),
        }
      );

      if (response.ok) {
        const data = await response.json();
        if (data.status === "success") {
          setState(s => ({ ...s, loading: false, isConnected: true, error: null }));
          setTimeout(() => {
            setState(s => ({ ...s, visible: false }));
          }, 1000);
        } else {
          setError(data.message || "Could not connect to the client. Please check your network or device settings.");
        }
      } else {
        setError("Failed to connect to the server. Please check your network connection.");
      }
    } catch (err) {
      console.error('Connection check failed:', err);
      setError("Could not connect to the client. Please check your network or device settings.");
    }
  };

  // Auto-check on active device change
  useEffect(() => {
    if (activeDevice?.ip) {
      checkConnection(activeDevice.ip, true);
    }
  }, [activeDevice?.id]);

  return (
    <ConnectionOverlayContext.Provider value={{ 
      state, 
      show, 
      hide, 
      setError, 
      setLoading, 
      checkConnection,
      isConnected: state.isConnected
    }}>
      {children}
    </ConnectionOverlayContext.Provider>
  );
};
