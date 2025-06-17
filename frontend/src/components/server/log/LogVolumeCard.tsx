// components/server/logs/LogVolumeCard.tsx
import React from 'react';
import { FaChartLine, FaArrowUp } from 'react-icons/fa';
import { useLogContext } from '../../../context/LogContext';
import './LogVolumeCard.css';

const LogVolumeCard: React.FC = () => {
  const { getLogStats } = useLogContext();
  const stats = getLogStats();

  return (
    <div className="server-logs-volume-card">
      <div className="server-logs-volume-header">
        <div className="server-logs-volume-icon-wrapper">
          <FaChartLine className="server-logs-volume-icon" />
        </div>
        <div className="server-logs-volume-trend">
          <FaArrowUp className="server-logs-trend-icon" />
          <span>Total</span>
        </div>
      </div>
      
      <div className="server-logs-volume-content">
        <h3 className="server-logs-volume-title">Log Volume</h3>
        <div className="server-logs-volume-count">{stats.totalLogs.toLocaleString()}</div>
        <p className="server-logs-volume-subtitle">entries loaded</p>
      </div>
      
      <div className="server-logs-volume-chart">
        <div className="server-logs-chart-bars">
          <div className="server-logs-chart-bar" style={{height: '40%'}}></div>
          <div className="server-logs-chart-bar" style={{height: '60%'}}></div>
          <div className="server-logs-chart-bar" style={{height: '80%'}}></div>
          <div className="server-logs-chart-bar" style={{height: '45%'}}></div>
          <div className="server-logs-chart-bar" style={{height: '90%'}}></div>
          <div className="server-logs-chart-bar" style={{height: '70%'}}></div>
          <div className="server-logs-chart-bar" style={{height: '100%'}}></div>
        </div>
      </div>
    </div>
  );
};

export default LogVolumeCard;
