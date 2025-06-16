// components/auth/RoleProtectedRoute.tsx
import React from 'react';
import { Navigate } from 'react-router-dom';
import { useRole } from '../../hooks';

interface RoleProtectedRouteProps {
  children: React.ReactNode;
  allowedRoles: ('admin' | 'viewer')[];
  redirectTo?: string;
}

const RoleProtectedRoute: React.FC<RoleProtectedRouteProps> = ({ 
  children, 
  allowedRoles, 
  redirectTo = '/health' 
}) => {
  const { role, loading } = useRole();

  if (loading) {
    return <div>Loading...</div>; // Or your loading component
  }

  if (!role || !allowedRoles.includes(role)) {
    return <Navigate to={redirectTo} replace />;
  }

  return <>{children}</>;
};

export default RoleProtectedRoute;
