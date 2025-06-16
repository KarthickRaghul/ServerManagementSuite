// hooks/useCommandExecution.ts
import { useState } from 'react';
import AuthService from '../../auth/auth';
import { useAppContext } from '../../context/AppContext';

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

interface CommandExecutionResult {
  output: string;
}

export const useCommandExecution = () => {
  const [executing, setExecuting] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const { activeDevice } = useAppContext();

  const executeCommand = async (command: string): Promise<string | null> => {
    if (!activeDevice) {
      throw new Error('No active device selected');
    }

    setExecuting(true);
    setError(null);

    try {
      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/config1/cmd`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            host: activeDevice.ip,
            command: command.trim()
          })
        }
      );

      if (response.ok) {
        const data: CommandExecutionResult = await response.json();
        return data.output || '';
      } else {
        throw new Error(`Command execution failed: ${response.status}`);
      }
    } catch (err) {
      console.error('Error executing command:', err);
      const errorMessage = err instanceof Error ? err.message : 'Failed to execute command';
      setError(errorMessage);
      throw new Error(errorMessage);
    } finally {
      setExecuting(false);
    }
  };

  return {
    executing,
    error,
    executeCommand
  };
};
