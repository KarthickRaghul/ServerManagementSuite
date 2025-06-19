// components/server/config2/NetworkInterfaceManager.tsx
import React from 'react';
import { FaCog, FaEthernet, FaWifi, FaSync, FaClock, FaSpinner } from 'react-icons/fa';
import { FiRefreshCw } from 'react-icons/fi';
import { useConfig2 } from '../../../../hooks/server/useConfig2';
import { useNotification } from '../../../../context/NotificationContext';
import './NetworkInterfaceManager.css';

interface NetworkInterface {
  id: string;
  name: string;
  status: string;
  type: 'wifi' | 'ethernet';
}

const NetworkInterfaceManager: React.FC = () => {
  const { networkBasics, loading, updateInterface, restartInterface } = useConfig2();
  const { addNotification } = useNotification();

  const handleToggle = async (iface: string, action: 'enable' | 'disable') => {
    try {
      const success = await updateInterface(iface, action);
      if (success) {
        addNotification({
          title: 'Interface Updated',
          message: `Interface ${iface} has been ${action}d successfully`,
          type: 'success',
          duration: 3000
        });
      } else {
        throw new Error(`Failed to ${action} interface`);
      }
    } catch (err) {
      addNotification({
        title: 'Interface Update Failed',
        message: err instanceof Error ? err.message : `Failed to ${action} interface`,
        type: 'error',
        duration: 5000
      });
    }
  };

  const handleRestart = async () => {
    try {
      const success = await restartInterface();
      if (success) {
        addNotification({
          title: 'Interface Restarted',
          message: 'Network interface has been restarted successfully',
          type: 'success',
          duration: 3000
        });
      } else {
        throw new Error('Failed to restart interface');
      }
    } catch (err) {
      addNotification({
        title: 'Restart Failed',
        message: err instanceof Error ? err.message : 'Failed to restart interface',
        type: 'error',
        duration: 5000
      });
    }
  };

  const getInterfaceArray = (): NetworkInterface[] => {
    if (!networkBasics?.interface) return [];
    return Object.entries(networkBasics.interface).map(([key, value]) => ({
      id: key,
      name: value.mode,
      status: value.status,
      type: value.mode.includes('wlan') ? 'wifi' : 'ethernet'
    }));
  };

  const interfaces = getInterfaceArray();

  if (loading.networkBasics && !networkBasics) {
    return (
      <div className="network-interface-manager-card">
        <div className="network-interface-manager-loading">
          <div className="network-interface-manager-loading-spinner">
            <FaSpinner className="spinning" />
          </div>
          <p>Loading interface information...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="network-interface-manager-card">
      <div className="network-interface-manager-header">
        <div className="network-interface-manager-title-section">
          <div className="network-interface-manager-icon-wrapper">
            <FaCog className="network-interface-manager-icon" />
          </div>
          <div>
            <h3 className="network-interface-manager-title">Interface Manager</h3>
            <p className="network-interface-manager-description">Control network interfaces</p>
          </div>
        </div>
      </div>

      <div className="network-interface-manager-content">
        {/* Interface List */}
        <div className="network-interface-manager-interfaces-section">
          <h4 className="network-interface-manager-section-title">Network Interfaces</h4>
          <div className="network-interface-manager-interfaces-list">
            {interfaces.length === 0 ? (
              <div className="network-interface-manager-no-interfaces">
                <FaEthernet className="network-interface-manager-no-interfaces-icon" />
                <p>No network interfaces found</p>
              </div>
            ) : (
              interfaces.map((iface) => (
                <div className="network-interface-manager-interface-item" key={iface.id}>
                  <div className="network-interface-manager-interface-info">
                    <div className="network-interface-manager-interface-icon-wrapper">
                      {iface.type === 'wifi' ? 
                        <FaWifi className="network-interface-manager-interface-icon" /> : 
                        <FaEthernet className="network-interface-manager-interface-icon" />
                      }
                    </div>
                    <div className="network-interface-manager-interface-details">
                      <span className="network-interface-manager-interface-name">{iface.name}</span>
                      <span className={`network-interface-manager-interface-status ${iface.status}`}>
                        {iface.status === 'active' ? 'Active' : 'Inactive'}
                      </span>
                    </div>
                  </div>
                  <div className="network-interface-manager-interface-controls">
                    <button
                      className={`network-interface-manager-control-btn ${iface.status === 'active' ? 'enabled' : 'enable'}`}
                      onClick={() => handleToggle(iface.name, 'enable')}
                      disabled={loading.updating || iface.status === 'active'}
                    >
                      {loading.updating ? <FaSpinner className="spinning" /> : 'Enable'}
                    </button>
                    <button
                      className={`network-interface-manager-control-btn ${iface.status === 'inactive' ? 'disabled' : 'disable'}`}
                      onClick={() => handleToggle(iface.name, 'disable')}
                      disabled={loading.updating || iface.status === 'inactive'}
                    >
                      {loading.updating ? <FaSpinner className="spinning" /> : 'Disable'}
                    </button>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Status Section */}
        <div className="network-interface-manager-status-section">
          <div className="network-interface-manager-status-grid">
            <div className="network-interface-manager-status-item">
              <div className="network-interface-manager-status-icon-wrapper status">
                <FaSync />
              </div>
              <div className="network-interface-manager-status-content">
                <label>Status</label>
                <span className={`network-interface-manager-status-value ${networkBasics?.ip_address ? 'online' : 'offline'}`}>
                  {networkBasics?.ip_address ? 'Online' : 'Offline'}
                </span>
              </div>
            </div>
            <div className="network-interface-manager-status-item">
              <div className="network-interface-manager-status-icon-wrapper uptime">
                <FaClock />
              </div>
              <div className="network-interface-manager-status-content">
                <label>Uptime</label>
                <span className="network-interface-manager-status-value">
                  {networkBasics?.uptime || 'N/A'}
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Restart Button */}
        <button
          className={`network-interface-manager-restart-btn ${loading.updating ? 'loading' : ''}`}
          onClick={handleRestart}
          disabled={loading.updating}
        >
          <FiRefreshCw className={`network-interface-manager-restart-icon ${loading.updating ? 'spinning' : ''}`} />
          {loading.updating ? 'Restarting Network...' : 'Restart Network Service'}
        </button>
      </div>
    </div>
  );
};

export default NetworkInterfaceManager;
