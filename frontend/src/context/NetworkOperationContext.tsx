// context/NetworkOperationContext.tsx
import React, { createContext, useContext, useState } from "react";

interface NetworkOperationState {
  visible: boolean;
  operationType: "interface_restart" | "network_config" | null;
  loading: boolean;
  countdown: number;
  message: string;
}

interface NetworkOperationContextType {
  state: NetworkOperationState;
  showInterfaceRestartOverlay: () => void;
  showNetworkConfigOverlay: () => void;
  hide: () => void;
}

const NetworkOperationContext = createContext<NetworkOperationContextType | undefined>(undefined);

export const useNetworkOperation = () => {
  const context = useContext(NetworkOperationContext);
  if (!context) {
    throw new Error("useNetworkOperation must be used within NetworkOperationProvider");
  }
  return context;
};

export const NetworkOperationProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [state, setState] = useState<NetworkOperationState>({
    visible: false,
    operationType: null,
    loading: false,
    countdown: 0,
    message: "",
  });

  const showInterfaceRestartOverlay = () => {
    setState({
      visible: true,
      operationType: "interface_restart",
      loading: true,
      countdown: 10,
      message: "Network interface is being restarted. Connection will be restored shortly.",
    });

    // Start countdown for interface restart (10 seconds)
    let timeLeft = 10;
    const countdownInterval = setInterval(() => {
      timeLeft--;
      setState(prev => ({ ...prev, countdown: timeLeft }));
      
      if (timeLeft <= 0) {
        clearInterval(countdownInterval);
        setState(prev => ({ ...prev, loading: false, message: "Interface restart completed. Testing connection..." }));
        
        // Auto-hide after 2 more seconds
        setTimeout(() => {
          setState(prev => ({ ...prev, visible: false }));
        }, 2000);
      }
    }, 1000);
  };

  const showNetworkConfigOverlay = () => {
    setState({
      visible: true,
      operationType: "network_config",
      loading: false,
      countdown: 0,
      message: "Network configuration has been changed. If you changed the IP address, please re-register the device with the new IP.",
    });
  };

  const hide = () => {
    setState({
      visible: false,
      operationType: null,
      loading: false,
      countdown: 0,
      message: "",
    });
  };

  return (
    <NetworkOperationContext.Provider value={{
      state,
      showInterfaceRestartOverlay,
      showNetworkConfigOverlay,
      hide,
    }}>
      {children}
    </NetworkOperationContext.Provider>
  );
};
