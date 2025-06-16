// hooks/useServerManagement.ts
import { useState, useEffect } from 'react';
import AuthService from '../../auth/auth';

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

interface ServerDevice {
  id: string;
  ip: string;
  tag: string;
  os: string;
  created_at: string;
  access_token?: string;
}

interface CreateServerRequest {
  ip: string;
  tag: string;
  os: string;
}

interface CreateServerResponse {
  access_token: string;
  created_by: string;
  device: ServerDevice;
  message: string;
  status: string;
}

interface DeleteServerResponse {
  deleted_by: string;
  deleted_ip: string;
  message: string;
  status: string;
}

export const useServerManagement = () => {
  const [servers, setServers] = useState<ServerDevice[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [creating, setCreating] = useState<boolean>(false);
  const [deleting, setDeleting] = useState<string | null>(null);

  const fetchServers = async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config1/device`,
        {
          method: 'GET'
        }
      );

      if (response.ok) {
        const data = await response.json();
        if (data.status === 'success' && data.devices) {
          setServers(data.devices);
        } else {
          setServers([]);
        }
      } else {
        throw new Error(`Failed to fetch servers: ${response.status}`);
      }
    } catch (err) {
      console.error('Error fetching servers:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch servers');
    } finally {
      setLoading(false);
    }
  };

  const createServer = async (serverData: CreateServerRequest): Promise<CreateServerResponse | boolean> => {
    setCreating(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config1/create`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(serverData)
        }
      );

      if (response.ok) {
        const data: CreateServerResponse = await response.json();
        if (data.status === 'success') {
          // Refresh the server list
          await fetchServers();
          return data; // Return the full response including access_token
        } else {
          throw new Error('Server creation failed: Invalid response status');
        }
      } else {
        throw new Error(`Failed to create server: ${response.status}`);
      }
    } catch (err) {
      console.error('Error creating server:', err);
      setError(err instanceof Error ? err.message : 'Failed to create server');
      return false;
    } finally {
      setCreating(false);
    }
  };

  const deleteServer = async (ip: string): Promise<boolean> => {
    setDeleting(ip);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config1/delete`,
        {
          method: 'DELETE',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ ip })
        }
      );

      if (response.ok) {
        const data: DeleteServerResponse = await response.json();
        if (data.status === 'success') {
          // Refresh the server list
          await fetchServers();
          return true;
        } else {
          throw new Error('Server deletion failed: Invalid response status');
        }
      } else {
        throw new Error(`Failed to delete server: ${response.status}`);
      }
    } catch (err) {
      console.error('Error deleting server:', err);
      setError(err instanceof Error ? err.message : 'Failed to delete server');
      return false;
    } finally {
      setDeleting(null);
    }
  };

  useEffect(() => {
    fetchServers();
  }, []);

  return {
    servers,
    loading,
    error,
    creating,
    deleting,
    fetchServers,
    createServer,
    deleteServer
  };
};
