// hooks/useSecurityManagement.ts
import { useState } from 'react';
import AuthService from '../../auth/auth';
import { useAppContext } from '../../context/AppContext';

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;



interface SecurityResponse {
  status: string;
  message?: string;
}

export const useSecurityManagement = () => {
  const [uploadingSSH, setUploadingSSH] = useState<boolean>(false);
  const [updatingPassword, setUpdatingPassword] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const { activeDevice } = useAppContext();

  const uploadSSHKey = async (sshKey: string): Promise<boolean> => {
    if (!activeDevice) {
      throw new Error('No active device selected');
    }

    setUploadingSSH(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config1/ssh`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            host: activeDevice.ip,
            key: sshKey.trim()
          })
        }
      );

      if (response.ok) {
        const data: SecurityResponse = await response.json();
        if (data.status === 'success') {
          return true;
        } else {
          throw new Error('SSH key upload failed: Invalid response status');
        }
      } else {
        throw new Error(`Failed to upload SSH key: ${response.status}`);
      }
    } catch (err) {
      console.error('Error uploading SSH key:', err);
      const errorMessage = err instanceof Error ? err.message : 'Failed to upload SSH key';
      setError(errorMessage);
      throw new Error(errorMessage);
    } finally {
      setUploadingSSH(false);
    }
  };

  const updatePassword = async (username: string, password: string): Promise<boolean> => {
    if (!activeDevice) {
      throw new Error('No active device selected');
    }

    setUpdatingPassword(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config1/pass`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            host: activeDevice.ip,
            username: username.trim(),
            password: password
          })
        }
      );

      if (response.ok) {
        const data: SecurityResponse = await response.json();
        if (data.status === 'success') {
          return true;
        } else {
          throw new Error('Password update failed: Invalid response status');
        }
      } else {
        throw new Error(`Failed to update password: ${response.status}`);
      }
    } catch (err) {
      console.error('Error updating password:', err);
      const errorMessage = err instanceof Error ? err.message : 'Failed to update password';
      setError(errorMessage);
      throw new Error(errorMessage);
    } finally {
      setUpdatingPassword(false);
    }
  };

  return {
    uploadingSSH,
    updatingPassword,
    error,
    uploadSSHKey,
    updatePassword
  };
};
