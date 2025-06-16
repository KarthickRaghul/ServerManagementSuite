// components/common/health/PortsModal.tsx
import React from 'react';
import { FaTimes, FaGlobe, FaExclamationTriangle } from 'react-icons/fa';
import './PortsModal.css';

interface Port {
  protocol: string;
  port: number;
  process: string;
}

interface PortsModalProps {
  isOpen: boolean;
  onClose: () => void;
  ports: Port[];
}

const PortsModal: React.FC<PortsModalProps> = ({ isOpen, onClose, ports }) => {
  if (!isOpen) return null;

  // Remove duplicate ports and group by port number
  const getUniquePortsWithDetails = () => {
    const portMap = new Map<number, Port[]>();
    
    // Group ports by port number
    ports.forEach(port => {
      if (!portMap.has(port.port)) {
        portMap.set(port.port, []);
      }
      portMap.get(port.port)!.push(port);
    });

    // Convert to array with unique ports and their details
    return Array.from(portMap.entries()).map(([portNumber, portDetails]) => ({
      port: portNumber,
      details: portDetails,
      isDuplicate: portDetails.length > 1,
      processes: [...new Set(portDetails.map(p => p.process))], // Unique processes
      protocols: [...new Set(portDetails.map(p => p.protocol))] // Unique protocols
    }));
  };

  const uniquePorts = getUniquePortsWithDetails();

  const getPortSeverity = (port: number): 'critical' | 'standard' => {
    const criticalPorts = [22, 80, 443, 3306, 5432, 1433, 21, 23, 25, 53, 110, 143, 993, 995];
    return criticalPorts.includes(port) ? 'critical' : 'standard';
  };

  const criticalPorts = uniquePorts.filter(port => getPortSeverity(port.port) === 'critical');
  const standardPorts = uniquePorts.filter(port => getPortSeverity(port.port) === 'standard');
  const duplicatedPorts = uniquePorts.filter(port => port.isDuplicate);

  const handleBackdropClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose();
    }
  };

  return (
    <div className="health-ports-modal-overlay" onClick={handleBackdropClick}>
      <div className="health-ports-modal-container">
        <div className="health-ports-modal-header">
          <div className="health-ports-modal-title-section">
            <FaGlobe className="health-ports-modal-icon" />
            <h2 className="health-ports-modal-title">Open Ports Details</h2>
          </div>
          <button 
            className="health-ports-modal-close"
            onClick={onClose}
            type="button"
          >
            <FaTimes />
          </button>
        </div>

        <div className="health-ports-modal-content">
          <div className="health-ports-modal-stats">
            <div className="health-ports-stat-item">
              <span className="health-ports-stat-value">{uniquePorts.length}</span>
              <span className="health-ports-stat-label">Unique Ports</span>
            </div>
            <div className="health-ports-stat-item">
              <span className="health-ports-stat-value">{criticalPorts.length}</span>
              <span className="health-ports-stat-label">Critical</span>
            </div>
            <div className="health-ports-stat-item">
              <span className="health-ports-stat-value">{duplicatedPorts.length}</span>
              <span className="health-ports-stat-label">Duplicated</span>
            </div>
          </div>

          <p className="health-ports-modal-subtitle">
            Complete list of unique open ports and their security status
          </p>

          {/* Duplicated Ports Warning */}
          {duplicatedPorts.length > 0 && (
            <div className="health-ports-modal-warning">
              <FaExclamationTriangle className="health-ports-modal-warning-icon" />
              <div className="health-ports-modal-warning-content">
                <h4>Duplicate Ports Detected</h4>
                <p>
                  {duplicatedPorts.length} port(s) have multiple bindings: {duplicatedPorts.map(p => p.port).join(', ')}
                </p>
              </div>
            </div>
          )}

          {/* Port Grid */}
          <div className="health-ports-modal-grid">
            {/* Critical Ports */}
            {criticalPorts.map((portInfo, index) => (
              <div 
                key={`critical-${index}`} 
                className={`health-ports-modal-item critical ${portInfo.isDuplicate ? 'duplicate' : ''}`}
                title={`Processes: ${portInfo.processes.join(', ')}\nProtocols: ${portInfo.protocols.join(', ')}`}
              >
                <span className="health-ports-modal-port-number">{portInfo.port}</span>
                <span className="health-ports-modal-port-label">
                  {portInfo.isDuplicate ? 'Critical (Dup)' : 'Critical'}
                </span>
                {portInfo.isDuplicate && (
                  <span className="health-ports-modal-duplicate-indicator">
                    {portInfo.details.length}x
                  </span>
                )}
              </div>
            ))}
            
            {/* Standard Ports */}
            {standardPorts.map((portInfo, index) => (
              <div 
                key={`standard-${index}`} 
                className={`health-ports-modal-item standard ${portInfo.isDuplicate ? 'duplicate' : ''}`}
                title={`Processes: ${portInfo.processes.join(', ')}\nProtocols: ${portInfo.protocols.join(', ')}`}
              >
                <span className="health-ports-modal-port-number">{portInfo.port}</span>
                <span className="health-ports-modal-port-label">
                  {portInfo.isDuplicate ? 'Standard (Dup)' : 'Standard'}
                </span>
                {portInfo.isDuplicate && (
                  <span className="health-ports-modal-duplicate-indicator">
                    {portInfo.details.length}x
                  </span>
                )}
              </div>
            ))}
          </div>

          {/* Detailed Port Information */}
          {duplicatedPorts.length > 0 && (
            <div className="health-ports-modal-details">
              <h4>Duplicate Port Details</h4>
              <div className="health-ports-detail-list">
                {duplicatedPorts.map((portInfo, index) => (
                  <div key={index} className="health-ports-detail-item">
                    <div className="health-ports-detail-header">
                      <span className="health-ports-detail-port">Port {portInfo.port}</span>
                      <span className="health-ports-detail-count">{portInfo.details.length} bindings</span>
                    </div>
                    <div className="health-ports-detail-processes">
                      {portInfo.details.map((detail, idx) => (
                        <div key={idx} className="health-ports-process-item">
                          <span className="health-ports-process-protocol">{detail.protocol}</span>
                          <span className="health-ports-process-name">{detail.process}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Security Notice */}
          <div className="health-ports-modal-notice">
            <FaExclamationTriangle className="health-ports-modal-notice-icon" />
            <div className="health-ports-modal-notice-content">
              <h4>Security Notice</h4>
              <p>
                {criticalPorts.length > 0 && `Critical ports (${criticalPorts.map(p => p.port).join(', ')}) are exposed. `}
                {duplicatedPorts.length > 0 && `${duplicatedPorts.length} duplicate port binding(s) detected. `}
                Ensure proper firewall rules and process management are in place.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default PortsModal;
