// hooks/server/useSecurityManagement.ts
import { useState } from "react";
import AuthService from "../../auth/auth";
import { useAppContext } from "../../context/AppContext";

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

interface SecurityResponse {
  status: string;
  message?: string;
}

// ✅ Standardized error response interface
interface ErrorResponse {
  status: string;
  message: string;
}

export const useSecurityManagement = () => {
  const [uploadingSSH, setUploadingSSH] = useState<boolean>(false);
  const [updatingPassword, setUpdatingPassword] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const { activeDevice } = useAppContext();

  // ✅ Enhanced uploadSSHKey with standardized error handling
  const uploadSSHKey = async (sshKey: string): Promise<boolean> => {
    if (!activeDevice) {
      throw new Error("No active device selected");
    }

    setUploadingSSH(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config1/ssh`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            host: activeDevice.ip,
            key: sshKey.trim(),
          }),
        },
      );

      if (response.ok) {
        const data: SecurityResponse = await response.json();

        // ✅ Check for standardized error response
        if (data.status === "failed") {
          throw new Error(data.message || "SSH key upload failed");
        }

        if (data.status === "success") {
          return true;
        } else {
          throw new Error("Invalid response status from server");
        }
      } else {
        // ✅ Handle HTTP error responses
        const errorData = (await response
          .json()
          .catch(() => ({}))) as ErrorResponse;
        throw new Error(
          errorData.message || `Failed to upload SSH key: ${response.status}`,
        );
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to upload SSH key";
      console.error("Error uploading SSH key:", err);
      setError(errorMessage);

      // ✅ Re-throw error for component to handle notifications
      throw err;
    } finally {
      setUploadingSSH(false);
    }
  };

  // ✅ Enhanced updatePassword with standardized error handling
  const updatePassword = async (
    username: string,
    password: string,
  ): Promise<boolean> => {
    if (!activeDevice) {
      throw new Error("No active device selected");
    }

    setUpdatingPassword(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config1/pass`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            host: activeDevice.ip,
            username: username.trim(),
            password: password,
          }),
        },
      );

      if (response.ok) {
        const data: SecurityResponse = await response.json();

        // ✅ Check for standardized error response
        if (data.status === "failed") {
          throw new Error(data.message || "Password update failed");
        }

        if (data.status === "success") {
          return true;
        } else {
          throw new Error("Invalid response status from server");
        }
      } else {
        // ✅ Handle HTTP error responses
        const errorData = (await response
          .json()
          .catch(() => ({}))) as ErrorResponse;
        throw new Error(
          errorData.message || `Failed to update password: ${response.status}`,
        );
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to update password";
      console.error("Error updating password:", err);
      setError(errorMessage);

      // ✅ Re-throw error for component to handle notifications
      throw err;
    } finally {
      setUpdatingPassword(false);
    }
  };

  return {
    uploadingSSH,
    updatingPassword,
    error,
    uploadSSHKey,
    updatePassword,
  };
};
