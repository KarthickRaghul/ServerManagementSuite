// components/auth/ConnectionProtectedRoute.tsx
import React from 'react';
import { useLocation } from 'react-router-dom';
import { useConnectionOverlay } from '../../context/ConnectionOverlayContext';
import { useRole } from '../../hooks/auth/useRole';
import './ConnectionProtectedRoute.css';

interface ConnectionProtectedRouteProps {
  children: React.ReactNode;
}

const ConnectionProtectedRoute: React.FC<ConnectionProtectedRouteProps> = ({ children }) => {
  const { isConnected, state } = useConnectionOverlay();
  const { isAdmin } = useRole();
  const location = useLocation();

  // Allow access to config page (/) even if not connected
  const isConfigPage = location.pathname === '/';
  
  // ✅ NEW: Check if overlay was manually closed
  const overlayWasClosed = !state.visible && state.deviceIp && !isConnected;
  
  // ✅ NEW: Allow access if:
  // 1. Connected
  // 2. On config page
  // 3. User manually closed overlay (respecting their choice)
  // 4. Admin (they can manage devices)
  const shouldAllowAccess = isConnected || 
                           isConfigPage || 
                           overlayWasClosed || 
                           isAdmin;

  // If should not allow access, show blocking message
  if (!shouldAllowAccess) {
    return (
      <div className="connection-blocked-container">
        <div className="connection-blocked-content">
          <h2>Connection Required</h2>
          <p>Please establish a connection to a device before accessing this page.</p>
          <p>Go to the Configuration page to manage your devices.</p>
        </div>
      </div>
    );
  }

  return <>{children}</>;
};

export default ConnectionProtectedRoute;
