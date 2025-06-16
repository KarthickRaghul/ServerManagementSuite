// components/auth/ProtectedRoute.tsx
import React from 'react';
import { useAuth } from '../../hooks';
import './ProtectedRoute.css';

interface ProtectedRouteProps {
  children: React.ReactNode;
}

const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ children }) => {
  const { isAuthenticated, isLoading, error } = useAuth();

  // Show loading spinner while checking authentication
  if (isLoading) {
    return (
      <div className="auth-loading-container">
        <div className="auth-loading-spinner"></div>
        <p>Checking authentication...</p>
      </div>
    );
  }

  // Show error if authentication check failed
  if (error) {
    return (
      <div className="auth-error-container">
        <div className="auth-error-message">
          <span className="auth-error-icon">⚠️</span>
          <p>{error}</p>
          <p>Redirecting to login...</p>
        </div>
      </div>
    );
  }

  // If not authenticated, show redirecting message
  // (The useAuth hook will handle the actual redirect)
  if (!isAuthenticated) {
    return (
      <div className="auth-loading-container">
        <div className="auth-loading-spinner"></div>
        <p>Redirecting to login...</p>
      </div>
    );
  }

  // If authenticated, render the protected content
  return <>{children}</>;
};

export default ProtectedRoute;
