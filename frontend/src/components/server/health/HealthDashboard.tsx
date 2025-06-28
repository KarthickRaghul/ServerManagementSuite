// components/server/health/HealthDashboard.tsx
import React, { useState } from "react";
import {
  FaSync,
  FaExpand,
  FaServer,
  FaMemory,
  FaHdd,
  FaNetworkWired,
  FaGlobe,
  FaSpinner,
  FaExclamationTriangle,
} from "react-icons/fa";
import { useHealthMetrics } from "../../../hooks/server/useHealthMetrics";
import { useNotification } from "../../../context/NotificationContext";
import PortsModal from "./PortsModal";
import "./HealthDashboard.css";

interface Port {
  protocol: string;
  port: number;
  process: string;
}

const HealthDashboard: React.FC = () => {
  const [showPortsModal, setShowPortsModal] = useState(false);
  const { healthData, loading, error, refreshMetrics } = useHealthMetrics();
  const { addNotification } = useNotification();

  const handleRefreshData = async () => {
    try {
      await refreshMetrics();
      addNotification({
        title: "Data Refreshed",
        message: "Health metrics have been refreshed successfully",
        type: "success",
        duration: 3000,
      });
    } catch (err) {
      // Error notification is already handled by the hook
      console.error("Failed to refresh health metrics:", err);
    }
  };

  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i];
  };

  const getOpenPorts = (): Port[] => {
    if (!healthData?.open_ports) return [];
    return healthData.open_ports.slice(0, 8);
  };

  const getPortColor = (port: number): string => {
    // Critical security ports - Red
    if ([22, 23, 21, 1433, 3306, 5432, 3389].includes(port)) return "#ef4444";
    // Web services - Blue
    if ([80, 443, 8080, 8443].includes(port)) return "#3b82f6";
    // Email services - Green
    if ([25, 110, 143, 993, 995].includes(port)) return "#22c55e";
    // Database services - Purple
    if ([3306, 5432, 1433, 6379, 27017, 9200].includes(port)) return "#a855f7";
    // DNS and system - Orange
    if ([53, 123].includes(port)) return "#f59e0b";
    // Development ports - Cyan
    if ([3000, 5173, 8000, 8500].includes(port)) return "#06b6d4";
    // Others - Gray
    return "#64748b";
  };

  const handleExpandPorts = () => {
    setShowPortsModal(true);
  };

  // ✅ Enhanced loading state
  if (loading && !healthData) {
    return (
      <div className="health-dashboard">
        <div className="health-dashboard-loading">
          <div className="health-dashboard-loading-spinner">
          </div>
          <p>Loading system health metrics...</p>
        </div>
      </div>
    );
  }

  // ✅ Enhanced error state
  if (error && !healthData) {
    return (
      <div className="health-dashboard">
        <div className="health-dashboard-error">
          <div className="health-dashboard-error-icon">
            <FaExclamationTriangle />
          </div>
          <div className="health-dashboard-error-content">
            <h4>Failed to Load Health Metrics</h4>
            <p>{error}</p>
            <button
              className="health-btn health-btn-primary"
              onClick={handleRefreshData}
              disabled={loading}
            >
              <FaSync className={loading ? "spinning" : ""} />
              Retry
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="health-dashboard">
      {/* Header Section */}
      <div className="health-header-section">
        <div className="health-title-section">
          <h1 className="health-page-title">System Health Dashboard</h1>
          <p className="health-page-subtitle">
            Real-time monitoring of system resources and performance metrics
          </p>
        </div>
        <div className="health-header-actions">
          <button
            className="health-btn health-btn-secondary"
            onClick={handleRefreshData}
            disabled={loading}
            title="Refresh health metrics"
          >
            <FaSync
              className={`health-btn-icon ${loading ? "spinning" : ""}`}
            />
            {loading ? "Refreshing..." : "Refresh Data"}
          </button>
        </div>
      </div>

      {/* Error Banner */}
      {error && healthData && (
        <div className="health-error-banner">
          <FaExclamationTriangle />
          <p>Warning: {error}</p>
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
              <p>Current processor load</p>
            </div>
          </div>
          <div className="health-metric-value">
            <span className="health-value-main">
              {loading ? "--" : `${healthData?.cpu?.usage_percent.toFixed(1)}%`}
            </span>
            <span className="health-value-label">current</span>
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
              <p>
                {healthData
                  ? `${formatBytes(healthData.ram.used_mb * 1024 * 1024)} / ${formatBytes(healthData.ram.total_mb * 1024 * 1024)}`
                  : "Loading..."}
              </p>
            </div>
          </div>
          <div className="health-metric-value">
            <span className="health-value-main">
              {loading ? "--" : `${healthData?.ram?.usage_percent.toFixed(1)}%`}
            </span>
            <span className="health-value-label">used</span>
          </div>
          <div className="health-metric-details">
            <div className="health-detail-item">
              <span className="health-detail-label">Available:</span>
              <span className="health-detail-value">
                {healthData
                  ? formatBytes(healthData.ram.free_mb * 1024 * 1024)
                  : "--"}
              </span>
            </div>
            <div className="health-detail-item">
              <span className="health-detail-label">Used:</span>
              <span className="health-detail-value">
                {healthData
                  ? formatBytes(healthData.ram.used_mb * 1024 * 1024)
                  : "--"}
              </span>
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
              <p>
                {healthData
                  ? `${formatBytes(healthData.disk.used_mb * 1024 * 1024)} / ${formatBytes(healthData.disk.total_mb * 1024 * 1024)}`
                  : "Loading..."}
              </p>
            </div>
          </div>
          <div className="health-metric-value">
            <span className="health-value-main">
              {loading
                ? "--"
                : `${healthData?.disk?.usage_percent.toFixed(1)}%`}
            </span>
            <span className="health-value-label">used</span>
          </div>
          <div className="health-metric-details">
            <div className="health-detail-item">
              <span className="health-detail-label">Available:</span>
              <span className="health-detail-value">
                {healthData
                  ? formatBytes(healthData.disk.free_mb * 1024 * 1024)
                  : "--"}
              </span>
            </div>
            <div className="health-detail-item">
              <span className="health-detail-label">Used:</span>
              <span className="health-detail-value">
                {healthData
                  ? formatBytes(healthData.disk.used_mb * 1024 * 1024)
                  : "--"}
              </span>
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
              <p>
                {healthData
                  ? `${healthData.net.name} - ${(healthData.net.bytes_sent_mb + healthData.net.bytes_recv_mb).toFixed(1)} MB total`
                  : "Real-time data transfer"}
              </p>
            </div>
          </div>
          <div className="health-network-stats">
            <div className="health-network-stat">
              <div className="health-network-indicator transmit"></div>
              <span className="health-network-label">Transmit</span>
              <span className="health-network-value">
                {healthData ? healthData.net.bytes_sent_mb.toFixed(1) : "--"}
              </span>
              <span className="health-network-unit">MB</span>
            </div>
            <div className="health-network-stat">
              <div className="health-network-indicator receive"></div>
              <span className="health-network-label">Receive</span>
              <span className="health-network-value">
                {healthData ? healthData.net.bytes_recv_mb.toFixed(1) : "--"}
              </span>
              <span className="health-network-unit">MB</span>
            </div>
          </div>
        </div>

        {/* ✅ Enhanced Ports Card using process names from API */}
        <div className="health-metric-card health-ports-card">
          <div className="health-metric-header">
            <div className="health-metric-icon health-ports-icon">
              <FaGlobe />
            </div>
            <div className="health-metric-title">
              <h3>Open Ports</h3>
              <p>
                {healthData
                  ? `${healthData.open_ports.length} active services`
                  : "Loading..."}
              </p>
            </div>
            <button
              className="health-expand-btn"
              onClick={handleExpandPorts}
              title="View all ports"
            >
              <FaExpand />
            </button>
          </div>
          <div className="health-ports-list">
            {loading ? (
              <div className="health-ports-loading">Loading ports...</div>
            ) : (
              getOpenPorts().map((port, index) => (
                <div
                  key={index}
                  className="health-port-service-item"
                  style={{ borderLeftColor: getPortColor(port.port) }}
                  title={`${port.process} on port ${port.port}\nProtocol: ${port.protocol.toUpperCase()}`}
                >
                  <div className="health-port-service-main">
                    <span className="health-port-service-name">
                      {port.process}
                    </span>
                    <span className="health-port-service-port">
                      :{port.port}
                    </span>
                  </div>
                  <div className="health-port-service-details">
                    <span className="health-port-service-protocol">
                      {port.protocol.toUpperCase()}
                    </span>
                    <span className="health-port-service-process">
                      {port.process}
                    </span>
                  </div>
                </div>
              ))
            )}
            {healthData && healthData.open_ports.length > 8 && (
              <div className="health-ports-more">
                <span>+{healthData.open_ports.length - 8} more services</span>
                <button onClick={handleExpandPorts}>View All</button>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* ✅ Loading indicator when updating */}
      {loading && healthData && (
        <div className="health-dashboard-updating">
          <FaSpinner className="spinning" />
          <span>Updating health metrics...</span>
        </div>
      )}

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
