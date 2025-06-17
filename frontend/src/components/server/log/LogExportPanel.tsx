// components/server/logs/LogExportPanel.tsx
import React from 'react';
import { FaDownload, FaFileCsv, FaFileCode, FaFileAlt } from 'react-icons/fa';
import { useLogContext } from '../../../context/LogContext';
import './LogExportPanel.css';

const LogExportPanel: React.FC = () => {
  const { exportLogs, loading, logs } = useLogContext();

  return (
    <div className="server-logs-export-panel">
      <div className="server-logs-export-header">
        <div className="server-logs-export-icon-wrapper">
          <FaDownload className="server-logs-export-icon" />
        </div>
        <h3 className="server-logs-export-title">Export Logs</h3>
      </div>
      
      <div className="server-logs-export-info">
        <span className="server-logs-export-count">{logs.length} entries ready</span>
      </div>
      
      <div className="server-logs-export-buttons">
        <button 
          className="server-logs-export-btn csv"
          onClick={() => exportLogs('csv')}
          disabled={loading || logs.length === 0}
        >
          <FaFileCsv className="server-logs-export-btn-icon" />
          Export as CSV
        </button>
        
        <button 
          className="server-logs-export-btn json"
          onClick={() => exportLogs('json')}
          disabled={loading || logs.length === 0}
        >
          <FaFileCode className="server-logs-export-btn-icon" />
          Export as JSON
        </button>
        
        <button 
          className="server-logs-export-btn txt"
          onClick={() => exportLogs('txt')}
          disabled={loading || logs.length === 0}
        >
          <FaFileAlt className="server-logs-export-btn-icon" />
          Export as TXT
        </button>
      </div>
    </div>
  );
};

export default LogExportPanel;
