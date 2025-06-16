// components/server/config2/NetworkConfigModal.tsx
import React, { useState, useEffect } from 'react';
import { FaTimes, FaNetworkWired, FaSave, FaWifi, FaServer } from 'react-icons/fa';
import './NetworkConfigModal.css';

interface NetworkBasics {
  ip_method: string;
  ip_address: string;
  gateway: string;
  subnet: string;
  dns: string;
  uptime: string;
  interface: {
    [key: string]: {
      mode: string;
      status: string;
    };
  };
}

interface NetworkConfigModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (networkData: {
    method: string;
    ip?: string;
    subnet?: string;
    gateway?: string;
    dns?: string;
  }) => Promise<boolean>;
  currentConfig: NetworkBasics | null;
  isLoading: boolean;
}

const NetworkConfigModal: React.FC<NetworkConfigModalProps> = ({
  isOpen,
  onClose,
  onSubmit,
  currentConfig,
  isLoading
}) => {
  const [formData, setFormData] = useState({
    method: 'static',
    ip: '',
    subnet: '',
    gateway: '',
    dns: ''
  });
  const [errors, setErrors] = useState<{[key: string]: string}>({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    if (currentConfig && isOpen) {
      setFormData({
        method: currentConfig.ip_method || 'static',
        ip: currentConfig.ip_address || '',
        subnet: currentConfig.subnet || '',
        gateway: currentConfig.gateway || '',
        dns: currentConfig.dns || ''
      });
      setErrors({});
    }
  }, [currentConfig, isOpen]);

  if (!isOpen) return null;

  const validateForm = () => {
    const newErrors: {[key: string]: string} = {};

    if (formData.method === 'static') {
      if (!formData.ip.trim()) {
        newErrors.ip = 'IP address is required for static configuration';
      } else if (!/^(\d{1,3}\.){3}\d{1,3}$/.test(formData.ip.trim())) {
        newErrors.ip = 'Please enter a valid IP address';
      }

      if (!formData.gateway.trim()) {
        newErrors.gateway = 'Gateway is required for static configuration';
      } else if (!/^(\d{1,3}\.){3}\d{1,3}$/.test(formData.gateway.trim())) {
        newErrors.gateway = 'Please enter a valid gateway address';
      }

      if (!formData.subnet.trim()) {
        newErrors.subnet = 'Subnet is required for static configuration';
      }

      if (formData.dns.trim() && !/^(\d{1,3}\.){3}\d{1,3}(,\s*(\d{1,3}\.){3}\d{1,3})*$/.test(formData.dns.trim())) {
        newErrors.dns = 'Please enter valid DNS servers (comma-separated)';
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) return;

    setIsSubmitting(true);

    const networkData: {
      method: string;
      ip?: string;
      subnet?: string;
      gateway?: string;
      dns?: string;
    } = {
      method: formData.method
    };

    if (formData.method === 'static') {
      networkData.ip = formData.ip.trim();
      networkData.subnet = formData.subnet.trim();
      networkData.gateway = formData.gateway.trim();
      if (formData.dns.trim()) {
        networkData.dns = formData.dns.trim();
      }
    }

    try {
      const success = await onSubmit(networkData);
      if (success) {
        setErrors({});
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleBackdropClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget && !isSubmitting) {
      onClose();
    }
  };

  const handleInputChange = (field: string, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: '' }));
    }
  };

  return (
    <div className="network-config-modal-overlay" onClick={handleBackdropClick}>
      <div className="network-config-modal-container">
        {/* Modal Header */}
        <div className="network-config-modal-header">
          <div className="network-config-modal-title-section">
            <div className="network-config-modal-icon-wrapper">
              <FaNetworkWired className="network-config-modal-icon" />
            </div>
            <div>
              <h2 className="network-config-modal-title">Network Configuration</h2>
              <p className="network-config-modal-subtitle">Configure network settings</p>
            </div>
          </div>
          <button 
            className="network-config-modal-close"
            onClick={onClose}
            type="button"
            disabled={isSubmitting}
          >
            <FaTimes />
          </button>
        </div>

        {/* Modal Content */}
        <form onSubmit={handleSubmit} className="network-config-modal-form">
          {/* IP Method Selection */}
          <div className="network-config-form-section">
            <h3 className="network-config-section-title">
              <FaWifi className="network-config-section-icon" />
              IP Configuration Method
            </h3>
            <div className="network-config-method-selector">
              <div 
                className={`network-config-method-option ${formData.method === 'dynamic' ? 'selected' : ''}`}
                onClick={() => !isSubmitting && handleInputChange('method', 'dynamic')}
              >
                <div className="network-config-method-radio">
                  <div className="network-config-method-radio-dot"></div>
                </div>
                <div className="network-config-method-content">
                  <h4>Dynamic (DHCP)</h4>
                  <p>Automatically obtain IP configuration from DHCP server</p>
                </div>
              </div>
              <div 
                className={`network-config-method-option ${formData.method === 'static' ? 'selected' : ''}`}
                onClick={() => !isSubmitting && handleInputChange('method', 'static')}
              >
                <div className="network-config-method-radio">
                  <div className="network-config-method-radio-dot"></div>
                </div>
                <div className="network-config-method-content">
                  <h4>Static IP</h4>
                  <p>Manually configure IP address and network settings</p>
                </div>
              </div>
            </div>
          </div>

          {/* Static IP Configuration */}
          {formData.method === 'static' && (
            <div className="network-config-form-section">
              <h3 className="network-config-section-title">
                <FaServer className="network-config-section-icon" />
                Static IP Configuration
              </h3>
              
              <div className="network-config-form-grid">
                <div className="network-config-form-group">
                  <label className="network-config-form-label">
                    IP Address *
                  </label>
                  <input
                    type="text"
                    className={`network-config-form-input ${errors.ip ? 'error' : ''}`}
                    placeholder="e.g., 192.168.1.100"
                    value={formData.ip}
                    onChange={(e) => handleInputChange('ip', e.target.value)}
                    disabled={isSubmitting}
                  />
                  {errors.ip && (
                    <span className="network-config-form-error">{errors.ip}</span>
                  )}
                </div>

                <div className="network-config-form-group">
                  <label className="network-config-form-label">
                    Gateway *
                  </label>
                  <input
                    type="text"
                    className={`network-config-form-input ${errors.gateway ? 'error' : ''}`}
                    placeholder="e.g., 192.168.1.1"
                    value={formData.gateway}
                    onChange={(e) => handleInputChange('gateway', e.target.value)}
                    disabled={isSubmitting}
                  />
                  {errors.gateway && (
                    <span className="network-config-form-error">{errors.gateway}</span>
                  )}
                </div>

                <div className="network-config-form-group network-config-form-group-full">
                  <label className="network-config-form-label">
                    Subnet *
                  </label>
                  <input
                    type="text"
                    className={`network-config-form-input ${errors.subnet ? 'error' : ''}`}
                    placeholder="e.g., 192.168.1.0/24 or 255.255.255.0"
                    value={formData.subnet}
                    onChange={(e) => handleInputChange('subnet', e.target.value)}
                    disabled={isSubmitting}
                  />
                  {errors.subnet && (
                    <span className="network-config-form-error">{errors.subnet}</span>
                  )}
                </div>

                <div className="network-config-form-group network-config-form-group-full">
                  <label className="network-config-form-label">
                    DNS Servers (Optional)
                  </label>
                  <input
                    type="text"
                    className={`network-config-form-input ${errors.dns ? 'error' : ''}`}
                    placeholder="e.g., 8.8.8.8, 8.8.4.4"
                    value={formData.dns}
                    onChange={(e) => handleInputChange('dns', e.target.value)}
                    disabled={isSubmitting}
                  />
                  {errors.dns && (
                    <span className="network-config-form-error">{errors.dns}</span>
                  )}
                  <small className="network-config-form-help">
                    Separate multiple DNS servers with commas
                  </small>
                </div>
              </div>
            </div>
          )}

          {/* Current Configuration Display */}
          {formData.method === 'dynamic' && currentConfig && (
            <div className="network-config-form-section">
              <h3 className="network-config-section-title">
                Current DHCP Configuration
              </h3>
              <div className="network-config-current-info">
                <div className="network-config-current-item">
                  <label>Current IP:</label>
                  <span>{currentConfig.ip_address || 'Not assigned'}</span>
                </div>
                <div className="network-config-current-item">
                  <label>Current Gateway:</label>
                  <span>{currentConfig.gateway || 'Not assigned'}</span>
                </div>
                <div className="network-config-current-item">
                  <label>Current DNS:</label>
                  <span>{currentConfig.dns || 'Not assigned'}</span>
                </div>
              </div>
            </div>
          )}

          {/* Modal Actions */}
          <div className="network-config-modal-actions">
            <button 
              type="button" 
              className="network-config-btn network-config-btn-cancel"
              onClick={onClose}
              disabled={isSubmitting}
            >
              Cancel
            </button>
            <button 
              type="submit" 
              className="network-config-btn network-config-btn-submit"
              disabled={isSubmitting || isLoading}
            >
              <FaSave className="network-config-btn-icon" />
              {isSubmitting ? 'Saving...' : 'Save Configuration'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default NetworkConfigModal;
