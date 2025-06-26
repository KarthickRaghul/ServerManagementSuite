// hooks/server/useCommandExecution.ts
import { useState } from "react";
import AuthService from "../../auth/auth";
import { useAppContext } from "../../context/AppContext";

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

// ✅ Updated interface to match backend response format
interface CommandExecutionResult {
  status: string;
  output: string;
}

// ✅ Standardized error response interface
interface ErrorResponse {
  status: string;
  message: string;
}

export const useCommandExecution = () => {
  const [executing, setExecuting] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const { activeDevice } = useAppContext();

  // ✅ Enhanced executeCommand with standardized error handling
  const executeCommand = async (command: string): Promise<string | null> => {
    if (!activeDevice) {
      throw new Error("No active device selected");
    }

    setExecuting(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config1/cmd`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            host: activeDevice.ip,
            command: command.trim(),
          }),
        },
      );

      if (response.ok) {
        const data: CommandExecutionResult = await response.json();

        // ✅ Check for standardized error response
        if (data.status === "failed") {
          throw new Error(data.output || "Command execution failed");
        }

        // ✅ Handle special case: command execution returns both status and output
        if (data.status === "success") {
          return data.output || "";
        } else {
          throw new Error("Invalid response status from server");
        }
      } else {
        // ✅ Handle HTTP error responses
        const errorData = (await response
          .json()
          .catch(() => ({}))) as ErrorResponse;
        throw new Error(
          errorData.message || `Command execution failed: ${response.status}`,
        );
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to execute command";
      console.error("Error executing command:", err);
      setError(errorMessage);

      // ✅ Re-throw error for component to handle notifications
      throw err;
    } finally {
      setExecuting(false);
    }
  };

  return {
    executing,
    error,
    executeCommand,
  };
};
