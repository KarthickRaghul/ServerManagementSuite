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
    rule?: string;
    protocol?: string;
    port?: string;
    source?: string;
    destination?: string;
    name?: string;
    displayName?: string;
    direction?: string;
    actionType?: string;
    enabled?: string;
    profile?: string;
    localPort?: string;
    remotePort?: string;
    localAddress?: string;
    remoteAddress?: string;
    program?: string;
    service?: string;
  }) => Promise<boolean>;
  isLoading: boolean;
  firewallType?: 'linux' | 'windows';
}

const FirewallRuleModal: React.FC<FirewallRuleModalProps> = ({
  isOpen,
  onClose,
  onAddRule,
  isLoading,
  firewallType = 'linux'
}) => {
  // Linux form data
  const [linuxFormData, setLinuxFormData] = useState({
    rule: 'accept',
    protocol: 'tcp',
    port: '',
    source: '',
    destination: ''
  });

  // Windows form data
  const [windowsFormData, setWindowsFormData] = useState({
    name: '',
    displayName: '',
    direction: 'Inbound',
    action: 'Allow',
    enabled: 'True',
    profile: 'Public',
    protocol: 'TCP',
    localPort: '',
    remotePort: '',
    localAddress: '',
    remoteAddress: '',
    program: '',
    service: ''
  });

  const [errors, setErrors] = useState<{[key: string]: string}>({});

  if (!isOpen) return null;

  const validateLinuxForm = () => {
    const newErrors: {[key: string]: string} = {};

    if (!linuxFormData.port.trim()) {
      newErrors.port = 'Port is required';
    } else if (!/^\d+(-\d+)?$/.test(linuxFormData.port.trim())) {
      newErrors.port = 'Please enter a valid port number or range (e.g., 80 or 8080-8090)';
    }

    if (linuxFormData.source.trim() && !/^(\d{1,3}\.){3}\d{1,3}(\/\d{1,2})?$/.test(linuxFormData.source.trim())) {
      newErrors.source = 'Please enter a valid IP address or CIDR notation';
    }

    if (linuxFormData.destination.trim() && !/^(\d{1,3}\.){3}\d{1,3}(\/\d{1,2})?$/.test(linuxFormData.destination.trim())) {
      newErrors.destination = 'Please enter a valid IP address or CIDR notation';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const validateWindowsForm = () => {
    const newErrors: {[key: string]: string} = {};

    if (!windowsFormData.name.trim()) {
      newErrors.name = 'Rule name is required';
    }

    if (!windowsFormData.displayName.trim()) {
      newErrors.displayName = 'Display name is required';
    }

    // Validate ports if provided
    if (windowsFormData.localPort.trim() && !/^\d+(-\d+)?(,\d+(-\d+)?)*$/.test(windowsFormData.localPort.trim())) {
      newErrors.localPort = 'Please enter valid port numbers (e.g., 80, 80-90, 80,443)';
    }

    if (windowsFormData.remotePort.trim() && !/^\d+(-\d+)?(,\d+(-\d+)?)*$/.test(windowsFormData.remotePort.trim())) {
      newErrors.remotePort = 'Please enter valid port numbers (e.g., 80, 80-90, 80,443)';
    }

    // Validate IP addresses if provided
    if (windowsFormData.localAddress.trim() && !/^(\d{1,3}\.){3}\d{1,3}(\/\d{1,2})?(,(\d{1,3}\.){3}\d{1,3}(\/\d{1,2})?)*$/.test(windowsFormData.localAddress.trim())) {
      newErrors.localAddress = 'Please enter valid IP addresses (e.g., 192.168.1.1, 192.168.1.0/24)';
    }

    if (windowsFormData.remoteAddress.trim() && !/^(\d{1,3}\.){3}\d{1,3}(\/\d{1,2})?(,(\d{1,3}\.){3}\d{1,3}(\/\d{1,2})?)*$/.test(windowsFormData.remoteAddress.trim())) {
      newErrors.remoteAddress = 'Please enter valid IP addresses (e.g., 192.168.1.1, 192.168.1.0/24)';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    let ruleData: {
      action: string;
      rule?: string;
      protocol?: string;
      port?: string;
      source?: string;
      destination?: string;
      name?: string;
      displayName?: string;
      direction?: string;
      actionType?: string;
      enabled?: string;
      profile?: string;
      localPort?: string;
      remotePort?: string;
      localAddress?: string;
      remoteAddress?: string;
      program?: string;
      service?: string;
    } = { action: '' };
    let isValid = false;

    if (firewallType === 'linux') {
      isValid = validateLinuxForm();
      if (isValid) {
        ruleData = {
          action: 'add',
          rule: linuxFormData.rule,
          protocol: linuxFormData.protocol,
          port: linuxFormData.port.trim()
        };

        if (linuxFormData.source.trim()) {
          ruleData.source = linuxFormData.source.trim();
        }

        if (linuxFormData.destination.trim()) {
          ruleData.destination = linuxFormData.destination.trim();
        }
      }
    } else {
      isValid = validateWindowsForm();
      if (isValid) {
        ruleData = {
          action: 'add',
          name: windowsFormData.name.trim(),
          displayName: windowsFormData.displayName.trim(),
          direction: windowsFormData.direction,
          actionType: windowsFormData.action,
          enabled: windowsFormData.enabled,
          profile: windowsFormData.profile,
          protocol: windowsFormData.protocol
        };

        // Add optional fields if provided
        if (windowsFormData.localPort.trim()) {
          ruleData.localPort = windowsFormData.localPort.trim();
        }
        if (windowsFormData.remotePort.trim()) {
          ruleData.remotePort = windowsFormData.remotePort.trim();
        }
        if (windowsFormData.localAddress.trim()) {
          ruleData.localAddress = windowsFormData.localAddress.trim();
        }
        if (windowsFormData.remoteAddress.trim()) {
          ruleData.remoteAddress = windowsFormData.remoteAddress.trim();
        }
        if (windowsFormData.program.trim()) {
          ruleData.program = windowsFormData.program.trim();
        }
        if (windowsFormData.service.trim()) {
          ruleData.service = windowsFormData.service.trim();
        }
      }
    }

    if (!isValid) return;

    const success = await onAddRule(ruleData);
    if (success) {
      // Reset forms
      setLinuxFormData({ rule: 'accept', protocol: 'tcp', port: '', source: '', destination: '' });
      setWindowsFormData({
        name: '',
        displayName: '',
        direction: 'Inbound',
        action: 'Allow',
        enabled: 'True',
        profile: 'Public',
        protocol: 'TCP',
        localPort: '',
        remotePort: '',
        localAddress: '',
        remoteAddress: '',
        program: '',
        service: ''
      });
      setErrors({});
      onClose();
    }
  };

  const handleBackdropClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose();
    }
  };

  const renderLinuxForm = () => (
    <>
      <div className="firewall-rule-form-row">
        <div className="firewall-rule-form-group">
          <label className="firewall-rule-form-label">Rule Action *</label>
          <select
            className="firewall-rule-form-select"
            value={linuxFormData.rule}
            onChange={(e) => setLinuxFormData({...linuxFormData, rule: e.target.value})}
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
            value={linuxFormData.protocol}
            onChange={(e) => setLinuxFormData({...linuxFormData, protocol: e.target.value})}
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
          value={linuxFormData.port}
          onChange={(e) => setLinuxFormData({...linuxFormData, port: e.target.value})}
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
          value={linuxFormData.source}
          onChange={(e) => setLinuxFormData({...linuxFormData, source: e.target.value})}
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
          value={linuxFormData.destination}
          onChange={(e) => setLinuxFormData({...linuxFormData, destination: e.target.value})}
          disabled={isLoading}
        />
        {errors.destination && (
          <span className="firewall-rule-form-error">{errors.destination}</span>
        )}
      </div>
    </>
  );

  const renderWindowsForm = () => (
    <>
      <div className="firewall-rule-form-row">
        <div className="firewall-rule-form-group">
          <label className="firewall-rule-form-label">Rule Name *</label>
          <input
            type="text"
            className={`firewall-rule-form-input ${errors.name ? 'error' : ''}`}
            placeholder="e.g., Allow-HTTP-In"
            value={windowsFormData.name}
            onChange={(e) => setWindowsFormData({...windowsFormData, name: e.target.value})}
            disabled={isLoading}
          />
          {errors.name && (
            <span className="firewall-rule-form-error">{errors.name}</span>
          )}
        </div>

        <div className="firewall-rule-form-group">
          <label className="firewall-rule-form-label">Display Name *</label>
          <input
            type="text"
            className={`firewall-rule-form-input ${errors.displayName ? 'error' : ''}`}
            placeholder="e.g., Allow HTTP Traffic (Inbound)"
            value={windowsFormData.displayName}
            onChange={(e) => setWindowsFormData({...windowsFormData, displayName: e.target.value})}
            disabled={isLoading}
          />
          {errors.displayName && (
            <span className="firewall-rule-form-error">{errors.displayName}</span>
          )}
        </div>
      </div>

      <div className="firewall-rule-form-row">
        <div className="firewall-rule-form-group">
          <label className="firewall-rule-form-label">Direction *</label>
          <select
            className="firewall-rule-form-select"
            value={windowsFormData.direction}
            onChange={(e) => setWindowsFormData({...windowsFormData, direction: e.target.value})}
            disabled={isLoading}
          >
            <option value="Inbound">Inbound</option>
            <option value="Outbound">Outbound</option>
          </select>
        </div>

        <div className="firewall-rule-form-group">
          <label className="firewall-rule-form-label">Action *</label>
          <select
            className="firewall-rule-form-select"
            value={windowsFormData.action}
            onChange={(e) => setWindowsFormData({...windowsFormData, action: e.target.value})}
            disabled={isLoading}
          >
            <option value="Allow">Allow</option>
            <option value="Block">Block</option>
          </select>
        </div>
      </div>

      <div className="firewall-rule-form-row">
        <div className="firewall-rule-form-group">
          <label className="firewall-rule-form-label">Protocol *</label>
          <select
            className="firewall-rule-form-select"
            value={windowsFormData.protocol}
            onChange={(e) => setWindowsFormData({...windowsFormData, protocol: e.target.value})}
            disabled={isLoading}
          >
            <option value="TCP">TCP</option>
            <option value="UDP">UDP</option>
            <option value="Any">Any</option>
          </select>
        </div>

        <div className="firewall-rule-form-group">
          <label className="firewall-rule-form-label">Profile *</label>
          <select
            className="firewall-rule-form-select"
            value={windowsFormData.profile}
            onChange={(e) => setWindowsFormData({...windowsFormData, profile: e.target.value})}
            disabled={isLoading}
          >
            <option value="Public">Public</option>
            <option value="Private">Private</option>
            <option value="Domain">Domain</option>
            <option value="Any">Any</option>
          </select>
        </div>
      </div>

      <div className="firewall-rule-form-row">
        <div className="firewall-rule-form-group">
          <label className="firewall-rule-form-label">Local Port</label>
          <input
            type="text"
            className={`firewall-rule-form-input ${errors.localPort ? 'error' : ''}`}
            placeholder="e.g., 80, 80-90, 80,443"
            value={windowsFormData.localPort}
            onChange={(e) => setWindowsFormData({...windowsFormData, localPort: e.target.value})}
            disabled={isLoading}
          />
          {errors.localPort && (
            <span className="firewall-rule-form-error">{errors.localPort}</span>
          )}
        </div>

        <div className="firewall-rule-form-group">
          <label className="firewall-rule-form-label">Remote Port</label>
          <input
            type="text"
            className={`firewall-rule-form-input ${errors.remotePort ? 'error' : ''}`}
            placeholder="e.g., 80, 80-90, 80,443"
            value={windowsFormData.remotePort}
            onChange={(e) => setWindowsFormData({...windowsFormData, remotePort: e.target.value})}
            disabled={isLoading}
          />
          {errors.remotePort && (
            <span className="firewall-rule-form-error">{errors.remotePort}</span>
          )}
        </div>
      </div>

      <div className="firewall-rule-form-row">
        <div className="firewall-rule-form-group">
          <label className="firewall-rule-form-label">Local Address</label>
          <input
            type="text"
            className={`firewall-rule-form-input ${errors.localAddress ? 'error' : ''}`}
            placeholder="e.g., 192.168.1.1, 192.168.1.0/24"
            value={windowsFormData.localAddress}
            onChange={(e) => setWindowsFormData({...windowsFormData, localAddress: e.target.value})}
            disabled={isLoading}
          />
          {errors.localAddress && (
            <span className="firewall-rule-form-error">{errors.localAddress}</span>
          )}
        </div>

        <div className="firewall-rule-form-group">
          <label className="firewall-rule-form-label">Remote Address</label>
          <input
            type="text"
            className={`firewall-rule-form-input ${errors.remoteAddress ? 'error' : ''}`}
            placeholder="e.g., 192.168.1.1, 192.168.1.0/24"
            value={windowsFormData.remoteAddress}
            onChange={(e) => setWindowsFormData({...windowsFormData, remoteAddress: e.target.value})}
            disabled={isLoading}
          />
          {errors.remoteAddress && (
            <span className="firewall-rule-form-error">{errors.remoteAddress}</span>
          )}
        </div>
      </div>

      <div className="firewall-rule-form-row">
        <div className="firewall-rule-form-group">
          <label className="firewall-rule-form-label">Program Path</label>
          <input
            type="text"
            className="firewall-rule-form-input"
            placeholder="e.g., C:\\Program Files\\MyApp\\app.exe"
            value={windowsFormData.program}
            onChange={(e) => setWindowsFormData({...windowsFormData, program: e.target.value})}
            disabled={isLoading}
          />
        </div>

        <div className="firewall-rule-form-group">
          <label className="firewall-rule-form-label">Service Name</label>
          <input
            type="text"
            className="firewall-rule-form-input"
            placeholder="e.g., Spooler, BITS"
            value={windowsFormData.service}
            onChange={(e) => setWindowsFormData({...windowsFormData, service: e.target.value})}
            disabled={isLoading}
          />
        </div>
      </div>

      <div className="firewall-rule-form-group">
        <label className="firewall-rule-form-label">Enabled *</label>
        <select
          className="firewall-rule-form-select"
          value={windowsFormData.enabled}
          onChange={(e) => setWindowsFormData({...windowsFormData, enabled: e.target.value})}
          disabled={isLoading}
        >
          <option value="True">Enabled</option>
          <option value="False">Disabled</option>
        </select>
      </div>
    </>
  );

  return createPortal(
    <div className="firewall-rule-modal-overlay" onClick={handleBackdropClick}>
      <div className="firewall-rule-modal-container">
        <div className="firewall-rule-modal-header">
          <div className="firewall-rule-modal-title-section">
            <div className="firewall-rule-modal-icon-wrapper">
              <FaShieldAlt className="firewall-rule-modal-icon" />
            </div>
            <h2 className="firewall-rule-modal-title">
              Add {firewallType === 'linux' ? 'iptables' : 'Windows Firewall'} Rule
            </h2>
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
          {firewallType === 'linux' ? renderLinuxForm() : renderWindowsForm()}

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
