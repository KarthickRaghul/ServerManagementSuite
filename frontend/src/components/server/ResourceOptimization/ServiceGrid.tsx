// components/server/ResourceOptimization/ServiceGrid.tsx
import React from "react";
import { FaCube } from "react-icons/fa";
import "./serviceGrid.css";

interface Service {
  pid: number;
  user: string;
  name: string;
  cmdline: string;
}

interface ServiceGridProps {
  services: Service[];
  loading: boolean;
}

const ServiceGrid: React.FC<ServiceGridProps> = ({ services, loading }) => {
  if (loading) {
    return (
      <div className="resource-service-loading">
        <div className="resource-service-loading-spinner"></div>
        <p>Loading services...</p>
      </div>
    );
  }

  if (services.length === 0) {
    return (
      <div className="resource-service-empty">
        <div className="resource-service-empty-icon">
          <FaCube />
        </div>
        <p>No services found</p>
      </div>
    );
  }

  // ✅ Create unique services to avoid duplicate keys
  const getUniqueServices = () => {
    const serviceMap = new Map<string, Service & { count: number }>();

    services.forEach((service) => {
      const key = `${service.pid}-${service.name}`;
      if (serviceMap.has(key)) {
        // If duplicate, increment count
        const existing = serviceMap.get(key)!;
        existing.count += 1;
      } else {
        // Add new service with count
        serviceMap.set(key, { ...service, count: 1 });
      }
    });

    return Array.from(serviceMap.entries()).map(([key, service], index) => ({
      ...service,
      uniqueKey: `${key}-${index}`, // Ensure absolutely unique keys
    }));
  };

  const uniqueServices = getUniqueServices();

  return (
    <div className="resource-service-grid">
      {uniqueServices.map((service) => (
        <div key={service.uniqueKey} className="resource-service-box">
          <div className="resource-service-header">
            <div className="resource-service-icon">
              <FaCube />
            </div>
            <div className="resource-service-info">
              <div className="resource-service-name">
                {service.name}
                {service.count > 1 && (
                  <span className="resource-service-count">
                    {" "}
                    ({service.count}x)
                  </span>
                )}
              </div>
              <div className="resource-service-user">User: {service.user}</div>
            </div>
          </div>

          <div className="resource-service-details">
            <div className="resource-service-pid">PID: {service.pid}</div>
            <div className="resource-service-status">
              <span className="resource-service-status-dot"></span>
              Running
            </div>
          </div>

          <div className="resource-service-cmdline">
            {service.cmdline.length > 50
              ? `${service.cmdline.substring(0, 50)}...`
              : service.cmdline}
          </div>

          {service.count > 1 && (
            <div className="resource-service-duplicate-notice">
              <span className="resource-service-duplicate-icon">⚠️</span>
              <span className="resource-service-duplicate-text">
                {service.count} instances of this service are running
              </span>
            </div>
          )}
        </div>
      ))}
    </div>
  );
};

export default ServiceGrid;
