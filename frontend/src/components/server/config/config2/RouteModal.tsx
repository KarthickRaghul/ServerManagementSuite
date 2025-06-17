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

  const validateForm = () => {
    const newErrors: {[key: string]: string} = {};

    if (!formData.destination.trim()) {
      newErrors.destination = 'Destination is required';
    } else if (!/^(\d{1,3}\.){3}\d{1,3}\/\d{1,2}$/.test(formData.destination.trim())) {
      newErrors.destination = 'Please enter a valid CIDR notation (e.g., 192.168.1.0/24)';
    }

    if (!formData.gateway.trim()) {
      newErrors.gateway = 'Gateway is required';
    } else if (!/^(\d{1,3}\.){3}\d{1,3}$/.test(formData.gateway.trim())) {
      newErrors.gateway = 'Please enter a valid IP address';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) return;

    const routeData: { 
      action: string; 
      destination: string; 
      gateway: string; 
      interface?: string; 
      metric?: string; 
    } = {
      action: 'add',
      destination: formData.destination.trim(),
      gateway: formData.gateway.trim()
    };

    if (formData.interface.trim()) {
      routeData.interface = formData.interface.trim();
    }

    if (formData.metric.trim()) {
      routeData.metric = formData.metric.trim();
    }

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

  // Render modal using Portal to document.body
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
              placeholder="e.g., 192.168.1.0/24"
              value={formData.destination}
              onChange={(e) => setFormData({...formData, destination: e.target.value})}
              disabled={isLoading}
            />
            {errors.destination && (
              <span className="route-form-error">{errors.destination}</span>
            )}
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
              placeholder="e.g., eth0"
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
              className="route-form-input"
              placeholder="e.g., 100"
              value={formData.metric}
              onChange={(e) => setFormData({...formData, metric: e.target.value})}
              disabled={isLoading}
            />
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
    document.body // Render to document.body instead of parent container
  );
};

export default RouteModal;
