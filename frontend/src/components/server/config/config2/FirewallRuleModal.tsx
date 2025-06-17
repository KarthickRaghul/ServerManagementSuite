// components/server/config2/FirewallRuleModal.tsx
import React, { useState } from 'react';
import { createPortal } from 'react-dom';
import { FaTimes, FaShieldAlt, FaPlus } from 'react-icons/fa';
import './FirewallRuleModal.css';

interface FirewallRuleModalProps {
  isOpen: boolean;
  onClose: () => void;
  onAddRule: (ruleData: {
    action: string;
    rule: string;
    protocol: string;
    port: string;
    source?: string;
    destination?: string;
  }) => Promise<boolean>;
  isLoading: boolean;
}

const FirewallRuleModal: React.FC<FirewallRuleModalProps> = ({
  isOpen,
  onClose,
  onAddRule,
  isLoading
}) => {
  const [formData, setFormData] = useState({
    rule: 'accept',
    protocol: 'tcp',
    port: '',
    source: '',
    destination: ''
  });
  const [errors, setErrors] = useState<{[key: string]: string}>({});

  if (!isOpen) return null;

  const validateForm = () => {
    const newErrors: {[key: string]: string} = {};

    if (!formData.port.trim()) {
      newErrors.port = 'Port is required';
    } else if (!/^\d+(-\d+)?$/.test(formData.port.trim())) {
      newErrors.port = 'Please enter a valid port number or range (e.g., 80 or 8080-8090)';
    }

    if (formData.source.trim() && !/^(\d{1,3}\.){3}\d{1,3}(\/\d{1,2})?$/.test(formData.source.trim())) {
      newErrors.source = 'Please enter a valid IP address or CIDR notation';
    }

    if (formData.destination.trim() && !/^(\d{1,3}\.){3}\d{1,3}(\/\d{1,2})?$/.test(formData.destination.trim())) {
      newErrors.destination = 'Please enter a valid IP address or CIDR notation';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) return;

    const ruleData: {
      action: string;
      rule: string;
      protocol: string;
      port: string;
      source?: string;
      destination?: string;
    } = {
      action: 'add',
      rule: formData.rule,
      protocol: formData.protocol,
      port: formData.port.trim()
    };

    if (formData.source.trim()) {
      ruleData.source = formData.source.trim();
    }

    if (formData.destination.trim()) {
      ruleData.destination = formData.destination.trim();
    }

    const success = await onAddRule(ruleData);
    if (success) {
      setFormData({ rule: 'accept', protocol: 'tcp', port: '', source: '', destination: '' });
      setErrors({});
      onClose();
    }
  };

  const handleBackdropClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose();
    }
  };

  return createPortal(
    <div className="firewall-rule-modal-overlay" onClick={handleBackdropClick}>
      <div className="firewall-rule-modal-container">
        <div className="firewall-rule-modal-header">
          <div className="firewall-rule-modal-title-section">
            <div className="firewall-rule-modal-icon-wrapper">
              <FaShieldAlt className="firewall-rule-modal-icon" />
            </div>
            <h2 className="firewall-rule-modal-title">Add Firewall Rule</h2>
          </div>
          <button 
            className="firewall-rule-modal-close"
            onClick={onClose}
            type="button"
            disabled={isLoading}
          >
            <FaTimes />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="firewall-rule-modal-form">
          <div className="firewall-rule-form-row">
            <div className="firewall-rule-form-group">
              <label className="firewall-rule-form-label">Rule Action *</label>
              <select
                className="firewall-rule-form-select"
                value={formData.rule}
                onChange={(e) => setFormData({...formData, rule: e.target.value})}
                disabled={isLoading}
              >
                <option value="accept">Accept</option>
                <option value="drop">Drop</option>
                <option value="reject">Reject</option>
              </select>
            </div>

            <div className="firewall-rule-form-group">
              <label className="firewall-rule-form-label">Protocol *</label>
              <select
                className="firewall-rule-form-select"
                value={formData.protocol}
                onChange={(e) => setFormData({...formData, protocol: e.target.value})}
                disabled={isLoading}
              >
                <option value="tcp">TCP</option>
                <option value="udp">UDP</option>
              </select>
            </div>
          </div>

          <div className="firewall-rule-form-group">
            <label className="firewall-rule-form-label">Port *</label>
            <input
              type="text"
              className={`firewall-rule-form-input ${errors.port ? 'error' : ''}`}
              placeholder="e.g., 80, 443, 8080-8090"
              value={formData.port}
              onChange={(e) => setFormData({...formData, port: e.target.value})}
              disabled={isLoading}
            />
            {errors.port && (
              <span className="firewall-rule-form-error">{errors.port}</span>
            )}
          </div>

          <div className="firewall-rule-form-group">
            <label className="firewall-rule-form-label">Source (Optional)</label>
            <input
              type="text"
              className={`firewall-rule-form-input ${errors.source ? 'error' : ''}`}
              placeholder="e.g., 192.168.1.0/24 or any"
              value={formData.source}
              onChange={(e) => setFormData({...formData, source: e.target.value})}
              disabled={isLoading}
            />
            {errors.source && (
              <span className="firewall-rule-form-error">{errors.source}</span>
            )}
          </div>

          <div className="firewall-rule-form-group">
            <label className="firewall-rule-form-label">Destination (Optional)</label>
            <input
              type="text"
              className={`firewall-rule-form-input ${errors.destination ? 'error' : ''}`}
              placeholder="e.g., 192.168.1.0/24 or any"
              value={formData.destination}
              onChange={(e) => setFormData({...formData, destination: e.target.value})}
              disabled={isLoading}
            />
            {errors.destination && (
              <span className="firewall-rule-form-error">{errors.destination}</span>
            )}
          </div>

          <div className="firewall-rule-modal-actions">
            <button 
              type="button" 
              className="firewall-rule-btn firewall-rule-btn-cancel"
              onClick={onClose}
              disabled={isLoading}
            >
              Cancel
            </button>
            <button 
              type="submit" 
              className="firewall-rule-btn firewall-rule-btn-submit"
              disabled={isLoading}
            >
              <FaPlus className="firewall-rule-btn-icon" />
              {isLoading ? 'Adding...' : 'Add Rule'}
            </button>
          </div>
        </form>
      </div>
    </div>,
    document.body
  );
};

export default FirewallRuleModal;
