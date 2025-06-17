// context/LogContext.tsx
import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import AuthService from '../auth/auth';
import { useAppContext } from './AppContext';
import { useNotification } from './NotificationContext';

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

interface LogEntry {
  timestamp: string;
  level: string;
  application: string;
  message: string;
}

interface LogFilters {
  search: string;
  level: string;
  application: string;
  date: string;
  time: string;
  lines: number;
}

interface LogContextType {
  logs: LogEntry[];
  loading: boolean;
  error: string | null;
  filters: LogFilters;
  updateFilters: (newFilters: Partial<LogFilters>) => void;
  fetchLogs: (customFilters?: Partial<LogFilters>) => Promise<void>;
  getLogStats: () => {
    errorCount: number;
    warningCount: number;
    infoCount: number;
    totalToday: number;
    totalLogs: number;
  };
  getUniqueApplications: () => string[];
  exportLogs: (format: 'csv' | 'json' | 'txt') => void;
  clearFilters: () => void;
}

const LogContext = createContext<LogContextType | undefined>(undefined);

export const useLogContext = () => {
  const context = useContext(LogContext);
  if (context === undefined) {
    throw new Error('useLogContext must be used within a LogProvider');
  }
  return context;
};

interface LogProviderProps {
  children: ReactNode;
}

