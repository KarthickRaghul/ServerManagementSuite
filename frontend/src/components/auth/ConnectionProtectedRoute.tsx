// components/auth/ConnectionProtectedRoute.tsx
import React from 'react';
import { useLocation } from 'react-router-dom';
import { useConnectionOverlay } from '../../context/ConnectionOverlayContext';
import './ConnectionProtectedRoute.css'; // Import your CSS for styling

interface ConnectionProtectedRouteProps {
  children: React.ReactNode;
}

const ConnectionProtectedRoute: React.FC<ConnectionProtectedRouteProps> = ({ children }) => {
  const { isConnected } = useConnectionOverlay();
  const location = useLocation();

  // Allow access to config page (/) even if not connected
  const isConfigPage = location.pathname === '/';
  
  // If not connected and not on config page, show a blocking message
  if (!isConnected && !isConfigPage) {
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
