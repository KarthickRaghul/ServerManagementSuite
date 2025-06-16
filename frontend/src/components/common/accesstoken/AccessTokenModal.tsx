// components/common/AccessTokenModal.tsx
import React, { useState } from 'react';
import './AccessTokenModal.css';
import { FaTimes, FaCopy, FaDownload, FaKey } from 'react-icons/fa';
import { useNotification } from '../../../context/NotificationContext';

interface AccessTokenModalProps {
  isOpen: boolean;
  onClose: () => void;
  accessToken: string;
  deviceInfo: {
    tag: string;
    ip: string;
    os: string;
  };
}

const AccessTokenModal: React.FC<AccessTokenModalProps> = ({
  isOpen,
  onClose,
  accessToken,
  deviceInfo
}) => {
  const [copied, setCopied] = useState(false);
  const { addNotification } = useNotification();

  if (!isOpen) return null;

  const handleCopyToken = async () => {
    try {
      await navigator.clipboard.writeText(accessToken);
      setCopied(true);
      addNotification({
        title: 'Token Copied',
        message: 'Access token has been copied to clipboard',
        type: 'success',
        duration: 3000
      });
      setTimeout(() => setCopied(false), 2000);
    } catch {
      addNotification({
        title: 'Copy Failed',
        message: 'Failed to copy token to clipboard',
        type: 'error',
        duration: 3000
      });
    }
  };

  const handleDownloadToken = () => {
    const tokenData = {
      device: deviceInfo,
      access_token: accessToken,
      generated_at: new Date().toISOString(),
      instructions: "Use this access token to configure the client device for secure communication with the management server."
    };

    const blob = new Blob([JSON.stringify(tokenData, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${deviceInfo.tag}-access-token.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);

    addNotification({
      title: 'Token Downloaded',
      message: 'Access token file has been downloaded',
      type: 'success',
      duration: 3000
    });
  };

  const handleBackdropClick = (e: React.MouseEvent) => {
    e.stopPropagation();
  };

  return (
    <div className="config1-servermgmt-accesstoken-overlay" onClick={handleBackdropClick}>
      <div className="config1-servermgmt-accesstoken-modal" onClick={(e) => e.stopPropagation()}>
        <div className="config1-servermgmt-accesstoken-header">
          <div className="config1-servermgmt-accesstoken-title">
            <FaKey className="config1-servermgmt-accesstoken-icon" />
            <h2>Device Access Token Generated</h2>
          </div>
          <button 
            className="config1-servermgmt-accesstoken-close"
            onClick={onClose}
            title="Close"
          >
            <FaTimes />
          </button>
        </div>

        <div className="config1-servermgmt-accesstoken-content">
          <div className="config1-servermgmt-accesstoken-device-info">
            <h3>Device Information</h3>
            <div className="config1-servermgmt-accesstoken-device-details">
              <div className="config1-servermgmt-accesstoken-detail-item">
                <span className="config1-servermgmt-accesstoken-detail-label">Device Name:</span>
                <span className="config1-servermgmt-accesstoken-detail-value">{deviceInfo.tag}</span>
              </div>
              <div className="config1-servermgmt-accesstoken-detail-item">
                <span className="config1-servermgmt-accesstoken-detail-label">IP Address:</span>
                <span className="config1-servermgmt-accesstoken-detail-value">{deviceInfo.ip}</span>
              </div>
              <div className="config1-servermgmt-accesstoken-detail-item">
                <span className="config1-servermgmt-accesstoken-detail-label">Operating System:</span>
                <span className="config1-servermgmt-accesstoken-detail-value">{deviceInfo.os}</span>
              </div>
            </div>
          </div>

          <div className="config1-servermgmt-accesstoken-section">
            <h3>Access Token</h3>
            <p className="config1-servermgmt-accesstoken-description">
              This access token is required for the client device to communicate securely with the management server. 
              Please copy or download this token and configure it on the target device.
            </p>
            
            <div className="config1-servermgmt-accesstoken-display">
              <div className="config1-servermgmt-accesstoken-value">
                {accessToken}
              </div>
              <div className="config1-servermgmt-accesstoken-actions">
                <button 
                  className={`config1-servermgmt-accesstoken-btn config1-servermgmt-accesstoken-btn-copy ${copied ? 'copied' : ''}`}
                  onClick={handleCopyToken}
                  title="Copy token"
                >
                  <FaCopy />
                  {copied ? 'Copied!' : 'Copy'}
                </button>
                <button 
                  className="config1-servermgmt-accesstoken-btn config1-servermgmt-accesstoken-btn-download"
                  onClick={handleDownloadToken}
                  title="Download token file"
                >
                  <FaDownload />
                  Download
                </button>
              </div>
            </div>
          </div>

          <div className="config1-servermgmt-accesstoken-warning">
            <div className="config1-servermgmt-accesstoken-warning-icon">⚠️</div>
            <div className="config1-servermgmt-accesstoken-warning-text">
              <strong>Important:</strong> This token will only be displayed once. Make sure to copy or download it before closing this window.
              Keep this token secure and do not share it with unauthorized personnel.
            </div>
          </div>
        </div>

        <div className="config1-servermgmt-accesstoken-footer">
          <button 
            className="config1-servermgmt-accesstoken-btn config1-servermgmt-accesstoken-btn-primary"
            onClick={onClose}
          >
            I have saved the token
          </button>
        </div>
      </div>
    </div>
  );
};

export default AccessTokenModal;
