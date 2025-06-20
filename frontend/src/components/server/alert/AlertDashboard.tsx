// components/server/alert/AlertDashboard.tsx
import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { FaSync, FaExclamationTriangle, FaFilter, FaClock, FaEye, FaCheck, FaTrash } from 'react-icons/fa';
import { useAlerts } from '../../../hooks/server/useAlerts';
import { useNotification } from '../../../context/NotificationContext';
import AlertDetailModal from './AlertDetailModal';
import './AlertDashboard.css';

interface Alert {
  id: number;
  host: string;
  severity: 'warning' | 'critical' | 'info';
  content: string;
  status: 'notseen' | 'seen';
  time: string;
}

const AlertDashboard: React.FC = () => {
  const [selectedFilter, setSelectedFilter] = useState('All Alerts');
  const [showDetailModal, setShowDetailModal] = useState(false);
  const [selectedAlert, setSelectedAlert] = useState<Alert | null>(null);
  const [selectedAlerts, setSelectedAlerts] = useState<number[]>([]);
  const [selectAll, setSelectAll] = useState(false);

  const { 
    alerts, 
    loading, 
    error, 
    markingAsSeen, 
    resolving,
    fetchAlerts,
    markSingleAlertAsSeen,
    resolveAlerts
  } = useAlerts();
  
  const { addNotification } = useNotification();

  // Memoize filtered alerts to prevent unnecessary recalculations
  const filteredAlerts = useMemo(() => {
    switch (selectedFilter) {
      case 'High Priority':
        return alerts.filter(alert => alert.severity === 'critical');
      case 'Medium Priority':
        return alerts.filter(alert => alert.severity === 'warning');
      case 'Low Priority':
        return alerts.filter(alert => alert.severity === 'info');
      case 'Unacknowledged':
        return alerts.filter(alert => alert.status === 'notseen');
      default:
        return alerts;
    }
  }, [alerts, selectedFilter]);

  // Fix the infinite loop by properly managing the select all state
  useEffect(() => {
    if (filteredAlerts.length === 0) {
      if (selectAll || selectedAlerts.length > 0) {
        setSelectAll(false);
        setSelectedAlerts([]);
      }
    } else {
      const allSelected = filteredAlerts.every(alert => selectedAlerts.includes(alert.id));
      if (allSelected !== selectAll) {
        setSelectAll(allSelected);
      }
    }
  }, [filteredAlerts.length, selectedAlerts.length, selectAll]); // Fixed dependencies

  const handleRefreshData = useCallback(async () => {
    await fetchAlerts();
    setSelectedAlerts([]);
    setSelectAll(false);
    addNotification({
      title: 'Data Refreshed',
      message: 'Alert data has been refreshed successfully',
      type: 'info',
      duration: 3000
    });
  }, [fetchAlerts, addNotification]);

  const handleViewDetails = useCallback((alert: Alert) => {
    setSelectedAlert(alert);
    setShowDetailModal(true);
  }, []);

  const handleSelectAlert = useCallback((alertId: number) => {
    setSelectedAlerts(prev => {
      if (prev.includes(alertId)) {
        return prev.filter(id => id !== alertId);
      } else {
        return [...prev, alertId];
      }
    });
  }, []);

  const handleSelectAll = useCallback(() => {
    if (selectAll) {
      setSelectedAlerts([]);
      setSelectAll(false);
    } else {
      setSelectedAlerts(filteredAlerts.map(alert => alert.id));
      setSelectAll(true);
    }
  }, [selectAll, filteredAlerts]);

  const handleBulkAcknowledge = useCallback(async () => {
    if (selectedAlerts.length === 0) {
      addNotification({
        title: 'No Selection',
        message: 'Please select alerts to acknowledge',
        type: 'warning',
        duration: 3000
      });
      return;
    }

    try {
      const promises = selectedAlerts.map(alertId => markSingleAlertAsSeen(alertId));
      await Promise.all(promises);
      
      addNotification({
        title: 'Alerts Acknowledged',
        message: `${selectedAlerts.length} alert(s) have been acknowledged`,
        type: 'success',
        duration: 3000
      });
      
      setSelectedAlerts([]);
      setSelectAll(false);
    } catch (err) {
      addNotification({
        title: 'Bulk Acknowledge Failed',
        message: err instanceof Error ? err.message : 'Failed to acknowledge selected alerts',
        type: 'error',
        duration: 5000
      });
    }
  }, [selectedAlerts, markSingleAlertAsSeen, addNotification]);

  const handleBulkResolve = useCallback(async () => {
    if (selectedAlerts.length === 0) {
      addNotification({
        title: 'No Selection',
        message: 'Please select alerts to resolve',
        type: 'warning',
        duration: 3000
      });
      return;
    }

    if (window.confirm(`Are you sure you want to resolve ${selectedAlerts.length} selected alert(s)? This action cannot be undone.`)) {
      try {
        const success = await resolveAlerts(selectedAlerts);
        if (success) {
          addNotification({
            title: 'Alerts Resolved',
            message: `${selectedAlerts.length} alert(s) have been resolved and deleted`,
            type: 'success',
            duration: 3000
          });
          
          setSelectedAlerts([]);
          setSelectAll(false);
        }
      } catch (err) {
        addNotification({
          title: 'Bulk Resolve Failed',
          message: err instanceof Error ? err.message : 'Failed to resolve selected alerts',
          type: 'error',
          duration: 5000
        });
      }
    }
  }, [selectedAlerts, resolveAlerts, addNotification]);

  const handleSingleAcknowledge = useCallback(async (alertId: number) => {
    try {
      const success = await markSingleAlertAsSeen(alertId);
      if (success) {
        addNotification({
          title: 'Alert Acknowledged',
          message: 'Alert has been marked as seen',
          type: 'success',
          duration: 3000
        });
      }
    } catch (err) {
      addNotification({
        title: 'Acknowledge Failed',
        message: err instanceof Error ? err.message : 'Failed to acknowledge alert',
        type: 'error',
        duration: 5000
      });
    }
  }, [markSingleAlertAsSeen, addNotification]);

  const handleSingleResolve = useCallback(async (alertId: number) => {
    if (window.confirm('Are you sure you want to resolve this alert? This action cannot be undone.')) {
      try {
        const success = await resolveAlerts([alertId]);
        if (success) {
          addNotification({
            title: 'Alert Resolved',
            message: 'Alert has been resolved and deleted',
            type: 'success',
            duration: 3000
          });
        }
      } catch (err) {
        addNotification({
          title: 'Resolve Failed',
          message: err instanceof Error ? err.message : 'Failed to resolve alert',
          type: 'error',
          duration: 5000
        });
      }
    }
  }, [resolveAlerts, addNotification]);

  const getAlertStats = useMemo(() => {
    const total = alerts.length;
    const unacknowledged = alerts.filter(alert => alert.status === 'notseen').length;
    const critical = alerts.filter(alert => alert.severity === 'critical').length;
    const warning = alerts.filter(alert => alert.severity === 'warning').length;
    const info = alerts.filter(alert => alert.severity === 'info').length;

    return {
      total,
      unacknowledged,
      acknowledged: total - unacknowledged,
      critical,
      warning,
      info
    };
  }, [alerts]);

  const getSeverityIcon = useCallback((severity: string) => {
    const iconMap = {
      critical: '🔴',
      warning: '🟡',
      info: '🔵'
    };
    return iconMap[severity as keyof typeof iconMap] || '⚠️';
  }, []);

  const getSeverityClass = useCallback((severity: string) => {
    return `monitoring-alerts-severity-${severity}`;
  }, []);

  const formatTime = useCallback((timeInput: string | { Time?: string; time?: string } | null) => {
    try {
      let timeString: string;
      
      if (typeof timeInput === 'object' && timeInput !== null) {
        if (timeInput.Time) {
          timeString = timeInput.Time;
        } else if (timeInput.time) {
          timeString = timeInput.time;
        } else {
          return 'Invalid time';
        }
      } else if (typeof timeInput === 'string') {
        timeString = timeInput;
      } else {
        return 'Invalid time';
      }
  
      const date = new Date(timeString);
      
      if (isNaN(date.getTime())) {
        return 'Invalid time';
      }
      
      const now = new Date();
      const diffInMilliseconds = now.getTime() - date.getTime();
      const diffInMinutes = Math.floor(Math.abs(diffInMilliseconds) / (1000 * 60));
      
      if (diffInMilliseconds < 0) {
        if (diffInMinutes < 1) return 'In a moment';
        if (diffInMinutes < 60) return `In ${diffInMinutes} minutes`;
        if (diffInMinutes < 1440) return `In ${Math.floor(diffInMinutes / 60)} hours`;
        return `In ${Math.floor(diffInMinutes / 1440)} days`;
      }
      
      if (diffInMinutes < 1) return 'Just now';
      if (diffInMinutes < 60) return `${diffInMinutes} minutes ago`;
      if (diffInMinutes < 1440) return `${Math.floor(diffInMinutes / 60)} hours ago`;
      return `${Math.floor(diffInMinutes / 1440)} days ago`;
      
    } catch (error) {
      console.error('Error formatting time:', error, 'Input:', timeInput);
      return 'Unknown time';
    }
  }, []);

  const stats = getAlertStats;

  return (
    <div className="monitoring-alerts-dashboard">
      {/* Header Section */}
      <div className="monitoring-alerts-header-section">
        <div className="monitoring-alerts-title-section">
          <h1 className="monitoring-alerts-page-title">Monitoring Dashboard</h1>
          <p className="monitoring-alerts-page-subtitle">Real-time system monitoring and alert management</p>
        </div>
        <div className="monitoring-alerts-header-actions">
          <button 
            className="monitoring-alerts-btn monitoring-alerts-btn-secondary"
            onClick={handleRefreshData}
            disabled={loading}
          >
            <FaSync className={`monitoring-alerts-btn-icon ${loading ? 'spinning' : ''}`} />
            Refresh Data
          </button>
        </div>
      </div>

      {/* Error Banner */}
      {error && (
        <div className="monitoring-alerts-error-banner">
          <p>Error: {error}</p>
        </div>
      )}

      {/* Stats Section */}
      <div className="monitoring-alerts-stats-container">
        {/* Issues Stats */}
        <div className="monitoring-alerts-stats-card">
          <div className="monitoring-alerts-stats-header">
            <h3>Alert Status</h3>
          </div>
          <div className="monitoring-alerts-stats-grid">
            <div className="monitoring-alerts-stat-item">
              <span className="monitoring-alerts-stat-value">{stats.total}</span>
              <span className="monitoring-alerts-stat-label">Total</span>
            </div>
            <div className="monitoring-alerts-stat-item monitoring-alerts-stat-new">
              <span className="monitoring-alerts-stat-value">{stats.unacknowledged}</span>
              <span className="monitoring-alerts-stat-label">Unacknowledged</span>
            </div>
            <div className="monitoring-alerts-stat-item monitoring-alerts-stat-resolved">
              <span className="monitoring-alerts-stat-value">{stats.acknowledged}</span>
              <span className="monitoring-alerts-stat-label">Acknowledged</span>
            </div>
            <div className="monitoring-alerts-stat-item">
              <span className="monitoring-alerts-stat-value">{stats.total > 0 ? Math.round((stats.acknowledged / stats.total) * 100) : 0}%</span>
              <span className="monitoring-alerts-stat-label">Completion</span>
            </div>
          </div>
          <div className="monitoring-alerts-progress-bar">
            <div 
              className="monitoring-alerts-progress-fill" 
              style={{ width: `${stats.total > 0 ? (stats.acknowledged / stats.total) * 100 : 0}%` }}
            ></div>
          </div>
        </div>

        {/* Alerts by Severity */}
        <div className="monitoring-alerts-stats-card">
          <div className="monitoring-alerts-stats-header">
            <h3>Alerts by Severity</h3>
          </div>
          <div className="monitoring-alerts-donut-container">
            <div className="monitoring-alerts-donut-chart">
              <div className="monitoring-alerts-donut-center">
                <span className="monitoring-alerts-donut-total">{stats.total}</span>
                <span className="monitoring-alerts-donut-label">Total</span>
              </div>
            </div>
            <div className="monitoring-alerts-donut-stats">
              <div className="monitoring-alerts-donut-stat monitoring-alerts-high">
                <span className="monitoring-alerts-donut-value">{stats.critical}</span>
                <span className="monitoring-alerts-donut-stat-label">Critical</span>
              </div>
              <div className="monitoring-alerts-donut-stat monitoring-alerts-medium">
                <span className="monitoring-alerts-donut-value">{stats.warning}</span>
                <span className="monitoring-alerts-donut-stat-label">Warning</span>
              </div>
              <div className="monitoring-alerts-donut-stat monitoring-alerts-low">
                <span className="monitoring-alerts-donut-value">{stats.info}</span>
                <span className="monitoring-alerts-donut-stat-label">Info</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Tabs Section */}
      <div className="monitoring-alerts-tabs-container">
        <div className="monitoring-alerts-tabs">
          <button className={`monitoring-alerts-tab active`}>
            <FaExclamationTriangle className="monitoring-alerts-tab-icon" />
            Alerts ({filteredAlerts.length})
          </button>
        </div>
        <div className="monitoring-alerts-filter-section">
          <select 
            className="monitoring-alerts-filter-select"
            value={selectedFilter}
            onChange={(e) => setSelectedFilter(e.target.value)}
          >
            <option value="All Alerts">All Alerts</option>
            <option value="High Priority">Critical</option>
            <option value="Medium Priority">Warning</option>
            <option value="Low Priority">Info</option>
            <option value="Unacknowledged">Unacknowledged</option>
          </select>
          <FaFilter className="monitoring-alerts-filter-icon" />
        </div>
      </div>

      {/* Bulk Actions */}
      {selectedAlerts.length > 0 && (
        <div className="monitoring-alerts-bulk-actions">
          <div className="monitoring-alerts-bulk-info">
            <span>{selectedAlerts.length} alert(s) selected</span>
          </div>
          <div className="monitoring-alerts-bulk-buttons">
            <button 
              className="monitoring-alerts-bulk-btn monitoring-alerts-bulk-acknowledge"
              onClick={handleBulkAcknowledge}
              disabled={markingAsSeen.length > 0}
            >
              <FaEye className="monitoring-alerts-bulk-btn-icon" />
              Acknowledge Selected
            </button>
            <button 
              className="monitoring-alerts-bulk-btn monitoring-alerts-bulk-resolve"
              onClick={handleBulkResolve}
              disabled={resolving.length > 0}
            >
              <FaCheck className="monitoring-alerts-bulk-btn-icon" />
              Resolve Selected
            </button>
          </div>
        </div>
      )}

      {/* Active Alerts Section */}
      <div className="monitoring-alerts-content-container">
        <div className="monitoring-alerts-section-header">
          <FaExclamationTriangle className="monitoring-alerts-section-icon" />
          <h3>Active Alerts</h3>
        </div>

        {loading ? (
          <div className="monitoring-alerts-loading">Loading alerts...</div>
        ) : (
          <div className="monitoring-alerts-table-container">
            {filteredAlerts.length === 0 ? (
              <div className="monitoring-alerts-empty">
                <div className="monitoring-alerts-empty-icon">✅</div>
                <p>No alerts found. Your system is running smoothly!</p>
              </div>
            ) : (
              <table className="monitoring-alerts-table">
                <thead>
                  <tr>
                    <th className="monitoring-alerts-checkbox-column">
                      <input
                        type="checkbox"
                        className="monitoring-alerts-checkbox"
                        checked={selectAll}
                        onChange={handleSelectAll}
                        disabled={filteredAlerts.length === 0}
                      />
                    </th>
                    <th>Severity</th>
                    <th>Message</th>
                    <th>Host</th>
                    <th>Status</th>
                    <th>Time</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredAlerts.map((alert) => (
                    <tr key={alert.id} className="monitoring-alerts-table-row">
                      <td className="monitoring-alerts-checkbox-column">
                        <input
                          type="checkbox"
                          className="monitoring-alerts-checkbox"
                          checked={selectedAlerts.includes(alert.id)}
                          onChange={() => handleSelectAlert(alert.id)}
                        />
                      </td>
                      <td>
                        <span className={`monitoring-alerts-severity-badge ${getSeverityClass(alert.severity)}`}>
                          {alert.severity}
                        </span>
                      </td>
                      <td>
                        <div className="monitoring-alerts-message">
                          <span className="monitoring-alerts-icon">{getSeverityIcon(alert.severity)}</span>
                          <span className="monitoring-alerts-text">{alert.content}</span>
                        </div>
                      </td>
                      <td>{alert.host}</td>
                      <td>
                        <span className={`monitoring-alerts-status-badge ${alert.status}`}>
                          {alert.status === 'seen' ? 'Acknowledged' : 'New'}
                        </span>
                      </td>
                      <td>
                        <div className="monitoring-alerts-time">
                          <FaClock className="monitoring-alerts-time-icon" />
                          <span>{formatTime(alert.time)}</span>
                        </div>
                      </td>
                      <td>
                        <div className="monitoring-alerts-action-buttons">
                          <button 
                            className="monitoring-alerts-action-btn monitoring-alerts-action-details"
                            onClick={() => handleViewDetails(alert)}
                            title="View Details"
                          >
                            <FaEye />
                          </button>
                          {alert.status === 'notseen' && (
                            <button 
                              className="monitoring-alerts-action-btn monitoring-alerts-action-acknowledge"
                              onClick={() => handleSingleAcknowledge(alert.id)}
                              disabled={markingAsSeen.includes(alert.id)}
                              title="Acknowledge"
                            >
                              <FaCheck />
                            </button>
                          )}
                          <button 
                            className="monitoring-alerts-action-btn monitoring-alerts-action-resolve"
                            onClick={() => handleSingleResolve(alert.id)}
                            disabled={resolving.includes(alert.id)}
                            title="Resolve"
                          >
                            <FaTrash />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        )}
      </div>

      {/* Alert Detail Modal */}
      <AlertDetailModal
        isOpen={showDetailModal}
        onClose={() => setShowDetailModal(false)}
        alert={selectedAlert}
        onAcknowledge={handleSingleAcknowledge}
        onResolve={handleSingleResolve}
        isAcknowledging={markingAsSeen.includes(selectedAlert?.id || 0)}
        isResolving={resolving.includes(selectedAlert?.id || 0)}
      />
    </div>
  );
};

export default AlertDashboard;
