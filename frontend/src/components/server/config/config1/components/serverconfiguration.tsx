// components/server/config/config1/components/serverconfiguration.tsx
import React, { useState, useEffect } from 'react';
import './serverconfiguration.css';
import { FaServer, FaCog, FaGlobe } from 'react-icons/fa';
import ModalWrapper from './modalwrapper';
import { useServerConfiguration } from '../../../../../hooks';
import { useNotification } from '../../../../../context/NotificationContext';

const ServerConfiguration: React.FC = () => {
  const [showModal, setShowModal] = useState(false);
  const [formData, setFormData] = useState({
    hostname: '',
    timezone: 'UTC+0 (GMT)'
  });
  const [submitError, setSubmitError] = useState<string | null>(null);

  const { data, loading, error, updating, updateConfiguration } = useServerConfiguration();
  const { addNotification } = useNotification();

  // Update form data when server data is loaded
  useEffect(() => {
    if (data) {
      setFormData({
        hostname: data.hostname,
        timezone: mapTimezoneToDisplay(data.timezone)
      });
    }
  }, [data]);

  // Map backend timezone to display format
  const mapTimezoneToDisplay = (backendTimezone: string): string => {
    const timezoneMap: { [key: string]: string } = {
      'Asia/Kolkata': 'UTC+5:30 (IST)',
      'UTC': 'UTC+0 (GMT)',
      'America/Los_Angeles': 'UTC-8 (PST)',
      'Europe/Berlin': 'UTC+1 (CET)',
      'America/New_York': 'UTC-5 (EST)',
      'Asia/Tokyo': 'UTC+9 (JST)'
    };
    return timezoneMap[backendTimezone] || 'UTC+0 (GMT)';
  };

  // Map display format to backend timezone
  const mapDisplayToTimezone = (displayTimezone: string): string => {
    const displayMap: { [key: string]: string } = {
      'UTC+5:30 (IST)': 'Asia/Kolkata',
      'UTC+0 (GMT)': 'UTC',
      'UTC-8 (PST)': 'America/Los_Angeles',
      'UTC+1 (CET)': 'Europe/Berlin',
      'UTC-5 (EST)': 'America/New_York',
      'UTC+9 (JST)': 'Asia/Tokyo'
    };
    return displayMap[displayTimezone] || 'UTC';
  };

  const handleApply = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitError(null); // Clear any previous errors
    
    try {
      const success = await updateConfiguration({
        hostname: formData.hostname.trim(),
        timezone: mapDisplayToTimezone(formData.timezone)
      });

      if (success) {
        // Show success notification but DON'T close modal
        addNotification({
          title: 'Configuration Updated',
          message: 'Server hostname and timezone have been successfully updated.',
          type: 'success',
          duration: 4000
        });
        setSubmitError(null);
        console.log("Configuration updated successfully");
        // Modal stays open - user must manually close it
      } else {
        // If update failed, show error notification and banner
        const errorMessage = "Failed to update configuration. Please try again.";
        setSubmitError(errorMessage);
        addNotification({
          title: 'Update Failed',
          message: errorMessage,
          type: 'error',
          duration: 5000
        });
      }
    } catch (err) {
      console.error("Failed to update configuration:", err);
      const errorMessage = err instanceof Error ? err.message : "Failed to update configuration. Please try again.";
      setSubmitError(errorMessage);
      addNotification({
        title: 'Update Error',
        message: errorMessage,
        type: 'error',
        duration: 5000
      });
    }
  };

  const handleClose = () => {
    setShowModal(false);
    setSubmitError(null); // Clear submit errors when closing
    // Reset form data to current server data
    if (data) {
      setFormData({
        hostname: data.hostname,
        timezone: mapTimezoneToDisplay(data.timezone)
      });
    }
  };

  const handleOpenModal = () => {
    setSubmitError(null); // Clear any previous submit errors
    if (data) {
      setFormData({
        hostname: data.hostname,
        timezone: mapTimezoneToDisplay(data.timezone)
      });
    }
    setShowModal(true);
  };

  return (
    <>
      <div className="config1-serverconfig-card-container" onClick={handleOpenModal}>
        <div className="config1-serverconfig-icon-wrapper">
          <FaServer size={20} color="white" />
        </div>
        <h3>System Configuration</h3>
        <p>Configure server hostname and timezone settings</p>
        {loading && <div className="config1-serverconfig-loading">Loading...</div>}
        {error && <div className="config1-serverconfig-error">Error loading config</div>}
      </div>

      {showModal && (
        <ModalWrapper title="System Configuration" onClose={handleClose}>
          <div className="config1-serverconfig-modal-content">
            <p className="config1-serverconfig-subtitle">
              Configure fundamental server parameters and system-level settings
            </p>
            
            {/* Show loading error */}
            {error && (
              <div className="config1-serverconfig-error-banner">
                <p>Error loading data: {error}</p>
              </div>
            )}

            {/* Show submit error */}
            {submitError && (
              <div className="config1-serverconfig-error-banner">
                <p>{submitError}</p>
              </div>
            )}
            
            <form className="config1-serverconfig-form" onSubmit={handleApply}>
              <div className="config1-serverconfig-input-group">
                <label className="config1-serverconfig-label">
                  <FaCog className="config1-serverconfig-label-icon" />
                  Server Hostname
                </label>
                <input 
                  className="config1-serverconfig-input"
                  name="hostname" 
                  placeholder="server.company.com" 
                  value={formData.hostname}
                  onChange={(e) => setFormData({...formData, hostname: e.target.value})}
                  required 
                  disabled={updating}
                />
              </div>
              
              <div className="config1-serverconfig-input-group">
                <label className="config1-serverconfig-label">
                  <FaGlobe className="config1-serverconfig-label-icon" />
                  System Timezone
                </label>
                <select 
                  className="config1-serverconfig-select"
                  name="timezone"
                  value={formData.timezone}
                  onChange={(e) => setFormData({...formData, timezone: e.target.value})}
                  disabled={updating}
                >
                  <option value="UTC+0 (GMT)">UTC+0 (GMT)</option>
                  <option value="UTC+5:30 (IST)">UTC+5:30 (IST)</option>
                  <option value="UTC-8 (PST)">UTC-8 (PST)</option>
                  <option value="UTC+1 (CET)">UTC+1 (CET)</option>
                  <option value="UTC-5 (EST)">UTC-5 (EST)</option>
                  <option value="UTC+9 (JST)">UTC+9 (JST)</option>
                </select>
              </div>

              <div className="config1-serverconfig-actions">
                <button 
                  type="button" 
                  className="config1-serverconfig-btn config1-serverconfig-btn-secondary"
                  onClick={handleClose}
                  disabled={updating}
                >
                  Close
                </button>
                <button 
                  type="submit" 
                  className="config1-serverconfig-btn config1-serverconfig-btn-primary"
                  disabled={!formData.hostname.trim() || updating}
                >
                  <FaServer className="config1-serverconfig-btn-icon" />
                  {updating ? 'Applying...' : 'Apply Configuration'}
                </button>
              </div>
            </form>
          </div>
        </ModalWrapper>
      )}
    </>
  );
};

export default ServerConfiguration;
