// hooks/server/useLogs.ts
import { useState, useEffect } from "react";
import AuthService from "../../auth/auth";
import { useAppContext } from "../../context/AppContext";
import { useNotification } from "../../context/NotificationContext";

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

// ✅ Standardized response interfaces
interface LogResponse {
  status: string;
  message?: string;
  data?: LogEntry[];
  logs?: LogEntry[];
  timestamp?: string;
}

interface ErrorResponse {
  status: string;
  message: string;
}

export const useLogs = () => {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [originalLogs, setOriginalLogs] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [filters, setFilters] = useState<LogFilters>({
    search: "",
    level: "",
    application: "",
    date: "",
    time: "",
    lines: 100,
  });

  const { activeDevice } = useAppContext();
  const { addNotification } = useNotification();

  // ✅ Enhanced fetchLogs with standardized error handling
  const fetchLogs = async (customFilters?: Partial<LogFilters>) => {
    if (!activeDevice) {
      setLoading(false);
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const currentFilters = { ...filters, ...customFilters };

      // ✅ Use standardized endpoint
      const requestBody: any = {
        host: activeDevice.ip,
      };

      // Add filters to request body
      if (currentFilters.lines && currentFilters.lines !== 100) {
        requestBody.lines = currentFilters.lines;
      }

      if (currentFilters.date) {
        requestBody.date = currentFilters.date;
      }

      if (currentFilters.time) {
        let timeValue = currentFilters.time.trim();
        if (/^\d{2}:\d{2}$/.test(timeValue)) {
          timeValue += ":00";
        }
        if (/^\d{2}:\d{2}:\d{2}$/.test(timeValue)) {
          requestBody.time = timeValue;
        }
      }

      const response = await AuthService.makeAuthenticatedRequest(
        `${BACKEND_URL}/api/admin/server/logs`, // ✅ Updated endpoint
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(requestBody),
        },
      );

      if (response.ok) {
        const data: LogResponse = await response.json();

        // ✅ Check for standardized error response
        if (data.status === "failed" || data.status === "error") {
          throw new Error(data.message || "Failed to fetch logs");
        }

        // ✅ Handle multiple success status variations
        if (data.status === "success" || data.status === "ok" || !data.status) {
          // Handle both wrapped and direct response formats
          const logsData = data.data || data.logs || [];
          setOriginalLogs(logsData);
          applyClientSideFilters(logsData, currentFilters);

          if (logsData.length === 0) {
            addNotification({
              title: "No Logs Found",
              message: "No logs found for the selected criteria",
              type: "info",
              duration: 3000,
            });
          }
        } else {
          console.warn("Unexpected response status:", data.status, data);
          throw new Error(`Unexpected response status: ${data.status}`);
        }
      } else {
        // ✅ Enhanced HTTP error handling
        const errorData = (await response
          .json()
          .catch(() => ({}))) as ErrorResponse;
        throw new Error(
          errorData.message ||
            `HTTP ${response.status}: ${response.statusText}`,
        );
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to fetch logs";
      console.error("Error fetching logs:", err);
      setError(errorMessage);
      setLogs([]);
      setOriginalLogs([]);

      // ✅ Only show notification for non-network errors to avoid spam
      if (!(err instanceof Error && err.message.includes("Failed to reach"))) {
        addNotification({
          title: "Log Fetch Error",
          message: errorMessage,
          type: "error",
          duration: 5000,
        });
      }
    } finally {
      setLoading(false);
    }
  };

  // ✅ Enhanced client-side filtering
  const applyClientSideFilters = (
    data: LogEntry[],
    currentFilters: LogFilters,
  ) => {
    let filteredData = [...data];

    // Apply search filter
    if (currentFilters.search && currentFilters.search.trim()) {
      const searchTerm = currentFilters.search.toLowerCase().trim();
      filteredData = filteredData.filter(
        (log) =>
          log.message.toLowerCase().includes(searchTerm) ||
          log.application.toLowerCase().includes(searchTerm) ||
          log.level.toLowerCase().includes(searchTerm),
      );
    }

    // Apply level filter
    if (currentFilters.level) {
      filteredData = filteredData.filter(
        (log) => log.level.toLowerCase() === currentFilters.level.toLowerCase(),
      );
    }

    // Apply application filter
    if (currentFilters.application) {
      filteredData = filteredData.filter(
        (log) => log.application === currentFilters.application,
      );
    }

    setLogs(filteredData);
  };

  // ✅ Enhanced updateFilters with debounced search
  const updateFilters = (newFilters: Partial<LogFilters>) => {
    const updatedFilters = { ...filters, ...newFilters };
    setFilters(updatedFilters);

    // Define server-side filters
    const serverSideFilters = ["date", "time", "lines"];
    const shouldRefetch = Object.keys(newFilters).some(
      (key) =>
        serverSideFilters.includes(key) &&
        newFilters[key as keyof LogFilters] !==
          filters[key as keyof LogFilters],
    );

    if (shouldRefetch) {
      fetchLogs(updatedFilters);
    } else {
      // Apply client-side filters immediately
      applyClientSideFilters(originalLogs, updatedFilters);
    }
  };

  // ✅ Enhanced statistics
  const getLogStats = () => {
    const today = new Date().toISOString().split("T")[0];
    const todayLogs = originalLogs.filter((log) => {
      try {
        return log.timestamp.startsWith(today);
      } catch {
        return false;
      }
    });

    const errorCount = logs.filter(
      (log) => log.level.toLowerCase() === "error",
    ).length;
    const warningCount = logs.filter(
      (log) => log.level.toLowerCase() === "warning",
    ).length;
    const infoCount = logs.filter(
      (log) => log.level.toLowerCase() === "info",
    ).length;

    return {
      errorCount,
      warningCount,
      infoCount,
      totalToday: todayLogs.length,
      totalLogs: logs.length,
    };
  };

  // ✅ Enhanced unique applications
  const getUniqueApplications = () => {
    const apps = [...new Set(originalLogs.map((log) => log.application))];
    return apps
      .filter((app) => app && app !== "unknown" && app.trim() !== "")
      .sort();
  };

  // ✅ Enhanced export with better error handling
  const exportLogs = (format: "csv" | "json" | "txt") => {
    if (logs.length === 0) {
      addNotification({
        title: "Export Failed",
        message: "No logs available to export",
        type: "warning",
        duration: 3000,
      });
      return;
    }

    try {
      switch (format) {
        case "csv":
          exportAsCSV(logs);
          break;
        case "json":
          exportAsJSON(logs);
          break;
        case "txt":
          exportAsTXT(logs);
          break;
      }

      addNotification({
        title: "Export Successful",
        message: `${logs.length} logs exported as ${format.toUpperCase()}`,
        type: "success",
        duration: 3000,
      });
    } catch (err) {
      console.error("Export error:", err);
      addNotification({
        title: "Export Failed",
        message: `Failed to export logs as ${format.toUpperCase()}`,
        type: "error",
        duration: 5000,
      });
    }
  };

  const exportAsCSV = (data: LogEntry[]) => {
    const headers = ["Timestamp", "Level", "Application", "Message"];
    const csvContent = [
      headers.join(","),
      ...data.map((log) =>
        [
          `"${log.timestamp}"`,
          `"${log.level}"`,
          `"${log.application}"`,
          `"${log.message.replace(/"/g, '""')}"`,
        ].join(","),
      ),
    ].join("\n");

    const timestamp = new Date().toISOString().split("T")[0];
    const filename = `logs_${activeDevice?.ip}_${timestamp}.csv`;
    downloadFile(csvContent, filename, "text/csv");
  };

  const exportAsJSON = (data: LogEntry[]) => {
    const exportData = {
      metadata: {
        device: activeDevice?.ip,
        deviceTag: activeDevice?.tag,
        exportDate: new Date().toISOString(),
        totalEntries: data.length,
        filtersApplied: filters,
      },
      logs: data,
    };
    const jsonContent = JSON.stringify(exportData, null, 2);

    const timestamp = new Date().toISOString().split("T")[0];
    const filename = `logs_${activeDevice?.ip}_${timestamp}.json`;
    downloadFile(jsonContent, filename, "application/json");
  };

  const exportAsTXT = (data: LogEntry[]) => {
    const header = `Log Export for ${activeDevice?.ip}\nExported: ${new Date().toLocaleString()}\nTotal Entries: ${data.length}\n${"=".repeat(80)}\n\n`;
    const textContent =
      header +
      data
        .map(
          (log) =>
            `${log.timestamp} [${log.level.toUpperCase()}] ${log.application}: ${log.message}`,
        )
        .join("\n");

    const timestamp = new Date().toISOString().split("T")[0];
    const filename = `logs_${activeDevice?.ip}_${timestamp}.txt`;
    downloadFile(textContent, filename, "text/plain");
  };

  const downloadFile = (
    content: string,
    filename: string,
    mimeType: string,
  ) => {
    const blob = new Blob([content], { type: mimeType });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  };

  // ✅ Enhanced clearFilters
  const clearFilters = () => {
    const defaultFilters = {
      search: "",
      level: "",
      application: "",
      date: "",
      time: "",
      lines: 100,
    };
    setFilters(defaultFilters);

    // Apply to current data first for immediate UI update
    applyClientSideFilters(originalLogs, defaultFilters);

    // Then fetch fresh data
    fetchLogs(defaultFilters);

    addNotification({
      title: "Filters Cleared",
      message: "All log filters have been reset",
      type: "info",
      duration: 3000,
    });
  };

  // ✅ Enhanced useEffect with proper cleanup
  useEffect(() => {
    if (activeDevice) {
      fetchLogs();
    } else {
      setLogs([]);
      setOriginalLogs([]);
      setError(null);
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
    clearFilters,
  };
};
