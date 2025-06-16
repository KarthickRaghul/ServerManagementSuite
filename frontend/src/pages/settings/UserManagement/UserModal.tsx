// pages/settings/UserManagement/UserModal.tsx
import React, { useState } from 'react';
import { FaTimes, FaUser, FaEnvelope, FaUserTag, FaLock } from 'react-icons/fa';
import './UserModal.css';

interface Props {
  onSubmit: (user: { username: string; email: string; role: 'admin' | 'viewer'; password: string }) => void;
  onClose: () => void;
  creating: boolean;
  submitError: string | null;
}

const UserModal: React.FC<Props> = ({ onClose, onSubmit, creating, submitError }) => {
  const [username, setUsername] = useState('');
  const [role, setRole] = useState<'admin' | 'viewer'>('viewer');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [errors, setErrors] = useState<{[key: string]: string}>({});

  const validateForm = () => {
    const newErrors: {[key: string]: string} = {};

    if (!username.trim()) {
      newErrors.username = 'Username is required';
    } else if (username.length < 3) {
      newErrors.username = 'Username must be at least 3 characters';
    }

    if (!email.trim()) {
      newErrors.email = 'Email is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      newErrors.email = 'Please enter a valid email address';
    }

    if (!password) {
      newErrors.password = 'Password is required';
    } else if (password.length < 6) {
      newErrors.password = 'Password must be at least 6 characters';
    }

    if (!confirmPassword) {
      newErrors.confirmPassword = 'Please confirm your password';
    } else if (password !== confirmPassword) {
      newErrors.confirmPassword = 'Passwords do not match';
    }

    if (!role) {
      newErrors.role = 'Role is required';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) return;
    
    onSubmit({ 
      username: username.trim(), 
      role, 
      email: email.trim(),
      password 
    });
  };

  const handleClose = () => {
    setUsername('');
    setEmail('');
    setPassword('');
    setConfirmPassword('');
    setRole('viewer');
    setErrors({});
    onClose();
  };

  const handleBackdropClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      handleClose();
    }
  };

  return (
    <div className="settings-usermgmt-modal-overlay" onClick={handleBackdropClick}>
      <div className="settings-usermgmt-modal-container">
        <div className="settings-usermgmt-modal-header">
          <h2 className="settings-usermgmt-modal-title">Add New User</h2>
          <button 
            className="settings-usermgmt-modal-close"
            onClick={handleClose}
            type="button"
            disabled={creating}
          >
            <FaTimes />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="settings-usermgmt-modal-form">
          {/* Show submit error */}
          {submitError && (
            <div className="settings-usermgmt-form-error-banner">
              <p>{submitError}</p>
            </div>
          )}

          <div className="settings-usermgmt-form-group">
            <label className="settings-usermgmt-form-label">
              <FaUser className="settings-usermgmt-form-icon" />
              Username
            </label>
            <input
              type="text"
              className={`settings-usermgmt-form-input ${errors.username ? 'error' : ''}`}
              placeholder="Enter username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              disabled={creating}
            />
            {errors.username && (
              <span className="settings-usermgmt-form-error">{errors.username}</span>
            )}
          </div>

          <div className="settings-usermgmt-form-group">
            <label className="settings-usermgmt-form-label">
              <FaEnvelope className="settings-usermgmt-form-icon" />
              Email Address
            </label>
            <input
              type="email"
              className={`settings-usermgmt-form-input ${errors.email ? 'error' : ''}`}
              placeholder="Enter email address"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              disabled={creating}
            />
            {errors.email && (
              <span className="settings-usermgmt-form-error">{errors.email}</span>
            )}
          </div>

          <div className="settings-usermgmt-form-group">
            <label className="settings-usermgmt-form-label">
              <FaLock className="settings-usermgmt-form-icon" />
              Password
            </label>
            <input
              type="password"
              className={`settings-usermgmt-form-input ${errors.password ? 'error' : ''}`}
              placeholder="Enter password (min. 6 characters)"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={creating}
            />
            {errors.password && (
              <span className="settings-usermgmt-form-error">{errors.password}</span>
            )}
          </div>

          <div className="settings-usermgmt-form-group">
            <label className="settings-usermgmt-form-label">
              <FaLock className="settings-usermgmt-form-icon" />
              Confirm Password
            </label>
            <input
              type="password"
              className={`settings-usermgmt-form-input ${errors.confirmPassword ? 'error' : ''}`}
              placeholder="Confirm your password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              disabled={creating}
            />
            {errors.confirmPassword && (
              <span className="settings-usermgmt-form-error">{errors.confirmPassword}</span>
            )}
          </div>

          <div className="settings-usermgmt-form-group">
            <label className="settings-usermgmt-form-label">
              <FaUserTag className="settings-usermgmt-form-icon" />
              User Role
            </label>
            <select 
              value={role} 
              onChange={(e) => setRole(e.target.value as 'admin' | 'viewer')}
              className={`settings-usermgmt-form-select ${errors.role ? 'error' : ''}`}
              disabled={creating}
            >
              <option value="viewer">Viewer</option>
              <option value="admin">Admin</option>
            </select>
            {errors.role && (
              <span className="settings-usermgmt-form-error">{errors.role}</span>
            )}
          </div>

          <div className="settings-usermgmt-modal-actions">
            <button 
              type="button" 
              className="settings-usermgmt-btn-cancel"
              onClick={handleClose}
              disabled={creating}
            >
              Cancel
            </button>
            <button 
              type="submit" 
              className="settings-usermgmt-btn-submit"
              disabled={creating}
            >
              {creating ? 'Adding User...' : 'Add User'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default UserModal;
