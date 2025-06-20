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

  // ✅ Enhanced fetchLogs with proper date/time formatting
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
      
      // ✅ Prepare request body with validated formats
      const requestBody: any = {
        host: activeDevice.ip
      };

      // ✅ Add lines filter
      if (currentFilters.lines && currentFilters.lines !== 100) {
        requestBody.lines = currentFilters.lines;
      }

      // ✅ Validate and format date (YYYY-MM-DD)
      if (currentFilters.date) {
        const dateValue = currentFilters.date.trim();
        if (/^\d{4}-\d{2}-\d{2}$/.test(dateValue)) {
          requestBody.date = dateValue;
          console.log('🔍 [LOG] Date filter applied:', dateValue);
        } else {
          console.warn('⚠️ [LOG] Invalid date format, skipping:', dateValue);
        }
      }

      // ✅ Validate and format time (HH:MM:SS)
      if (currentFilters.time) {
        let timeValue = currentFilters.time.trim();
        
        // Convert HH:MM to HH:MM:SS if needed
        if (/^\d{2}:\d{2}$/.test(timeValue)) {
          timeValue += ':00';
        }
        
        if (/^\d{2}:\d{2}:\d{2}$/.test(timeValue)) {
          requestBody.time = timeValue;
          console.log('🔍 [LOG] Time filter applied:', timeValue);
        } else {
          console.warn('⚠️ [LOG] Invalid time format, skipping:', timeValue);
        }
      }

      console.log('🔍 [LOG] Sending request to backend:', requestBody);

      const response = await AuthService.makeAuthenticatedRequest(`${BACKEND_URL}/api/server/log`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestBody)
      });

      if (response.ok) {
        const data: LogEntry[] = await response.json();
        console.log('✅ [LOG] Received logs from backend:', data.length, 'entries');
        
        setOriginalLogs(data);
        applyClientSideFilters(data, currentFilters);
        
        if (data.length === 0) {
          addNotification({
            title: 'No Logs Found',
            message: 'No logs found for the selected date/time. Try a different date.',
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
      console.error('❌ [LOG] Error fetching logs:', err);
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

  // ✅ Enhanced client-side filtering
  const applyClientSideFilters = (data: LogEntry[], currentFilters: LogFilters) => {
    let filteredData = [...data];

    console.log('🔍 [LOG] Applying client-side filters:', {
      search: currentFilters.search,
      level: currentFilters.level,
      application: currentFilters.application,
      totalEntries: data.length
    });

    // Apply search filter (case-insensitive)
    if (currentFilters.search && currentFilters.search.trim()) {
      const searchTerm = currentFilters.search.toLowerCase().trim();
      filteredData = filteredData.filter(log => 
        log.message.toLowerCase().includes(searchTerm) ||
        log.application.toLowerCase().includes(searchTerm) ||
        log.level.toLowerCase().includes(searchTerm)
      );
      console.log(`🔍 [LOG] Search filter applied: ${filteredData.length} entries match "${searchTerm}"`);
    }

    // Apply level filter
    if (currentFilters.level) {
      const levelFilter = currentFilters.level.toLowerCase();
      filteredData = filteredData.filter(log => 
        log.level.toLowerCase() === levelFilter
      );
      console.log(`🔍 [LOG] Level filter applied: ${filteredData.length} entries match level "${levelFilter}"`);
    }

    // Apply application filter
    if (currentFilters.application) {
      filteredData = filteredData.filter(log => 
        log.application === currentFilters.application
      );
      console.log(`🔍 [LOG] Application filter applied: ${filteredData.length} entries match "${currentFilters.application}"`);
    }

    console.log(`✅ [LOG] Final filtered results: ${filteredData.length} of ${data.length} entries`);
    setLogs(filteredData);
  };

  // ✅ Enhanced updateFilters with better server-side filter detection
  const updateFilters = (newFilters: Partial<LogFilters>) => {
    const updatedFilters = { ...filters, ...newFilters };
    setFilters(updatedFilters);
    
    console.log('🔄 [LOG] Filters updated:', newFilters);
    
    // ✅ Define which filters require server refetch
    const serverSideFilters = ['date', 'time', 'lines'];
    const shouldRefetch = Object.keys(newFilters).some(key => {
      if (!serverSideFilters.includes(key)) return false;
      
      const newValue = newFilters[key as keyof LogFilters];
      const oldValue = filters[key as keyof LogFilters];
      
      // Handle empty string vs undefined/null comparison
      const normalizedNew = newValue === '' ? undefined : newValue;
      const normalizedOld = oldValue === '' ? undefined : oldValue;
      
      return normalizedNew !== normalizedOld;
    });
    
    if (shouldRefetch) {
      console.log('🔄 [LOG] Server-side filter changed, refetching logs...');
      fetchLogs(updatedFilters);
    } else {
      console.log('🔄 [LOG] Client-side filter changed, applying local filters...');
      // Apply client-side filters only
      applyClientSideFilters(originalLogs, updatedFilters);
    }
  };

  // ✅ Enhanced log statistics
  const getLogStats = () => {
    const today = new Date().toISOString().split('T')[0];
    const todayLogs = originalLogs.filter(log => {
      try {
        const logDate = new Date(log.timestamp).toISOString().split('T')[0];
        return logDate === today;
      } catch {
        return false;
      }
    });
    
    // Calculate stats from filtered logs (what user sees)
    const errorCount = logs.filter(log => log.level.toLowerCase() === 'error').length;
    const warningCount = logs.filter(log => log.level.toLowerCase() === 'warning').length;
    const infoCount = logs.filter(log => log.level.toLowerCase() === 'info').length;
    const totalToday = todayLogs.length;

    return {
      errorCount,
      warningCount,
      infoCount,
      totalToday,
      totalLogs: logs.length // Show filtered count
    };
  };

  // ✅ Enhanced unique applications from original logs
  const getUniqueApplications = () => {
    const apps = [...new Set(originalLogs.map(log => log.application))];
    return apps
      .filter(app => app && app !== 'unknown' && app.trim() !== '')
      .sort((a, b) => a.localeCompare(b));
  };

  // ✅ Enhanced export with better metadata
  const exportLogs = (format: 'csv' | 'json' | 'txt') => {
    try {
      const dataToExport = logs; // Export filtered logs
      
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
      console.error('❌ [LOG] Export failed:', err);
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

    const timestamp = new Date().toISOString().split('T')[0];
    const filename = `logs_${activeDevice?.ip}_${timestamp}.csv`;
    downloadFile(csvContent, filename, 'text/csv');
  };

  const exportAsJSON = (data: LogEntry[]) => {
    const exportData = {
      metadata: {
        device: activeDevice?.ip,
        deviceTag: activeDevice?.tag,
        exportDate: new Date().toISOString(),
        totalEntries: data.length,
        originalEntries: originalLogs.length,
        filtersApplied: filters
      },
      logs: data
    };
    const jsonContent = JSON.stringify(exportData, null, 2);
    
    const timestamp = new Date().toISOString().split('T')[0];
    const filename = `logs_${activeDevice?.ip}_${timestamp}.json`;
    downloadFile(jsonContent, filename, 'application/json');
  };

  const exportAsTXT = (data: LogEntry[]) => {
    const timestamp = new Date().toLocaleString();
    const header = [
      `SNSMS Log Export`,
      `Device: ${activeDevice?.tag} (${activeDevice?.ip})`,
      `Exported: ${timestamp}`,
      `Total Entries: ${data.length}`,
      `Original Entries: ${originalLogs.length}`,
      `Filters Applied: ${JSON.stringify(filters)}`,
      '='.repeat(80),
      ''
    ].join('\n');
    
    const textContent = header + data.map(log => 
      `${log.timestamp} [${log.level.toUpperCase().padEnd(7)}] ${log.application.padEnd(15)}: ${log.message}`
    ).join('\n');
    
    const fileTimestamp = new Date().toISOString().split('T')[0];
    const filename = `logs_${activeDevice?.ip}_${fileTimestamp}.txt`;
    downloadFile(textContent, filename, 'text/plain');
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

  // ✅ Enhanced clearFilters
  const clearFilters = () => {
    console.log('🧹 [LOG] Clearing all filters');
    const defaultFilters = {
      search: '',
      level: '',
      application: '',
      date: '',
      time: '',
      lines: 100
    };
    setFilters(defaultFilters);
    
    // Apply filters to current data first for immediate UI update
    applyClientSideFilters(originalLogs, defaultFilters);
    
    // Then fetch fresh data
    fetchLogs(defaultFilters);
  };

  // ✅ Auto-fetch logs on mount and device change
  useEffect(() => {
    if (activeDevice) {
      console.log('🔄 [LOG] Active device changed, fetching logs for:', activeDevice.tag);
      fetchLogs();
    } else {
      console.log('⚠️ [LOG] No active device, clearing logs');
      setLogs([]);
      setOriginalLogs([]);
      setError(null);
    }
  }, [activeDevice]);

  // ✅ Debug logging for filter changes
  useEffect(() => {
    console.log('🔍 [LOG] Current filters:', filters);
    console.log('📊 [LOG] Current logs count:', logs.length, 'of', originalLogs.length);
  }, [filters, logs.length, originalLogs.length]);

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
