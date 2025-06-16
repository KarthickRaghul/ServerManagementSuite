// components/server/ResourceOptimization/ResourceOptimization.tsx
import React, { useState } from 'react';
import UsageStats from './UsageStats';
import ResourceCard from './ResourceCard';
import ServiceGrid from './ServiceGrid';
import { FaSync, FaChartLine, FaServer, FaPlay, FaInfoCircle } from 'react-icons/fa';
import { useHealthMetrics } from '../../../hooks/server/useHealthMetrics';
import { useServerOverview } from '../../../hooks/server/useServerOverview';
import { useResourceOptimization } from '../../../hooks/server/useResourceOptimization';
import { useNotification } from '../../../context/NotificationContext';
import "./Resource.css";

const ResourceOptimization: React.FC = () => {
  const [serviceToRestart, setServiceToRestart] = useState('');
  
  const { metrics, loading: metricsLoading, error: metricsError, refreshMetrics } = useHealthMetrics();
  const { data: overviewData, loading: overviewLoading } = useServerOverview();
  const { 
    cleanupInfo, 
    services, 
    loading: resourceLoading, 
    error: resourceError,
    optimizing,
    restartingService,
    optimizeSystem,
    restartService,
    refreshData
  } = useResourceOptimization();
  const { addNotification } = useNotification();

  const handleRefreshData = async () => {
    await Promise.all([refreshMetrics(), refreshData()]);
    addNotification({
      title: 'Data Refreshed',
      message: 'Resource data has been refreshed successfully',
      type: 'info',
      duration: 3000
    });
  };

  const handleOptimizeSystem = async () => {
    try {
      const success = await optimizeSystem();
      if (success) {
        addNotification({
          title: 'System Optimized',
          message: 'Temporary files have been cleaned successfully',
          type: 'success',
          duration: 4000
        });
      }
    } catch (err) {
      addNotification({
        title: 'Optimization Failed',
        message: err instanceof Error ? err.message : 'Failed to optimize system',
        type: 'error',
        duration: 5000
      });
    }
  };

  const handleRestartService = async () => {
    if (!serviceToRestart.trim()) {
      addNotification({
        title: 'Service Name Required',
        message: 'Please enter a service name to restart',
        type: 'warning',
        duration: 3000
      });
      return;
    }

    try {
      const success = await restartService(serviceToRestart.trim());
      if (success) {
        addNotification({
          title: 'Service Restarted',
          message: `Service "${serviceToRestart}" has been restarted successfully`,
          type: 'success',
          duration: 4000
        });
        setServiceToRestart(''); // Clear the input after successful restart
      }
    } catch (err) {
      addNotification({
        title: 'Restart Failed',
        message: err instanceof Error ? err.message : `Failed to restart service "${serviceToRestart}"`,
        type: 'error',
        duration: 5000
      });
    }
  };

  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const getTotalCleanupSize = (): string => {
    if (!cleanupInfo) return '0 B';
    const total = Object.values(cleanupInfo.sizes).reduce((sum, size) => sum + size, 0);
    return formatBytes(total);
  };

  const getServicesCount = (): number => {
    return services.length;
  };

  return (
    <div className="resource-dashboard">
      {/* Header Section */}
      <div className="resource-header-section">
        <div className="resource-title-section">
          <h1 className="resource-page-title">Resource Optimization</h1>
          <p className="resource-page-subtitle">System performance monitoring and optimization tools</p>
        </div>
        <div className="resource-header-actions">
          <button 
            className="resource-btn resource-btn-secondary"
            onClick={handleRefreshData}
            disabled={metricsLoading || resourceLoading}
          >
            <FaSync className={`resource-btn-icon ${(metricsLoading || resourceLoading) ? 'spinning' : ''}`} />
            Refresh Data
          </button>
        </div>
      </div>

      {/* Show errors */}
      {(metricsError || resourceError) && (
        <div className="resource-error-banner">
          <p>Error: {metricsError || resourceError}</p>
        </div>
      )}

      {/* Metrics Section */}
      <div className="resource-metrics-container">
        <div className="resource-metrics-grid">
          <UsageStats 
            label="Memory Usage" 
            value={metrics?.ram || 0} 
            type="memory" 
            loading={metricsLoading}
          />
          <UsageStats 
            label="CPU Usage" 
            value={metrics?.cpu || 0} 
            type="cpu" 
            loading={metricsLoading}
          />
          <UsageStats 
            label="Disk Usage" 
            value={metrics?.disk || 0} 
            type="disk" 
            loading={metricsLoading}
          />
          <div className="resource-uptime-card">
            <div className="resource-uptime-icon">
              <FaServer />
            </div>
            <div className="resource-uptime-info">
              <p className="resource-uptime-value">
                {overviewLoading ? 'Loading...' : (overviewData?.uptime || 'N/A')}
              </p>
              <p className="resource-uptime-label">Server Uptime</p>
            </div>
          </div>
        </div>
      </div>

      {/* Resource Cards Section */}
      <div className="resource-cards-container">
        <div className="resource-section-header">
          <h2 className="resource-section-title">
            <FaChartLine className="resource-section-icon" />
            System Optimization
          </h2>
        </div>
        <div className="resource-cards-grid">
          <ResourceCard
            icon={<FaSync />}
            title="Temporary File Cleanup"
            description="Clean system temporary files and cache"
            stats={[
              getTotalCleanupSize(),
              cleanupInfo ? `${cleanupInfo.folders.length} folders` : 'Loading...'
            ]}
            type="cleanup"
            onAction={handleOptimizeSystem}
            loading={optimizing}
            disabled={!cleanupInfo || optimizing}
          />
          
          {/* Service Restart Card */}
          <div className="resource-service-restart-card">
            <div className="resource-optimization-header">
              <div className="resource-optimization-icon resource-service-restart-icon">
                <FaServer />
              </div>
              <div className="resource-optimization-info">
                <h4 className="resource-optimization-title">Service Management</h4>
                <p className="resource-optimization-description">Restart system daemon services</p>
              </div>
            </div>
            
            <div className="resource-optimization-stats">
              <div className="resource-optimization-stat">
                <span className="resource-optimization-stat-value">{getServicesCount()}</span>
                <span className="resource-optimization-stat-label">Services Available</span>
              </div>
              <div className="resource-optimization-stat">
                <span className="resource-optimization-stat-value">
                  {restartingService ? '1' : '0'}
                </span>
                <span className="resource-optimization-stat-label">Currently Restarting</span>
              </div>
            </div>

            <div className="resource-service-restart-notice">
              <FaInfoCircle className="resource-service-notice-icon" />
              <span className="resource-service-notice-text">
                Only daemon services can be restarted. Enter the exact service name below.
              </span>
            </div>
            
            <div className="resource-service-restart-controls">
              <input
                type="text"
                className="resource-service-restart-input"
                placeholder="Enter service name (e.g., nginx, postgresql)"
                value={serviceToRestart}
                onChange={(e) => setServiceToRestart(e.target.value)}
                disabled={!!restartingService}
              />
              <button 
                className="resource-service-restart-btn"
                onClick={handleRestartService}
                disabled={!serviceToRestart.trim() || !!restartingService}
              >
                {restartingService ? (
                  <>
                    <FaSync className="resource-service-btn-icon spinning" />
                    Restarting...
                  </>
                ) : (
                  <>
                    <FaPlay className="resource-service-btn-icon" />
                    Restart Service
                  </>
                )}
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Service Overview Section */}
      <div className="resource-services-container">
        <div className="resource-section-header">
          <h2 className="resource-section-title">Service Overview</h2>
          <p className="resource-section-subtitle">View running system services ({services.length} services)</p>
        </div>
        <ServiceGrid 
          services={services}
          loading={resourceLoading}
        />
      </div>
    </div>
  );
};

export default ResourceOptimization;
