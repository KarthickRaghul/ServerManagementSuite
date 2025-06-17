// components/server/logs/SuccessRequestCard.tsx
import React from 'react';
import { FaCheckCircle } from 'react-icons/fa';
import { useLogs } from '../../../hooks/server/useLogs';
import './SuccessRequestCard.css';

const SuccessRequestCard: React.FC = () => {
  const { getLogStats } = useLogs();
  const stats = getLogStats();

  return (
    <div className="server-logs-success-card">
      <div className="server-logs-success-header">
        <div className="server-logs-success-icon-wrapper">
          <FaCheckCircle className="server-logs-success-icon" />
        </div>
      </div>
      
      <div className="server-logs-success-content">
        <h3 className="server-logs-success-title">Info Logs Today</h3>
        <div className="server-logs-success-count">{stats.infoCount}</div>
        <p className="server-logs-success-subtitle">informational entries</p>
      </div>
    </div>
  );
};

export default SuccessRequestCard;
