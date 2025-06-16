// components/server/ResourceOptimization/ServiceGrid.tsx
import React from 'react';
import { FaCube } from 'react-icons/fa';
import './serviceGrid.css';

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

  return (
    <div className="resource-service-grid">
      {services.map(service => (
        <div key={`${service.pid}-${service.name}`} className="resource-service-box">
          <div className="resource-service-header">
            <div className="resource-service-icon">
              <FaCube />
            </div>
            <div className="resource-service-info">
              <div className="resource-service-name">{service.name}</div>
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
              : service.cmdline
            }
          </div>
        </div>
      ))}
    </div>
  );
};

export default ServiceGrid;