export const LogProvider: React.FC<LogProviderProps> = ({ children }) => {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [originalLogs, setOriginalLogs] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [filters, setFilters] = useState<LogFilters>({
    search: '',
    level: '',
    application: '',
    date: '',
    time: '',
    lines: 100
  });
  
  const { activeDevice } = useAppContext();
  const { addNotification } = useNotification();

  // Fetch logs from backend
  const fetchLogs = async (customFilters?: Partial<LogFilters>) => {
    if (!activeDevice) {
      addNotification({
        title: 'Log Fetch Error',
        message: 'No active device selected',
        type: 'error',
        duration: 5000
      });
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const currentFilters = { ...filters, ...customFilters };
      
      // Build URL with query parameters for lines
      let url = `${BACKEND_URL}/api/server/log`;
      if (currentFilters.lines && currentFilters.lines !== 100) {
        url += `?lines=${currentFilters.lines}`;
      }

      // Prepare request body
      interface RequestBody {
        host: string;
        date?: string;
        time?: string;
      }

      const requestBody: RequestBody = {
        host: activeDevice.ip
      };

      // Add date/time filters if provided
      if (currentFilters.date) {
        requestBody.date = currentFilters.date;
      }
      if (currentFilters.time) {
        requestBody.time = currentFilters.time;
      }

      const response = await AuthService.makeAuthenticatedRequest(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestBody)
      });

      if (response.ok) {
        const data: LogEntry[] = await response.json();
        setOriginalLogs(data);
        applyClientSideFilters(data, currentFilters);
        
        if (data.length === 0) {
          addNotification({
            title: 'No Logs Found',
            message: 'No logs found for the selected criteria',
            type: 'info',
            duration: 3000
          });
        }
      } else {
        const errorText = await response.text();
        throw new Error(`Failed to fetch logs: ${response.status} - ${errorText}`);
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch logs';
      console.error('Error fetching logs:', err);
      setError(errorMessage);
      setLogs([]);
      setOriginalLogs([]);
      addNotification({
        title: 'Log Fetch Error',
        message: errorMessage,
        type: 'error',
        duration: 5000
      });
    } finally {
      setLoading(false);
    }
  };

  // Apply client-side filters
  const applyClientSideFilters = (data: LogEntry[], currentFilters: LogFilters) => {
    let filteredData = [...data];

    // Apply search filter
    if (currentFilters.search) {
      const searchTerm = currentFilters.search.toLowerCase();
      filteredData = filteredData.filter(log => 
        log.message.toLowerCase().includes(searchTerm) ||
        log.application.toLowerCase().includes(searchTerm)
      );
    }

    // Apply level filter
    if (currentFilters.level) {
      filteredData = filteredData.filter(log => 
        log.level.toLowerCase() === currentFilters.level.toLowerCase()
      );
    }

    // Apply application filter
    if (currentFilters.application) {
      filteredData = filteredData.filter(log => 
        log.application.toLowerCase().includes(currentFilters.application.toLowerCase())
      );
    }

    setLogs(filteredData);
  };

  // Update filters
  const updateFilters = (newFilters: Partial<LogFilters>) => {
    const updatedFilters = { ...filters, ...newFilters };
    setFilters(updatedFilters);
    
    // Check if server-side filters changed
    const serverSideFilters = ['date', 'time', 'lines'];
    const shouldRefetch = Object.keys(newFilters).some(key => 
      serverSideFilters.includes(key)
    );
    
    if (shouldRefetch) {
      fetchLogs(updatedFilters);
    } else {
      // Apply client-side filters only
      applyClientSideFilters(originalLogs, updatedFilters);
    }
  };

  // Get log statistics
  const getLogStats = () => {
    const today = new Date().toISOString().split('T')[0];
    const todayLogs = originalLogs.filter(log => log.timestamp.startsWith(today));
    
    const errorCount = originalLogs.filter(log => log.level.toLowerCase() === 'error').length;
    const warningCount = originalLogs.filter(log => log.level.toLowerCase() === 'warning').length;
    const infoCount = originalLogs.filter(log => log.level.toLowerCase() === 'info').length;
    const totalToday = todayLogs.length;

    return {
      errorCount,
      warningCount,
      infoCount,
      totalToday,
      totalLogs: originalLogs.length
    };
  };

  // Get unique applications
  const getUniqueApplications = () => {
    const apps = [...new Set(originalLogs.map(log => log.application))];
    return apps.filter(app => app && app !== 'unknown').sort();
  };

  // Export logs
  const exportLogs = (format: 'csv' | 'json' | 'txt') => {
    try {
      const dataToExport = logs;
      
      if (dataToExport.length === 0) {
        addNotification({
          title: 'Export Failed',
          message: 'No logs to export',
          type: 'warning',
          duration: 3000
        });
        return;
      }
      
      switch (format) {
        case 'csv':
          exportAsCSV(dataToExport);
          break;
        case 'json':
          exportAsJSON(dataToExport);
          break;
        case 'txt':
          exportAsTXT(dataToExport);
          break;
      }
      
      addNotification({
        title: 'Export Successful',
        message: `${dataToExport.length} logs exported as ${format.toUpperCase()}`,
        type: 'success',
        duration: 3000
      });
    } catch (err) {
      addNotification({
        title: 'Export Failed',
        message: `Failed to export logs as ${format.toUpperCase()}`,
        type: 'error',
        duration: 5000
      });
    }
  };

  const exportAsCSV = (data: LogEntry[]) => {
    const headers = ['Timestamp', 'Level', 'Application', 'Message'];
    const csvContent = [
      headers.join(','),
      ...data.map(log => [
        `"${log.timestamp}"`,
        `"${log.level}"`,
        `"${log.application}"`,
        `"${log.message.replace(/"/g, '""')}"`
      ].join(','))
    ].join('\n');

    downloadFile(csvContent, `logs_${activeDevice?.ip}_${new Date().toISOString().split('T')[0]}.csv`, 'text/csv');
  };

  const exportAsJSON = (data: LogEntry[]) => {
    const exportData = {
      device: activeDevice?.ip,
      exportDate: new Date().toISOString(),
      totalEntries: data.length,
      filters: filters,
      logs: data
    };
    const jsonContent = JSON.stringify(exportData, null, 2);
    downloadFile(jsonContent, `logs_${activeDevice?.ip}_${new Date().toISOString().split('T')[0]}.json`, 'application/json');
  };

  const exportAsTXT = (data: LogEntry[]) => {
    const header = `Log Export for ${activeDevice?.ip}\nExported: ${new Date().toLocaleString()}\nTotal Entries: ${data.length}\n${'='.repeat(80)}\n\n`;
    const textContent = header + data.map(log => 
      `${log.timestamp} [${log.level.toUpperCase()}] ${log.application}: ${log.message}`
    ).join('\n');
    
    downloadFile(textContent, `logs_${activeDevice?.ip}_${new Date().toISOString().split('T')[0]}.txt`, 'text/plain');
  };

  const downloadFile = (content: string, filename: string, mimeType: string) => {
    const blob = new Blob([content], { type: mimeType });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  };

  // Clear filters
  const clearFilters = () => {
    const defaultFilters = {
      search: '',
      level: '',
      application: '',
      date: '',
      time: '',
      lines: 100
    };
    setFilters(defaultFilters);
    fetchLogs(defaultFilters);
  };

  // Auto-fetch logs on mount and device change
  useEffect(() => {
    if (activeDevice) {
      fetchLogs();
    }
  }, [activeDevice]);

  const value: LogContextType = {
    logs,
    loading,
    error,
    filters,
    updateFilters,
    fetchLogs,
    getLogStats,
    getUniqueApplications,
    exportLogs,
    clearFilters
  };

  return (
    <LogContext.Provider value={value}>
      {children}
    </LogContext.Provider>
  );
};
