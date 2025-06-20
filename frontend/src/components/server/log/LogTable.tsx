// components/server/log/LogTable.tsx
import React from 'react';
import { FaClock, FaExclamationTriangle, FaInfoCircle, FaTimesCircle, FaExclamationCircle } from 'react-icons/fa';
import { useLogContext } from '../../../context/LogContext';
import './LogTable.css';

const LogTable: React.FC = () => {
  const { logs, loading, error } = useLogContext();

  const getLevelIcon = (level: string) => {
    switch (level.toLowerCase()) {
      case 'error':
        return <FaTimesCircle className="server-logs-level-icon error" />;
      case 'warning':
        return <FaExclamationTriangle className="server-logs-level-icon warning" />;
      case 'info':
        return <FaInfoCircle className="server-logs-level-icon info" />;
      default:
        return <FaInfoCircle className="server-logs-level-icon debug" />;
    }
  };

  const getLevelClass = (level: string) => {
    switch (level.toLowerCase()) {
      case 'error':
        return 'server-logs-badge error';
      case 'warning':
        return 'server-logs-badge warning';
      case 'info':
        return 'server-logs-badge info';
      default:
        return 'server-logs-badge debug';
    }
  };

  const formatTimestamp = (timestamp: string) => {
    try {
      const date = new Date(timestamp);
      return date.toLocaleString('en-US', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        hour12: false
      });
    } catch {
      return timestamp;
    }
  };

  if (loading) {
    return (
      <div className="server-logs-table-container">
        <div className="server-logs-table-header">
          <h3 className="server-logs-table-title">Recent Log Entries</h3>
          <span className="server-logs-table-count">Loading...</span>
        </div>
        <div className="server-logs-table-loading">
          <div className="server-logs-loading-spinner"></div>
          <p>Fetching logs from server...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="server-logs-table-container">
        <div className="server-logs-table-header">
          <h3 className="server-logs-table-title">Recent Log Entries</h3>
          <span className="server-logs-table-count">Error</span>
        </div>
        <div className="server-logs-table-error">
          <FaExclamationCircle className="server-logs-error-icon" />
          <p>Failed to load logs</p>
          <span className="server-logs-error-message">{error}</span>
        </div>
      </div>
    );
  }

  return (
    <div className="server-logs-table-container">
      <div className="server-logs-table-header">
        <h3 className="server-logs-table-title">Recent Log Entries</h3>
        <span className="server-logs-table-count">{logs.length} entries</span>
      </div>
      
      <div className="server-logs-table-wrapper">
        {logs.length === 0 ? (
          <div className="server-logs-table-empty">
            <FaInfoCircle className="server-logs-empty-icon" />
            <p>No log entries found</p>
            <span className="server-logs-empty-subtitle">Try adjusting your filters or check if the device is online</span>
          </div>
        ) : (
          <table className="server-logs-table">
            <thead>
              <tr>
                <th>Timestamp</th>
                <th>Level</th>
                <th>Application</th>
                <th>Message</th>
              </tr>
            </thead>
            <tbody>
              {logs.map((log, index) => (
                <tr key={`${log.timestamp}-${index}`} className="server-logs-table-row">
                  <td className="server-logs-timestamp">
                    <FaClock className="server-logs-timestamp-icon" />
                    {formatTimestamp(log.timestamp)}
                  </td>
                  <td>
                    <span className={getLevelClass(log.level)}>
                      {getLevelIcon(log.level)}
                      {log.level.toUpperCase()}
                    </span>
                  </td>
                  <td className="server-logs-source">{log.application}</td>
                  <td className="server-logs-message" title={log.message}>
                    {log.message}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};

export default LogTable;
