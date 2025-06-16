// hooks/settings/useUserManagement.ts
import { useState, useEffect } from 'react';
import AuthService from '../../auth/auth';

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

interface User {
  id: string;
  name: string;
  email: string;
  role: 'admin' | 'viewer';
}

interface ListUsersResponse {
  count: number;
  status: string;
  users: User[];
}

interface CreateUserRequest {
  username: string;
  role: 'admin' | 'viewer';
  email: string;
  password: string;
}

interface CreateUserResponse {
  message: string;
  status: string;
  user: User;
}

interface DeleteUserResponse {
  deleted_by: string;
  deleted_user: string;
  message: string;
  status: string;
}

export const useUserManagement = () => {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [creating, setCreating] = useState<boolean>(false);
  const [deleting, setDeleting] = useState<string | null>(null);

  const fetchUsers = async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/settings/listuser`,
        {
          method: 'GET'
        }
      );

      if (response.ok) {
        const data: ListUsersResponse = await response.json();
        if (data.status === 'success') {
          setUsers(data.users);
        } else {
          throw new Error('Failed to fetch users: Invalid response status');
        }
      } else {
        throw new Error(`Failed to fetch users: ${response.status}`);
      }
    } catch (err) {
      console.error('Error fetching users:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch users');
      setUsers([]);
    } finally {
      setLoading(false);
    }
  };

  const createUser = async (userData: CreateUserRequest): Promise<boolean> => {
    setCreating(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/settings/adduser`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(userData)
        }
      );

      if (response.ok) {
        const data: CreateUserResponse = await response.json();
        if (data.status === 'success') {
          // Refresh the user list after successful creation
          await fetchUsers();
          return true;
        } else {
          throw new Error('User creation failed: Invalid response status');
        }
      } else {
        throw new Error(`Failed to create user: ${response.status}`);
      }
    } catch (err) {
      console.error('Error creating user:', err);
      setError(err instanceof Error ? err.message : 'Failed to create user');
      return false;
    } finally {
      setCreating(false);
    }
  };

  const deleteUser = async (username: string): Promise<boolean> => {
    setDeleting(username);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/settings/removeuser`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ username })
        }
      );

      if (response.ok) {
        const data: DeleteUserResponse = await response.json();
        if (data.status === 'success') {
          // Refresh the user list after successful deletion
          await fetchUsers();
          return true;
        } else {
          throw new Error('User deletion failed: Invalid response status');
        }
      } else {
        throw new Error(`Failed to delete user: ${response.status}`);
      }
    } catch (err) {
      console.error('Error deleting user:', err);
      setError(err instanceof Error ? err.message : 'Failed to delete user');
      return false;
    } finally {
      setDeleting(null);
    }
  };

  useEffect(() => {
    fetchUsers();
  }, []);

  return {
    users,
    loading,
    error,
    creating,
    deleting,
    fetchUsers,
    createUser,
    deleteUser
  };
};
