// components/server/config/config1/components/serverconfiguration.tsx
import React, { useState, useEffect } from 'react';
import './serverconfiguration.css';
import { FaServer, FaCog, FaGlobe, FaInfoCircle } from 'react-icons/fa';
import ModalWrapper from './modalwrapper';
import { useServerConfiguration } from '../../../../../hooks';
import { useNotification } from '../../../../../context/NotificationContext';

const ServerConfiguration: React.FC = () => {
  const [showModal, setShowModal] = useState(false);
  const [formData, setFormData] = useState({
    hostname: '',
    timezone: 'UTC'
  });
  const [submitError, setSubmitError] = useState<string | null>(null);

  const { data, loading, error, updating, updateConfiguration } = useServerConfiguration();
  const { addNotification } = useNotification();

  // Update form data when server data is loaded
  useEffect(() => {
    if (data) {
      setFormData({
        hostname: data.hostname,
        timezone: data.timezone
      });
    }
  }, [data]);

  const handleApply = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitError(null);
    
    // Validate hostname
    if (formData.hostname.trim().length > 15) {
      setSubmitError('Hostname must be 15 characters or less');
      return;
    }

    if (!/^[a-zA-Z0-9-]+$/.test(formData.hostname.trim())) {
      setSubmitError('Hostname can only contain letters, numbers, and hyphens');
      return;
    }
    
    try {
      const success = await updateConfiguration({
        hostname: formData.hostname.trim(),
        timezone: formData.timezone
      });

      if (success) {
        addNotification({
          title: 'Configuration Updated',
          message: 'Server hostname and timezone have been successfully updated. A restart may be required for hostname changes to take effect.',
          type: 'success',
          duration: 6000
        });
        setSubmitError(null);
        console.log("Configuration updated successfully");
      } else {
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
    setSubmitError(null);
    // Reset form data to current server data
    if (data) {
      setFormData({
        hostname: data.hostname,
        timezone: data.timezone
      });
    }
  };

  const handleOpenModal = () => {
    setSubmitError(null);
    if (data) {
      setFormData({
        hostname: data.hostname,
        timezone: data.timezone
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
            
            {error && (
              <div className="config1-serverconfig-error-banner">
                <p>Error loading data: {error}</p>
              </div>
            )}

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
                  placeholder="server01" 
                  value={formData.hostname}
                  onChange={(e) => setFormData({...formData, hostname: e.target.value})}
                  required 
                  disabled={updating}
                  maxLength={15}
                  pattern="[a-zA-Z0-9-]+"
                />
                <div className="config1-serverconfig-help">
                  <FaInfoCircle className="config1-serverconfig-help-icon" />
                  <small>Maximum 15 characters. Only letters, numbers, and hyphens allowed.</small>
                </div>
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
                  <optgroup label="UTC/GMT">
                    <option value="UTC">UTC</option>
                    <option value="GMT">GMT (Greenwich Mean Time)</option>
                  </optgroup>
                  
                  <optgroup label="Americas">
                    <option value="America/New_York">Eastern Time (New York)</option>
                    <option value="America/Chicago">Central Time (Chicago)</option>
                    <option value="America/Denver">Mountain Time (Denver)</option>
                    <option value="America/Los_Angeles">Pacific Time (Los Angeles)</option>
                    <option value="America/Phoenix">Arizona Time (Phoenix)</option>
                    <option value="America/Anchorage">Alaska Time (Anchorage)</option>
                    <option value="America/Toronto">Eastern Time (Toronto)</option>
                    <option value="America/Vancouver">Pacific Time (Vancouver)</option>
                    <option value="America/Mexico_City">Central Time (Mexico City)</option>
                    <option value="America/Sao_Paulo">Brazil Time (São Paulo)</option>
                    <option value="America/Buenos_Aires">Argentina Time (Buenos Aires)</option>
                  </optgroup>
                  
                  <optgroup label="Europe">
                    <option value="Europe/London">London (GMT/BST)</option>
                    <option value="Europe/Berlin">Berlin (CET/CEST)</option>
                    <option value="Europe/Paris">Paris (CET/CEST)</option>
                    <option value="Europe/Rome">Rome (CET/CEST)</option>
                    <option value="Europe/Madrid">Madrid (CET/CEST)</option>
                    <option value="Europe/Amsterdam">Amsterdam (CET/CEST)</option>
                    <option value="Europe/Vienna">Vienna (CET/CEST)</option>
                    <option value="Europe/Zurich">Zurich (CET/CEST)</option>
                    <option value="Europe/Stockholm">Stockholm (CET/CEST)</option>
                    <option value="Europe/Moscow">Moscow (MSK)</option>
                  </optgroup>
                  
                  <optgroup label="Asia">
                    <option value="Asia/Kolkata">India (IST)</option>
                    <option value="Asia/Tokyo">Tokyo (JST)</option>
                    <option value="Asia/Shanghai">Shanghai (CST)</option>
                    <option value="Asia/Hong_Kong">Hong Kong (HKT)</option>
                    <option value="Asia/Singapore">Singapore (SGT)</option>
                    <option value="Asia/Seoul">Seoul (KST)</option>
                    <option value="Asia/Bangkok">Bangkok (ICT)</option>
                    <option value="Asia/Dubai">Dubai (GST)</option>
                    <option value="Asia/Tehran">Tehran (IRST)</option>
                    <option value="Asia/Karachi">Karachi (PKT)</option>
                  </optgroup>
                  
                  <optgroup label="Australia & Pacific">
                    <option value="Australia/Sydney">Sydney (AEDT)</option>
                    <option value="Australia/Melbourne">Melbourne (AEDT)</option>
                    <option value="Australia/Brisbane">Brisbane (AEST)</option>
                    <option value="Australia/Perth">Perth (AWST)</option>
                    <option value="Pacific/Auckland">Auckland (NZDT)</option>
                    <option value="Pacific/Honolulu">Honolulu (HST)</option>
                  </optgroup>
                  
                  <optgroup label="Africa">
                    <option value="Africa/Cairo">Cairo (EET)</option>
                    <option value="Africa/Johannesburg">Johannesburg (SAST)</option>
                    <option value="Africa/Lagos">Lagos (WAT)</option>
                    <option value="Africa/Nairobi">Nairobi (EAT)</option>
                  </optgroup>
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
