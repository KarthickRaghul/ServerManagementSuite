// components/server/config2/RouteModal.tsx
import React, { useState } from 'react';
import { createPortal } from 'react-dom';
import { FaTimes, FaRoute, FaPlus } from 'react-icons/fa';
import './RouteModal.css';

interface RouteModalProps {
  isOpen: boolean;
  onClose: () => void;
  onAddRoute: (routeData: {
    action: string;
    destination: string;
    gateway: string;
    interface?: string;
    metric?: string;
  }) => Promise<boolean>;
  isLoading: boolean;
}

const RouteModal: React.FC<RouteModalProps> = ({
  isOpen,
  onClose,
  onAddRoute,
  isLoading
}) => {
  const [formData, setFormData] = useState({
    destination: '',
    gateway: '',
    interface: '',
    metric: ''
  });
  const [errors, setErrors] = useState<{[key: string]: string}>({});

  if (!isOpen) return null;

  // ✅ Enhanced validation with CIDR auto-correction
  const validateForm = () => {
    const newErrors: {[key: string]: string} = {};

    if (!formData.destination.trim()) {
      newErrors.destination = 'Destination is required';
    } else {
      const dest = formData.destination.trim();
      
      // Check if it's a valid IP or CIDR
      if (dest !== "default" && dest !== "0.0.0.0") {
        // If it contains /, validate as CIDR
        if (dest.includes('/')) {
          const cidrRegex = /^(\d{1,3}\.){3}\d{1,3}\/(\d{1,2})$/;
          if (!cidrRegex.test(dest)) {
            newErrors.destination = 'Please enter a valid CIDR notation (e.g., 192.168.1.0/24)';
          } else {
            // Validate CIDR prefix length
            const prefix = parseInt(dest.split('/')[1]);
            if (prefix < 0 || prefix > 32) {
              newErrors.destination = 'CIDR prefix must be between 0 and 32';
            }
          }
        } else {
          // If no /, validate as IP and suggest CIDR
          const ipRegex = /^(\d{1,3}\.){3}\d{1,3}$/;
          if (!ipRegex.test(dest)) {
            newErrors.destination = 'Please enter a valid IP address or CIDR notation';
          }
        }
      }
    }

    if (!formData.gateway.trim()) {
      newErrors.gateway = 'Gateway is required';
    } else if (!/^(\d{1,3}\.){3}\d{1,3}$/.test(formData.gateway.trim())) {
      newErrors.gateway = 'Please enter a valid IP address';
    }

    // Validate metric if provided
    if (formData.metric.trim() && !/^\d+$/.test(formData.metric.trim())) {
      newErrors.metric = 'Metric must be a number';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // ✅ Auto-format destination to include CIDR if missing
  const formatDestination = (destination: string): string => {
    const dest = destination.trim();
    
    // Handle special cases
    if (dest === "default") return "0.0.0.0/0";
    if (dest === "0.0.0.0") return "0.0.0.0/0";
    
    // If already has CIDR, return as is
    if (dest.includes('/')) return dest;
    
    // Auto-add CIDR based on common patterns
    if (dest.endsWith('.0')) {
      // Looks like a network address, add /24
      return dest + '/24';
    } else {
      // Looks like a host address, add /32
      return dest + '/32';
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) return;

    // ✅ Format destination with proper CIDR notation
    const formattedDestination = formatDestination(formData.destination);

    const routeData: { 
      action: string; 
      destination: string; 
      gateway: string; 
      interface?: string; 
      metric?: string; 
    } = {
      action: 'add',
      destination: formattedDestination,
      gateway: formData.gateway.trim()
    };

    if (formData.interface.trim()) {
      routeData.interface = formData.interface.trim();
    }

    if (formData.metric.trim()) {
      routeData.metric = formData.metric.trim();
    }

    console.log('🔍 Submitting route data:', routeData);

    const success = await onAddRoute(routeData);
    if (success) {
      setFormData({ destination: '', gateway: '', interface: '', metric: '' });
      setErrors({});
      onClose();
    }
  };

  const handleBackdropClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose();
    }
  };

  // ✅ Helper function to show CIDR examples
  const getDestinationPlaceholder = () => {
    return "e.g., 192.168.1.0/24 or default";
  };

  const getDestinationHelp = () => {
    if (formData.destination && !formData.destination.includes('/') && formData.destination !== 'default') {
      const formatted = formatDestination(formData.destination);
      return `Will be formatted as: ${formatted}`;
    }
    return "Enter network address with CIDR notation";
  };

  return createPortal(
    <div className="route-modal-overlay" onClick={handleBackdropClick}>
      <div className="route-modal-container">
        <div className="route-modal-header">
          <div className="route-modal-title-section">
            <div className="route-modal-icon-wrapper">
              <FaRoute className="route-modal-icon" />
            </div>
            <h2 className="route-modal-title">Add New Route</h2>
          </div>
          <button 
            className="route-modal-close"
            onClick={onClose}
            type="button"
            disabled={isLoading}
          >
            <FaTimes />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="route-modal-form">
          <div className="route-form-group">
            <label className="route-form-label">
              Destination Network *
            </label>
            <input
              type="text"
              className={`route-form-input ${errors.destination ? 'error' : ''}`}
              placeholder={getDestinationPlaceholder()}
              value={formData.destination}
              onChange={(e) => setFormData({...formData, destination: e.target.value})}
              disabled={isLoading}
            />
            {!errors.destination && formData.destination && (
              <span className="route-form-help">{getDestinationHelp()}</span>
            )}
            {errors.destination && (
              <span className="route-form-error">{errors.destination}</span>
            )}
            <div className="route-form-examples">
              <small>Examples: 192.168.1.0/24, 10.0.0.0/8, default, 172.16.1.100/32</small>
            </div>
          </div>

          <div className="route-form-group">
            <label className="route-form-label">
              Gateway *
            </label>
            <input
              type="text"
              className={`route-form-input ${errors.gateway ? 'error' : ''}`}
              placeholder="e.g., 192.168.1.1"
              value={formData.gateway}
              onChange={(e) => setFormData({...formData, gateway: e.target.value})}
              disabled={isLoading}
            />
            {errors.gateway && (
              <span className="route-form-error">{errors.gateway}</span>
            )}
          </div>

          <div className="route-form-group">
            <label className="route-form-label">
              Interface (Optional)
            </label>
            <input
              type="text"
              className="route-form-input"
              placeholder="e.g., eth0, enp3s0"
              value={formData.interface}
              onChange={(e) => setFormData({...formData, interface: e.target.value})}
              disabled={isLoading}
            />
          </div>

          <div className="route-form-group">
            <label className="route-form-label">
              Metric (Optional)
            </label>
            <input
              type="text"
              className={`route-form-input ${errors.metric ? 'error' : ''}`}
              placeholder="e.g., 100"
              value={formData.metric}
              onChange={(e) => setFormData({...formData, metric: e.target.value})}
              disabled={isLoading}
            />
            {errors.metric && (
              <span className="route-form-error">{errors.metric}</span>
            )}
          </div>

          <div className="route-modal-actions">
            <button 
              type="button" 
              className="route-btn route-btn-cancel"
              onClick={onClose}
              disabled={isLoading}
            >
              Cancel
            </button>
            <button 
              type="submit" 
              className="route-btn route-btn-submit"
              disabled={isLoading}
            >
              <FaPlus className="route-btn-icon" />
              {isLoading ? 'Adding...' : 'Add Route'}
            </button>
          </div>
        </form>
      </div>
    </div>,
    document.body
  );
};

export default RouteModal;
