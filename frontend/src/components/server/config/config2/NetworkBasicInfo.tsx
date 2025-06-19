// components/server/config2/NetworkBasicInfo.tsx
import React, { useState } from 'react';
import { FaNetworkWired, FaEdit, FaWifi, FaServer, FaSpinner } from 'react-icons/fa';
import { useConfig2 } from '../../../../hooks/server/useConfig2';
import { useNotification } from '../../../../context/NotificationContext';
import NetworkConfigModal from './NetworkConfigModal';
import './NetworkBasicInfo.css';

interface NetworkUpdateData {
  method: string;
  ip?: string;
  subnet?: string;
  gateway?: string;
  dns?: string;
}

const NetworkBasicInfo: React.FC = () => {
  const [showConfigModal, setShowConfigModal] = useState(false);
  const { networkBasics, loading, updateNetwork } = useConfig2();
  const { addNotification } = useNotification();

  const handleUpdateNetwork = async (networkData: NetworkUpdateData) => {
    try {
      const success = await updateNetwork(networkData);
      if (success) {
        addNotification({
          title: 'Network Updated',
          message: 'Network configuration has been updated successfully',
          type: 'success',
          duration: 3000
        });
        setShowConfigModal(false);
        return true;
      } else {
        throw new Error('Failed to update network configuration');
      }
    } catch (err) {
      addNotification({
        title: 'Network Update Failed',
        message: err instanceof Error ? err.message : 'Failed to update network configuration',
        type: 'error',
        duration: 5000
      });
      return false;
    }
  };

  if (loading.networkBasics && !networkBasics) {
    return (
      <div className="network-basic-info-card">
        <div className="network-basic-info-loading">
          <div className="network-basic-info-loading-spinner">
            <FaSpinner className="spinning" />
          </div>
          <p>Loading network information...</p>
        </div>
      </div>
    );
  }

  return (
    <>
      <div className="network-basic-info-card">
        <div className="network-basic-info-header">
          <div className="network-basic-info-title-section">
            <div className="network-basic-info-icon-wrapper">
              <FaNetworkWired className="network-basic-info-icon" />
            </div>
            <div>
              <h3 className="network-basic-info-title">Network Settings</h3>
              <p className="network-basic-info-description">Current network configuration</p>
            </div>
          </div>
          <button 
            className="network-basic-info-edit-btn"
            onClick={() => setShowConfigModal(true)}
            disabled={loading.updating}
            title="Edit Network Configuration"
          >
            {loading.updating ? <FaSpinner className="spinning" /> : <FaEdit />}
          </button>
        </div>

        <div className="network-basic-info-content">
          <div className="network-basic-info-method-section">
            <div className="network-basic-info-method-indicator">
              <div className={`network-basic-info-method-badge ${networkBasics?.ip_method || 'unknown'}`}>
                <span className="network-basic-info-method-text">
                  {networkBasics?.ip_method === 'static' ? 'Static IP' : 
                   networkBasics?.ip_method === 'dynamic' ? 'Dynamic IP' : 'Unknown'}
                </span>
              </div>
            </div>
          </div>

          <div className="network-basic-info-details-grid">
            <div className="network-basic-info-detail-item">
              <div className="network-basic-info-detail-icon">
                <FaServer />
              </div>
              <div className="network-basic-info-detail-content">
                <label>IP Address</label>
                <span>{networkBasics?.ip_address || 'Not configured'}</span>
              </div>
            </div>

            <div className="network-basic-info-detail-item">
              <div className="network-basic-info-detail-icon">
                <FaWifi />
              </div>
              <div className="network-basic-info-detail-content">
                <label>Gateway</label>
                <span>{networkBasics?.gateway || 'Not configured'}</span>
              </div>
            </div>

            <div className="network-basic-info-detail-item">
              <div className="network-basic-info-detail-icon">
                <FaNetworkWired />
              </div>
              <div className="network-basic-info-detail-content">
                <label>Subnet</label>
                <span>{networkBasics?.subnet || 'Not configured'}</span>
              </div>
            </div>

            <div className="network-basic-info-detail-item network-basic-info-detail-dns">
              <div className="network-basic-info-detail-icon">
                <FaServer />
              </div>
              <div className="network-basic-info-detail-content">
                <label>DNS Servers</label>
                <span>{networkBasics?.dns || 'Not configured'}</span>
              </div>
            </div>
          </div>

          <div className="network-basic-info-status-bar">
            <div className="network-basic-info-status-indicator">
              <div className={`network-basic-info-status-dot ${networkBasics?.ip_address ? 'online' : 'offline'}`}></div>
              <span className="network-basic-info-status-text">
                {networkBasics?.ip_address ? 'Network Active' : 'Network Inactive'}
              </span>
            </div>
          </div>
        </div>
      </div>

      <NetworkConfigModal
        isOpen={showConfigModal}
        onClose={() => setShowConfigModal(false)}
        onSubmit={handleUpdateNetwork}
        currentConfig={networkBasics}
        isLoading={loading.updating}
      />
    </>
  );
};

export default NetworkBasicInfo;
