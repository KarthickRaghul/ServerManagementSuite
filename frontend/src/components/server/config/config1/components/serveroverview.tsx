// components/server/config/config1/components/serveroverview.tsx
import React from "react";
import "./serveroverview.css";
import { useServerOverview } from "../../../../../hooks";

const ServerOverview: React.FC = () => {
  const { data, loading, error, refresh } = useServerOverview();

  const getStatusClass = (status: string): string => {
    switch (status?.toLowerCase()) {
      case "online":
        return "online";
      case "offline":
        return "offline";
      case "warning":
        return "warning";
      default:
        return "unknown";
    }
  };

  const getStatusDisplay = (status: string): string => {
    return status
      ? status.charAt(0).toUpperCase() + status.slice(1)
      : "Unknown";
  };

  // ✅ Enhanced loading state
  if (loading && !data) {
    return (
      <div className="server-overview">
        <div className="overview-header">
          <h2 className="overview-title">Server Overview</h2>
        </div>
        <div className="overview-metrics">
          <div className="metric-block">
            <div className="metric-value loading">Loading...</div>
            <div className="metric-label">Server Status</div>
          </div>
          <div className="metric-block">
            <div className="metric-value loading">Loading...</div>
            <div className="metric-label">Uptime</div>
          </div>
        </div>
      </div>
    );
  }

  // ✅ Enhanced error state with retry functionality
  if (error) {
    return (
      <div className="server-overview">
        <div className="overview-header">
          <h2 className="overview-title">Server Overview</h2>
        </div>
        <div className="overview-error">
          <p>Failed to load server overview</p>
          <p className="error-details">{error}</p>
          <button className="retry-button" onClick={refresh} disabled={loading}>
            {loading ? "Retrying..." : "Retry"}
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="server-overview">
      <div className="overview-header">
        <h2 className="overview-title">Server Overview</h2>
      </div>

      <div className="overview-metrics">
        <div className="metric-block">
          <div
            className={`metric-value ${data ? getStatusClass(data.status) : "unknown"}`}
          >
            {data ? getStatusDisplay(data.status) : "Unknown"}
          </div>
          <div className="metric-label">Server Status</div>
        </div>

        <div className="metric-block">
          <div className="metric-value uptime">
            {data ? data.uptime : "Unknown"}
          </div>
          <div className="metric-label">Uptime</div>
        </div>
      </div>

      {/* ✅ Enhanced updating indicator */}
      {loading && data && (
        <div className="updating-indicator">
          <small>Updating...</small>
        </div>
      )}
    </div>
  );
};

export default ServerOverview;
