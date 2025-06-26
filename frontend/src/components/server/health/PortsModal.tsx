// components/common/health/PortsModal.tsx
import React from "react";
import {
  FaTimes,
  FaGlobe,
  FaExclamationTriangle,
  FaInfoCircle,
} from "react-icons/fa";
import "./PortsModal.css";

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

  // ✅ Enhanced port categorization based on port numbers and process names
  const getPortCategory = (
    port: number,
    process: string,
  ): { category: string; severity: "critical" | "warning" | "standard" } => {
    // Critical system ports
    if ([22, 23, 21, 3389].includes(port)) {
      return { category: "Remote Access", severity: "critical" };
    }
    // Database ports
    if (
      [1433, 3306, 5432, 6379, 27017, 9200].includes(port) ||
      process.toLowerCase().includes("postgres")
    ) {
      return { category: "Database", severity: "warning" };
    }
    // Web services
    if (
      [80, 443, 8080, 8443].includes(port) ||
      process.toLowerCase().includes("node")
    ) {
      return { category: "Web Service", severity: "standard" };
    }
    // Email services
    if ([25, 110, 143, 993, 995].includes(port)) {
      return { category: "Email Service", severity: "standard" };
    }
    // DNS and system services
    if ([53, 123].includes(port) || process.toLowerCase().includes("dns")) {
      return { category: "System Service", severity: "standard" };
    }
    // Development services
    if (
      [3000, 5173, 8000, 8500].includes(port) ||
      process.toLowerCase().includes("eslint") ||
      process.toLowerCase().includes("node")
    ) {
      return { category: "Development", severity: "standard" };
    }
    // Container services
    if (
      process.toLowerCase().includes("containerd") ||
      process.toLowerCase().includes("docker")
    ) {
      return { category: "Container", severity: "standard" };
    }
    return { category: "Application", severity: "standard" };
  };

  // Remove duplicate ports and group by port number with enhanced details
  const getUniquePortsWithDetails = () => {
    const portMap = new Map<number, Port[]>();

    // Group ports by port number
    ports.forEach((port) => {
      if (!portMap.has(port.port)) {
        portMap.set(port.port, []);
      }
      portMap.get(port.port)!.push(port);
    });

    // Convert to array with unique ports and their details
    return Array.from(portMap.entries())
      .map(([portNumber, portDetails]) => {
        const category = getPortCategory(portNumber, portDetails[0].process);
        return {
          port: portNumber,
          processName: portDetails[0].process, // ✅ Use actual process name from API
          category: category.category,
          severity: category.severity,
          details: portDetails,
          isDuplicate: portDetails.length > 1,
          processes: [...new Set(portDetails.map((p) => p.process))],
          protocols: [...new Set(portDetails.map((p) => p.protocol))],
        };
      })
      .sort((a, b) => a.port - b.port); // Sort by port number
  };

  const uniquePorts = getUniquePortsWithDetails();

  const criticalPorts = uniquePorts.filter(
    (port) => port.severity === "critical",
  );
  const warningPorts = uniquePorts.filter(
    (port) => port.severity === "warning",
  );
  const standardPorts = uniquePorts.filter(
    (port) => port.severity === "standard",
  );
  const duplicatedPorts = uniquePorts.filter((port) => port.isDuplicate);

  // ✅ Enhanced port statistics
  const getPortStats = () => {
    const categories = uniquePorts.reduce(
      (acc, port) => {
        acc[port.category] = (acc[port.category] || 0) + 1;
        return acc;
      },
      {} as { [key: string]: number },
    );

    return {
      total: uniquePorts.length,
      critical: criticalPorts.length,
      warning: warningPorts.length,
      standard: standardPorts.length,
      duplicated: duplicatedPorts.length,
      categories,
    };
  };

  const stats = getPortStats();

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
            <h2 className="health-ports-modal-title">Open Ports & Services</h2>
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
          {/* ✅ Enhanced Statistics */}
          <div className="health-ports-modal-stats">
            <div className="health-ports-stat-item">
              <span className="health-ports-stat-value">{stats.total}</span>
              <span className="health-ports-stat-label">Total Services</span>
            </div>
            <div className="health-ports-stat-item">
              <span className="health-ports-stat-value critical">
                {stats.critical}
              </span>
              <span className="health-ports-stat-label">Critical</span>
            </div>
            <div className="health-ports-stat-item">
              <span className="health-ports-stat-value warning">
                {stats.warning}
              </span>
              <span className="health-ports-stat-label">Database</span>
            </div>
            <div className="health-ports-stat-item">
              <span className="health-ports-stat-value">
                {stats.duplicated}
              </span>
              <span className="health-ports-stat-label">Duplicated</span>
            </div>
          </div>

          <p className="health-ports-modal-subtitle">
            Complete overview of active network services and their security
            classifications
          </p>

          {/* Security Warnings */}
          {criticalPorts.length > 0 && (
            <div className="health-ports-modal-warning">
              <FaExclamationTriangle className="health-ports-modal-warning-icon" />
              <div className="health-ports-modal-warning-content">
                <h4>Critical Services Exposed</h4>
                <p>
                  {criticalPorts.length} critical service(s) detected:{" "}
                  {criticalPorts
                    .map((p) => `${p.processName} (${p.port})`)
                    .join(", ")}
                  . Ensure proper firewall rules and access controls are in
                  place.
                </p>
              </div>
            </div>
          )}

          {/* Duplicate Ports Warning */}
          {duplicatedPorts.length > 0 && (
            <div className="health-ports-modal-warning">
              <FaInfoCircle className="health-ports-modal-warning-icon" />
              <div className="health-ports-modal-warning-content">
                <h4>Duplicate Port Bindings</h4>
                <p>
                  {duplicatedPorts.length} port(s) have multiple bindings:{" "}
                  {duplicatedPorts
                    .map((p) => `${p.port} (${p.processName})`)
                    .join(", ")}
                </p>
              </div>
            </div>
          )}

          {/* ✅ Enhanced Service Grid using actual process names */}
          <div className="health-ports-modal-grid">
            {uniquePorts.map((portInfo, index) => (
              <div
                key={index}
                className={`health-ports-modal-item ${portInfo.severity} ${portInfo.isDuplicate ? "duplicate" : ""}`}
                title={`${portInfo.processName}\nCategory: ${portInfo.category}\nProcesses: ${portInfo.processes.join(", ")}\nProtocols: ${portInfo.protocols.join(", ")}`}
              >
                <div className="health-ports-modal-service-header">
                  <span className="health-ports-modal-port-number">
                    {portInfo.port}
                  </span>
                  <span className="health-ports-modal-category-badge">
                    {portInfo.category}
                  </span>
                </div>
                <div className="health-ports-modal-service-info">
                  <span className="health-ports-modal-service-name">
                    {portInfo.processName}
                  </span>
                  <div className="health-ports-modal-service-details">
                    <span className="health-ports-modal-protocol">
                      {portInfo.protocols.join(", ")}
                    </span>
                  </div>
                </div>
                {portInfo.isDuplicate && (
                  <span className="health-ports-modal-duplicate-indicator">
                    {portInfo.details.length}x
                  </span>
                )}
              </div>
            ))}
          </div>

          {/* ✅ Enhanced Detailed Information */}
          {duplicatedPorts.length > 0 && (
            <div className="health-ports-modal-details">
              <h4>Duplicate Service Details</h4>
              <div className="health-ports-detail-list">
                {duplicatedPorts.map((portInfo, index) => (
                  <div key={index} className="health-ports-detail-item">
                    <div className="health-ports-detail-header">
                      <span className="health-ports-detail-service">
                        {portInfo.processName} (Port {portInfo.port})
                      </span>
                      <span className="health-ports-detail-count">
                        {portInfo.details.length} instances
                      </span>
                    </div>
                    <div className="health-ports-detail-processes">
                      {portInfo.details.map((detail, idx) => (
                        <div key={idx} className="health-ports-process-item">
                          <span className="health-ports-process-protocol">
                            {detail.protocol.toUpperCase()}
                          </span>
                          <span className="health-ports-process-name">
                            {detail.process}
                          </span>
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* ✅ Enhanced Security Notice */}
          <div className="health-ports-modal-notice">
            <FaInfoCircle className="health-ports-modal-notice-icon" />
            <div className="health-ports-modal-notice-content">
              <h4>Security Recommendations</h4>
              <p>
                {criticalPorts.length > 0 &&
                  `• Secure ${criticalPorts.length} critical service(s) with strong authentication and firewall rules. `}
                {warningPorts.length > 0 &&
                  `• Review ${warningPorts.length} database service(s) for proper access controls. `}
                {duplicatedPorts.length > 0 &&
                  `• Investigate ${duplicatedPorts.length} duplicate binding(s) for potential conflicts. `}
                Regular security audits and monitoring are recommended for all
                exposed services.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default PortsModal;
