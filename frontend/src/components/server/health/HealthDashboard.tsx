// components/common/health/HealthDashboard.tsx
import React, { useState } from 'react';
import { FaSync, FaDownload, FaExpand, FaServer, FaMemory, FaHdd, FaNetworkWired, FaGlobe } from 'react-icons/fa';
import { useHealthMetrics } from '../../../hooks/server/useHealthMetrics';
import { useNotification } from '../../../context/NotificationContext';
import PortsModal from './PortsModal';
import './HealthDashboard.css';

const HealthDashboard: React.FC = () => {
  const [showPortsModal, setShowPortsModal] = useState(false);
  const { healthData, metrics, loading, error, refreshMetrics } = useHealthMetrics();
  const { addNotification } = useNotification();

  const handleRefreshData = async () => {
    await refreshMetrics();
    addNotification({
      title: 'Data Refreshed',
      message: 'Health metrics have been refreshed successfully',
      type: 'info',
      duration: 3000
    });
  };

  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + sizes[i];
  };

  const getOpenPorts = () => {
    if (!healthData?.open_ports) return [];
    return healthData.open_ports.slice(0, 8); // Show first 8 ports
  };

  const getPortColor = (port: number): string => {
    if ([80, 443].includes(port)) return '#ef4444'; // Red for HTTP/HTTPS
    if ([22, 3306, 5432].includes(port)) return '#ef4444'; // Red for SSH, MySQL, PostgreSQL
    return '#64748b'; // Gray for others
  };

  const handleExpandPorts = () => {
    setShowPortsModal(true);
  };

  return (
    <div className="health-dashboard">
      {/* Header Section */}
      <div className="health-header-section">
        <div className="health-title-section">
          <h1 className="health-page-title">System Health Dashboard</h1>
          <p className="health-page-subtitle">Real-time monitoring of system resources and performance metrics</p>
        </div>
        <div className="health-header-actions">
          <button 
            className="health-btn health-btn-secondary"
            onClick={handleRefreshData}
            disabled={loading}
          >
            <FaSync className={`health-btn-icon ${loading ? 'spinning' : ''}`} />
            Refresh Data
          </button>
          <button className="health-btn health-btn-primary">
            <FaDownload className="health-btn-icon" />
            Export Report
          </button>
        </div>
      </div>

      {/* Error Banner */}
      {error && (
        <div className="health-error-banner">
          <p>Error: {error}</p>
        </div>
      )}

      {/* Main Metrics Grid */}
      <div className="health-metrics-grid">
        {/* CPU Usage Card */}
        <div className="health-metric-card health-cpu-card">
          <div className="health-metric-header">
            <div className="health-metric-icon health-cpu-icon">
              <FaServer />
            </div>
            <div className="health-metric-title">
              <h3>CPU Usage</h3>
              <p>Current load</p>
            </div>
          </div>
          <div className="health-metric-value">
            <span className="health-value-main">{loading ? '--' : `${metrics?.cpu?.toFixed(1) || 0}%`}</span>
            <span className="health-value-label">current</span>
          </div>
          <div className="health-metric-details">
            <div className="health-detail-item">
              <span className="health-detail-label">Avg:</span>
              <span className="health-detail-value">{loading ? '--' : `${((metrics?.cpu || 0) * 1.5).toFixed(1)}%`}</span>
            </div>
            <div className="health-detail-item">
              <span className="health-detail-label">Peak:</span>
              <span className="health-detail-value">{loading ? '--' : `${((metrics?.cpu || 0) * 2).toFixed(1)}%`}</span>
            </div>
          </div>
        </div>

        {/* Memory Usage Card */}
        <div className="health-metric-card health-memory-card">
          <div className="health-metric-header">
            <div className="health-metric-icon health-memory-icon">
              <FaMemory />
            </div>
            <div className="health-metric-title">
              <h3>Memory Usage</h3>
              <p>{healthData ? `${formatBytes(healthData.ram.used_mb * 1024 * 1024)} / ${formatBytes(healthData.ram.total_mb * 1024 * 1024)}` : 'Loading...'}</p>
            </div>
          </div>
          <div className="health-metric-value">
            <span className="health-value-main">{loading ? '--' : `${metrics?.ram?.toFixed(1) || 0}%`}</span>
            <span className="health-value-label">used</span>
          </div>
          <div className="health-metric-details">
            <div className="health-detail-item">
              <span className="health-detail-label">Available:</span>
              <span className="health-detail-value">{healthData ? formatBytes(healthData.ram.free_mb * 1024 * 1024) : '--'}</span>
            </div>
            <div className="health-detail-item">
              <span className="health-detail-label">Used:</span>
              <span className="health-detail-value">{healthData ? formatBytes(healthData.ram.used_mb * 1024 * 1024) : '--'}</span>
            </div>
          </div>
        </div>

        {/* Storage Usage Card */}
        <div className="health-metric-card health-storage-card">
          <div className="health-metric-header">
            <div className="health-metric-icon health-storage-icon">
              <FaHdd />
            </div>
            <div className="health-metric-title">
              <h3>Storage Usage</h3>
              <p>{healthData ? `${formatBytes(healthData.disk.used_mb * 1024 * 1024)} / ${formatBytes(healthData.disk.total_mb * 1024 * 1024)}` : 'Loading...'}</p>
            </div>
          </div>
          <div className="health-metric-value">
            <span className="health-value-main">{loading ? '--' : `${metrics?.disk?.toFixed(1) || 0}%`}</span>
            <span className="health-value-label">used</span>
          </div>
          <div className="health-metric-details">
            <div className="health-detail-item">
              <span className="health-detail-label">Available:</span>
              <span className="health-detail-value">{healthData ? formatBytes(healthData.disk.free_mb * 1024 * 1024) : '--'}</span>
            </div>
            <div className="health-detail-item">
              <span className="health-detail-label">Used:</span>
              <span className="health-detail-value">{healthData ? formatBytes(healthData.disk.used_mb * 1024 * 1024) : '--'}</span>
            </div>
          </div>
        </div>

        {/* Network Traffic Card */}
        <div className="health-metric-card health-network-card">
          <div className="health-metric-header">
            <div className="health-metric-icon health-network-icon">
              <FaNetworkWired />
            </div>
            <div className="health-metric-title">
              <h3>Network Traffic</h3>
              <p>Real-time data transfer</p>
            </div>
          </div>
          <div className="health-network-stats">
            <div className="health-network-stat">
              <div className="health-network-indicator transmit"></div>
              <span className="health-network-label">Transmit</span>
              <span className="health-network-value">{healthData ? `${healthData.net.bytes_sent_mb.toFixed(1)}` : '--'}</span>
              <span className="health-network-unit">MB/s</span>
            </div>
            <div className="health-network-stat">
              <div className="health-network-indicator receive"></div>
              <span className="health-network-label">Receive</span>
              <span className="health-network-value">{healthData ? `${healthData.net.bytes_recv_mb.toFixed(1)}` : '--'}</span>
              <span className="health-network-unit">MB/s</span>
            </div>
          </div>
        </div>

        {/* Ports Opened Card */}
        <div className="health-metric-card health-ports-card">
          <div className="health-metric-header">
            <div className="health-metric-icon health-ports-icon">
              <FaGlobe />
            </div>
            <div className="health-metric-title">
              <h3>Ports Opened</h3>
              <p>{healthData ? `${healthData.open_ports.length} active ports` : 'Loading...'}</p>
            </div>
            <button className="health-expand-btn" onClick={handleExpandPorts}>
              <FaExpand />
            </button>
          </div>
          <div className="health-ports-grid">
            {loading ? (
              <div className="health-ports-loading">Loading ports...</div>
            ) : (
              getOpenPorts().map((port, index) => (
                <div 
                  key={index} 
                  className="health-port-item"
                  style={{ backgroundColor: getPortColor(port.port) }}
                >
                  {port.port}
                </div>
              ))
            )}
            {healthData && healthData.open_ports.length > 8 && (
              <div className="health-ports-more">
                +{healthData.open_ports.length - 8} more ports
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Ports Modal */}
      <PortsModal
        isOpen={showPortsModal}
        onClose={() => setShowPortsModal(false)}
        ports={healthData?.open_ports || []}
      />
    </div>
  );
};

export default HealthDashboard;
