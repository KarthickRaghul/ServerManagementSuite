// components/server/logs/ErrorSummaryCard.tsx
import React from 'react';
import { FaExclamationTriangle } from 'react-icons/fa';
import { useLogContext } from '../../../context/LogContext';
import './ErrorSummaryCard.css';

const ErrorSummaryCard: React.FC = () => {
  const { getLogStats } = useLogContext();
  const stats = getLogStats();

  return (
    <div className="server-logs-error-card">
      <div className="server-logs-error-header">
        <div className="server-logs-error-icon-wrapper">
          <FaExclamationTriangle className="server-logs-error-icon" />
        </div>
      </div>
      
      <div className="server-logs-error-content">
        <h3 className="server-logs-error-title">Errors Today</h3>
        <div className="server-logs-error-count">{stats.errorCount}</div>
        <p className="server-logs-error-subtitle">critical issues detected</p>
      </div>
    </div>
  );
};

export default ErrorSummaryCard;
