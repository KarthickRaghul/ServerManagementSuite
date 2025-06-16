// components/server/alert/AlertDetailModal.tsx
import React from 'react';
import { FaTimes, FaExclamationTriangle, FaCheck, FaEye } from 'react-icons/fa';
import './AlertDetailModal.css';

interface Alert {
  id: number;
  host: string;
  severity: 'warning' | 'critical' | 'info';
  content: string;
  status: 'notseen' | 'seen';
  time: string;
}

interface AlertDetailModalProps {
  isOpen: boolean;
  onClose: () => void;
  alert: Alert | null;
  onAcknowledge: (alertId: number) => Promise<void>;
  onResolve: (alertId: number) => Promise<void>;
  isAcknowledging: boolean;
  isResolving: boolean;
}

const AlertDetailModal: React.FC<AlertDetailModalProps> = ({
  isOpen,
  onClose,
  alert,
  onAcknowledge,
  onResolve,
  isAcknowledging,
  isResolving
}) => {
  if (!isOpen || !alert) return null;

  const getSeverityClass = (severity: string) => {
    switch (severity) {
      case 'critical':
        return 'monitoring-alerts-detail-severity-critical';
      case 'warning':
        return 'monitoring-alerts-detail-severity-warning';
      case 'info':
        return 'monitoring-alerts-detail-severity-info';
      default:
        return 'monitoring-alerts-detail-severity-warning';
    }
  };

  const formatTime = (timeString: string) => {
    try {
      return new Date(timeString).toLocaleString();
    } catch {
      return timeString;
    }
  };

  const handleBackdropClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose();
    }
  };

  return (
    <div className="monitoring-alerts-detail-overlay" onClick={handleBackdropClick}>
      <div className="monitoring-alerts-detail-modal">
        <div className="monitoring-alerts-detail-header">
          <div className="monitoring-alerts-detail-title-section">
            <FaExclamationTriangle className="monitoring-alerts-detail-icon" />
            <h2 className="monitoring-alerts-detail-title">Alert Details</h2>
          </div>
          <button 
            className="monitoring-alerts-detail-close"
            onClick={onClose}
            type="button"
          >
            <FaTimes />
          </button>
        </div>

        <div className="monitoring-alerts-detail-content">
          <div className="monitoring-alerts-detail-info">
            <div className="monitoring-alerts-detail-field">
              <label className="monitoring-alerts-detail-label">Alert ID:</label>
              <span className="monitoring-alerts-detail-value">{alert.id}</span>
            </div>
            
            <div className="monitoring-alerts-detail-field">
              <label className="monitoring-alerts-detail-label">Host:</label>
              <span className="monitoring-alerts-detail-value">{alert.host}</span>
            </div>
            
            <div className="monitoring-alerts-detail-field">
              <label className="monitoring-alerts-detail-label">Severity:</label>
              <span className={`monitoring-alerts-detail-severity-badge ${getSeverityClass(alert.severity)}`}>
                {alert.severity.toUpperCase()}
              </span>
            </div>
            
            <div className="monitoring-alerts-detail-field">
              <label className="monitoring-alerts-detail-label">Status:</label>
              <span className={`monitoring-alerts-detail-status-badge ${alert.status === 'seen' ? 'seen' : 'unseen'}`}>
                {alert.status === 'seen' ? 'Acknowledged' : 'Not Acknowledged'}
              </span>
            </div>
            
            <div className="monitoring-alerts-detail-field">
              <label className="monitoring-alerts-detail-label">Time:</label>
              <span className="monitoring-alerts-detail-value">{formatTime(alert.time)}</span>
            </div>
          </div>

          <div className="monitoring-alerts-detail-message">
            <label className="monitoring-alerts-detail-label">Message:</label>
            <div className="monitoring-alerts-detail-message-content">
              {alert.content}
            </div>
          </div>
        </div>

        <div className="monitoring-alerts-detail-actions">
          {alert.status === 'notseen' && (
            <button 
              className="monitoring-alerts-detail-btn monitoring-alerts-detail-btn-acknowledge"
              onClick={() => onAcknowledge(alert.id)}
              disabled={isAcknowledging}
            >
              <FaEye className="monitoring-alerts-detail-btn-icon" />
              {isAcknowledging ? 'Acknowledging...' : 'Acknowledge'}
            </button>
          )}
          
          <button 
            className="monitoring-alerts-detail-btn monitoring-alerts-detail-btn-resolve"
            onClick={() => onResolve(alert.id)}
            disabled={isResolving}
          >
            <FaCheck className="monitoring-alerts-detail-btn-icon" />
            {isResolving ? 'Resolving...' : 'Resolve'}
          </button>
          
          <button 
            className="monitoring-alerts-detail-btn monitoring-alerts-detail-btn-cancel"
            onClick={onClose}
          >
            Cancel
          </button>
        </div>
      </div>
    </div>
  );
};

export default AlertDetailModal;
