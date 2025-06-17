// hooks/server/useLogs.ts
import { useState, useEffect } from 'react';
import AuthService from '../../auth/auth';
import { useAppContext } from '../../context/AppContext';
import { useNotification } from '../../context/NotificationContext';

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

export const useLogs = () => {
  const [logs, setLogs] = useState<LogEntry[]>([]);
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

      // Prepare request body with host and optional date/time filters
      interface RequestBody {
        host: string;
        date?: string;
        time?: string;
      }

      const requestBody: RequestBody = {
        host: activeDevice.ip
      };

      // Add date/time filters if provided (backend expects these in request body)
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
        let filteredData = data;

        // Apply client-side filters
        if (currentFilters.search) {
          filteredData = filteredData.filter(log => 
            log.message.toLowerCase().includes(currentFilters.search.toLowerCase()) ||
            log.application.toLowerCase().includes(currentFilters.search.toLowerCase())
          );
        }

        if (currentFilters.level) {
          filteredData = filteredData.filter(log => 
            log.level.toLowerCase() === currentFilters.level.toLowerCase()
          );
        }

        if (currentFilters.application) {
          filteredData = filteredData.filter(log => 
            log.application.toLowerCase().includes(currentFilters.application.toLowerCase())
          );
        }

        setLogs(filteredData);
        
        if (filteredData.length === 0 && data.length > 0) {
          addNotification({
            title: 'No Results',
            message: 'No logs match the current filters',
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

  // Update filters and fetch logs
  const updateFilters = (newFilters: Partial<LogFilters>) => {
    const updatedFilters = { ...filters, ...newFilters };
    setFilters(updatedFilters);
    
    // Only fetch if server-side filters changed (date, time, lines)
    const serverSideFilters = ['date', 'time', 'lines'];
    const shouldRefetch = Object.keys(newFilters).some(key => 
      serverSideFilters.includes(key) && newFilters[key as keyof LogFilters] !== filters[key as keyof LogFilters]
    );
    
    if (shouldRefetch) {
      fetchLogs(updatedFilters);
    } else {
      // Apply client-side filters immediately
      applyClientSideFilters(updatedFilters);
    }
  };

  // Apply client-side filters without refetching
  const applyClientSideFilters = (currentFilters: LogFilters) => {
    // This would need access to the original unfiltered data
    // For now, we'll refetch to keep it simple
    fetchLogs(currentFilters);
  };

  // Get log statistics
  const getLogStats = () => {
    const today = new Date().toISOString().split('T')[0];
    const todayLogs = logs.filter(log => log.timestamp.startsWith(today));
    
    const errorCount = logs.filter(log => log.level.toLowerCase() === 'error').length;
    const warningCount = logs.filter(log => log.level.toLowerCase() === 'warning').length;
    const infoCount = logs.filter(log => log.level.toLowerCase() === 'info').length;
    const totalToday = todayLogs.length;

    return {
      errorCount,
      warningCount,
      infoCount,
      totalToday,
      totalLogs: logs.length
    };
  };

  // Get unique applications for filter dropdown
  const getUniqueApplications = () => {
    const apps = [...new Set(logs.map(log => log.application))];
    return apps.filter(app => app && app !== 'unknown').sort();
  };

  // Export logs functionality
  const exportLogs = (format: 'csv' | 'json' | 'pdf') => {
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
        case 'pdf':
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
    setFilters({
      search: '',
      level: '',
      application: '',
      date: '',
      time: '',
      lines: 100
    });
    fetchLogs({
      search: '',
      level: '',
      application: '',
      date: '',
      time: '',
      lines: 100
    });
  };

  // Auto-fetch logs on mount and device change
  useEffect(() => {
    if (activeDevice) {
      fetchLogs();
    }
  }, [activeDevice]);

  return {
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
};
