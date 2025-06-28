// components/server/config2/config2.tsx
import React, { useState } from 'react';
import { FaSync, FaSpinner } from 'react-icons/fa';
import NetworkBasicInfo from './NetworkBasicInfo';
import NetworkInterfaceManager from './NetworkInterfaceManager';
import RouteTable from './RouteTable';
import FirewallTable from './FirewallTable';
import { useConfig2 } from '../../../../hooks/server/useConfig2';
import { useNotification } from '../../../../context/NotificationContext';
import './config2.css';

const Config2: React.FC = () => {
  const { 
    loading, 
    refreshAllData,
    refreshTrigger // ✅ NEW: Get refresh trigger from hook
  } = useConfig2();
  const { addNotification } = useNotification();
  
  const [isRefreshing, setIsRefreshing] = useState(false);

  const handleRefreshAll = async () => {
    try {
      setIsRefreshing(true);

      addNotification({
        title: "Refreshing Network Data",
        message: "Updating all network configuration data...",
        type: "info",
        duration: 2000,
      });

      // ✅ This will trigger the refresh and update refreshTrigger
      await refreshAllData();

      addNotification({
        title: "Refresh Complete",
        message: "All network data has been updated successfully",
        type: "success",
        duration: 3000,
      });
    } catch (error) {
      addNotification({
        title: "Refresh Failed",
        message: "Failed to refresh some network data. Please try again.",
        type: "error",
        duration: 5000,
      });
    } finally {
      setIsRefreshing(false);
    }
  };

  const isAnyComponentLoading = isRefreshing || 
                                loading.networkBasics || 
                                loading.routeTable || 
                                loading.firewallData || 
                                loading.updating;

  return (
    <div className="network-config2-main-container">
      <div className="network-config2-header">
        <div className="network-config2-header-content">
          <div className="network-config2-title-section">
            <h1 className="network-config2-title">Network Configuration</h1>
            <p className="network-config2-subtitle">
              Manage network settings and interface configurations
            </p>
          </div>
          
          <div className="network-config2-header-actions">
            <button
              className={`network-config2-refresh-btn ${isAnyComponentLoading ? 'loading' : ''}`}
              onClick={handleRefreshAll}
              disabled={isAnyComponentLoading}
              title="Refresh all network data"
            >
              {isAnyComponentLoading ? (
                <FaSpinner className="spinning refresh-icon" />
              ) : (
                <FaSync className="refresh-icon" />
              )}
              <span className="refresh-text">
                {isAnyComponentLoading ? 'Refreshing...' : 'Refresh All'}
              </span>
            </button>
          </div>
        </div>

        {isAnyComponentLoading && (
          <div className="network-config2-loading-progress">
            <div className="loading-progress-bar">
              <div className="loading-progress-fill"></div>
            </div>
            <span className="loading-progress-text">
              {isRefreshing ? 'Fetching network data...' : 'Updating network configuration data...'}
            </span>
          </div>
        )}
      </div>
      
      {/* ✅ Pass refreshTrigger to child components */}
      <div className="network-config2-cards-grid">
        <NetworkBasicInfo key={`network-${refreshTrigger}`} />
        <NetworkInterfaceManager key={`interface-${refreshTrigger}`} />
      </div>

      <div className="network-config2-route-section">
        <RouteTable key={`route-${refreshTrigger}`} />
      </div>

      <div className="network-config2-firewall-section">
        <FirewallTable key={`firewall-${refreshTrigger}`} />
      </div>
    </div>
  );
};

export default Config2;
