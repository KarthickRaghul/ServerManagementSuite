// components/server/config/config1/components/securitymanagement.tsx
import { useState } from "react";
import "./securitymanagement.css";
import ModalWrapper from "./modalwrapper";
import { FaShieldAlt, FaKey, FaLock, FaUpload } from "react-icons/fa";
import { useSecurityManagement } from "../../../../../hooks";
import { useNotification } from "../../../../../context/NotificationContext";

const SecurityManagement = () => {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [sshKey, setSshKey] = useState("");
  const [credentials, setCredentials] = useState({
    username: "",
    currentPassword: "",
    newPassword: "",
    confirmPassword: ""
  });
  const [submitError, setSubmitError] = useState<string | null>(null);

  const { uploadingSSH, updatingPassword, error, uploadSSHKey, updatePassword } = useSecurityManagement();
  const { addNotification } = useNotification();

  const handleSshKeyUpload = async () => {
    if (!sshKey.trim()) return;
    
    setSubmitError(null);

    try {
      const success = await uploadSSHKey(sshKey);
      
      if (success) {
        addNotification({
          title: 'SSH Key Uploaded',
          message: 'SSH public key has been successfully uploaded and configured',
          type: 'success',
          duration: 4000
        });
        setSshKey("");
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to upload SSH key';
      setSubmitError(errorMessage);
      addNotification({
        title: 'SSH Upload Failed',
        message: errorMessage,
        type: 'error',
        duration: 5000
      });
    }
  };

  const handleCredentialUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitError(null);

    if (credentials.newPassword !== credentials.confirmPassword) {
      const errorMessage = "New password and confirmation password do not match!";
      setSubmitError(errorMessage);
      addNotification({
        title: 'Password Mismatch',
        message: errorMessage,
        type: 'error',
        duration: 4000
      });
      return;
    }

    if (credentials.newPassword.length < 6) {
      const errorMessage = "New password must be at least 6 characters long";
      setSubmitError(errorMessage);
      addNotification({
        title: 'Password Too Short',
        message: errorMessage,
        type: 'error',
        duration: 4000
      });
      return;
    }

    try {
      const success = await updatePassword(credentials.username, credentials.newPassword);
      
      if (success) {
        addNotification({
          title: 'Password Updated',
          message: `Password for user "${credentials.username}" has been successfully updated`,
          type: 'success',
          duration: 4000
        });
        setCredentials({
          username: "",
          currentPassword: "",
          newPassword: "",
          confirmPassword: ""
        });
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to update password';
      setSubmitError(errorMessage);
      addNotification({
        title: 'Password Update Failed',
        message: errorMessage,
        type: 'error',
        duration: 5000
      });
    }
  };

  const handleClose = () => {
    setIsModalOpen(false);
    setSshKey("");
    setCredentials({
      username: "",
      currentPassword: "",
      newPassword: "",
      confirmPassword: ""
    });
    setSubmitError(null);
  };

  const handleOpenModal = () => {
    setSubmitError(null);
    setIsModalOpen(true);
  };

  const validateSSHKey = (key: string): boolean => {
    const sshKeyPattern = /^(ssh-rsa|ssh-dss|ssh-ed25519|ecdsa-sha2-nistp256|ecdsa-sha2-nistp384|ecdsa-sha2-nistp521)\s+[A-Za-z0-9+/]+[=]{0,3}(\s+.*)?$/;
    return sshKeyPattern.test(key.trim());
  };

  const isSSHKeyValid = validateSSHKey(sshKey);

  return (
    <>
      <div className="config1-security-card-container" onClick={handleOpenModal}>
        <div className="config1-security-icon-wrapper">
          <FaShieldAlt size={20} color="white" />
        </div>
        <h3>Security Management</h3>
        <p>Manage SSH keys, passwords, and authentication credentials</p>
      </div>

      {isModalOpen && (
        <ModalWrapper
          title="Security Management"
          onClose={handleClose}
        >
          <div className="config1-security-modal-content">
            <p className="config1-security-subtitle">
              Manage authentication credentials and security access controls
            </p>

            {/* Show loading error */}
            {error && (
              <div className="config1-security-error-banner">
                <p>Error: {error}</p>
              </div>
            )}

            {/* Show submit error */}
            {submitError && (
              <div className="config1-security-error-banner">
                <p>{submitError}</p>
              </div>
            )}

            {/* SSH Key Management Section */}
            <div className="config1-security-section">
              <label className="config1-security-label">
                <FaKey className="config1-security-label-icon" />
                SSH Public Key Management
              </label>
              <div className="config1-security-ssh-wrapper">
                <textarea
                  className={`config1-security-textarea ${sshKey && !isSSHKeyValid ? 'config1-security-textarea-error' : ''}`}
                  placeholder="Paste SSH public key here...
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAB... user@hostname"
                  value={sshKey}
                  onChange={(e) => setSshKey(e.target.value)}
                  rows={4}
                  disabled={uploadingSSH}
                />
                {sshKey && !isSSHKeyValid && (
                  <div className="config1-security-validation-error">
                    Please enter a valid SSH public key
                  </div>
                )}
                <button
                  type="button"
                  className="config1-security-btn config1-security-btn-upload"
                  onClick={handleSshKeyUpload}
                  disabled={!sshKey.trim() || !isSSHKeyValid || uploadingSSH}
                >
                  <FaUpload className="config1-security-btn-icon" />
                  {uploadingSSH ? 'Uploading...' : 'Upload SSH Key'}
                </button>
              </div>
            </div>

            {/* Password Management Section */}
            <div className="config1-security-section">
              <label className="config1-security-label">
                <FaLock className="config1-security-label-icon" />
                Password Management
              </label>
              <form className="config1-security-form" onSubmit={handleCredentialUpdate}>
                <div className="config1-security-input-group">
                  <input
                    type="text"
                    className="config1-security-input"
                    placeholder="Enter username"
                    value={credentials.username}
                    onChange={(e) => setCredentials({...credentials, username: e.target.value})}
                    required
                    disabled={updatingPassword}
                  />
                </div>
                
                
                <div className="config1-security-input-group">
                  <input
                    type="password"
                    className="config1-security-input"
                    placeholder="New password (min. 6 characters)"
                    value={credentials.newPassword}
                    onChange={(e) => setCredentials({...credentials, newPassword: e.target.value})}
                    required
                    disabled={updatingPassword}
                    minLength={6}
                  />
                </div>
                
                <div className="config1-security-input-group">
                  <input
                    type="password"
                    className={`config1-security-input ${
                      credentials.newPassword && credentials.confirmPassword && 
                      credentials.newPassword !== credentials.confirmPassword ? 
                      'config1-security-input-error' : ''
                    }`}
                    placeholder="Confirm new password"
                    value={credentials.confirmPassword}
                    onChange={(e) => setCredentials({...credentials, confirmPassword: e.target.value})}
                    required
                    disabled={updatingPassword}
                  />
                  {credentials.newPassword && credentials.confirmPassword && 
                   credentials.newPassword !== credentials.confirmPassword && (
                    <div className="config1-security-validation-error">
                      Passwords do not match
                    </div>
                  )}
                </div>

                <button
                  type="submit"
                  className="config1-security-btn config1-security-btn-update"
                  disabled={
                    !credentials.username || 
                    !credentials.newPassword || 
                    credentials.newPassword !== credentials.confirmPassword ||
                    updatingPassword
                  }
                >
                  <FaLock className="config1-security-btn-icon" />
                  {updatingPassword ? 'Updating...' : 'Update Password'}
                </button>
              </form>
            </div>

            <div className="config1-security-actions">
              <button 
                type="button" 
                className="config1-security-btn config1-security-btn-secondary"
                onClick={handleClose}
                disabled={uploadingSSH || updatingPassword}
              >
                Close
              </button>
            </div>
          </div>
        </ModalWrapper>
      )}
    </>
  );
};

export default SecurityManagement;
