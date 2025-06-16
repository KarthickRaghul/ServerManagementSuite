// hooks/useAuth.ts
import { useState, useEffect } from 'react';
import AuthService from '../../auth/auth';

export const useAuth = () => {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const checkAuth = async () => {
      setIsLoading(true);
      setError(null);
      
      try {
        const authorized = await AuthService.authorized();
        setIsAuthenticated(authorized);
        
        // If not authorized, the AuthService will handle redirect to login
        if (!authorized) {
          console.log('User not authenticated, redirecting to login...');
        }
      } catch (error) {
        console.error('Auth check failed:', error);
        setIsAuthenticated(false);
        setError('Authentication check failed');
        
        // On error, also redirect to login after a brief delay
        setTimeout(() => {
          window.location.href = '/login';
        }, 1000);
      } finally {
        setIsLoading(false);
      }
    };

    checkAuth();
  }, []);

  // Function to manually refresh authentication (useful for retry scenarios)
  const refreshAuth = async () => {
    setIsLoading(true);
    setError(null);
    
    try {
      const authorized = await AuthService.authorized();
      setIsAuthenticated(authorized);
      return authorized;
    } catch (error) {
      console.error('Auth refresh failed:', error);
      setIsAuthenticated(false);
      setError('Authentication refresh failed');
      return false;
    } finally {
      setIsLoading(false);
    }
  };

  // Function to logout
  const logout = async () => {
    setIsLoading(true);
    try {
      await AuthService.logout();
      setIsAuthenticated(false);
    } catch (error) {
      console.error('Logout failed:', error);
      // Even if logout API fails, clear local state
      setIsAuthenticated(false);
    } finally {
      setIsLoading(false);
    }
  };

  return { 
    isAuthenticated, 
    isLoading, 
    error,
    refreshAuth,
    logout
  };
};


