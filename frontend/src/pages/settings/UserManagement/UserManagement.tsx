// pages/settings/UserManagement/UserManagement.tsx
import React, { useState } from 'react';
import UserTable from './Usertable';
import UserModal from './UserModal';
import { FaPlus, FaSync } from 'react-icons/fa';
import { useUserManagement } from '../../../hooks/settings/useUserManagement';
import { useNotification } from '../../../context/NotificationContext';
import './UserManagement.css';

const UserManagement: React.FC = () => {
  const [showModal, setShowModal] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);

  const { 
    users, 
    loading, 
    error, 
    creating, 
    deleting, 
    createUser, 
    deleteUser,
    fetchUsers,
    canDeleteUser,
    getAdminCount
  } = useUserManagement();
  
  const { addNotification } = useNotification();

  const adminCount = getAdminCount();

  const handleAddUser = async (newUser: { username: string; email: string; role: 'admin' | 'viewer'; password: string }) => {
    setSubmitError(null);

    try {
      const success = await createUser({
        username: newUser.username,
        email: newUser.email,
        role: newUser.role,
        password: newUser.password
      });

      if (success) {
        addNotification({
          title: 'User Created',
          message: `User "${newUser.username}" has been successfully created`,
          type: 'success',
          duration: 4000
        });
        setShowModal(false);
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to create user';
      setSubmitError(errorMessage);
      addNotification({
        title: 'Creation Failed',
        message: errorMessage,
        type: 'error',
        duration: 5000
      });
    }
  };

  const handleDeleteUser = async (username: string) => {
    // Find the user to get their role
    const userToDelete = users.find(user => user.name === username);
    
    if (!userToDelete) {
      addNotification({
        title: 'Deletion Failed',
        message: 'User not found',
        type: 'error',
        duration: 5000
      });
      return;
    }

    // Check if this user can be deleted
    if (!canDeleteUser(username, userToDelete.role)) {
      addNotification({
        title: 'Cannot Delete User',
        message: 'Cannot delete the last admin user. At least one admin must remain in the system.',
        type: 'warning',
        duration: 6000
      });
      return;
    }

    // Show confirmation dialog with additional warning for admin users
    let confirmMessage = `Are you sure you want to delete user "${username}"?`;
    
    if (userToDelete.role === 'admin') {
      confirmMessage += `\n\nWarning: This is an admin user. After deletion, there will be ${adminCount - 1} admin user(s) remaining.`;
    }

    if (window.confirm(confirmMessage)) {
      try {
        const success = await deleteUser(username);
        
        if (success) {
          addNotification({
            title: 'User Deleted',
            message: `User "${username}" has been successfully deleted`,
            type: 'success',
            duration: 4000
          });
        }
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : 'Failed to delete user';
        addNotification({
          title: 'Deletion Failed',
          message: errorMessage,
          type: 'error',
          duration: 5000
        });
      }
    }
  };

  const handleRefreshUsers = async () => {
    await fetchUsers();
    addNotification({
      title: 'Users Refreshed',
      message: 'User list has been refreshed',
      type: 'info',
      duration: 2000
    });
  };

  const handleCloseModal = () => {
    setShowModal(false);
    setSubmitError(null);
  };

  return (
    <div className="settings-usermgmt-container">
      <div className="settings-usermgmt-header">
        <div className="settings-usermgmt-title-section">
          <h2 className="settings-usermgmt-title">User Management</h2>
          <p className="settings-usermgmt-subtitle">
            Manage system users and their access permissions ({users.length} users, {adminCount} admin{adminCount !== 1 ? 's' : ''})
          </p>
        </div>
        <div className="settings-usermgmt-header-actions">
          <button 
            className="settings-usermgmt-refresh-btn" 
            onClick={handleRefreshUsers}
            disabled={loading}
            title="Refresh user list"
          >
            <FaSync className={`settings-usermgmt-refresh-icon ${loading ? 'spinning' : ''}`} />
          </button>
          <button 
            className="settings-usermgmt-add-btn" 
            onClick={() => setShowModal(true)}
            disabled={creating}
          >
            <FaPlus className="settings-usermgmt-add-icon" />
            {creating ? 'Adding...' : 'Add User'}
          </button>
        </div>
      </div>

      {/* Admin Warning */}
      {adminCount <= 1 && (
        <div className="settings-usermgmt-warning">
          <p>
            <strong>⚠️ Warning:</strong> You have only {adminCount} admin user{adminCount !== 1 ? 's' : ''} in the system. 
            The last admin user cannot be deleted to prevent system lockout.
          </p>
        </div>
      )}

      {/* Show loading error */}
      {error && (
        <div className="settings-usermgmt-error">
          <p>Error: {error}</p>
          <button 
            className="settings-usermgmt-retry-btn"
            onClick={handleRefreshUsers}
            disabled={loading}
          >
            Retry
          </button>
        </div>
      )}
      
      <UserTable 
        users={users} 
        onDelete={handleDeleteUser}
        deleting={deleting}
        loading={loading}
        canDeleteUser={canDeleteUser}
      />
      
      {showModal && (
        <UserModal 
          onClose={handleCloseModal} 
          onSubmit={handleAddUser}
          creating={creating}
          submitError={submitError}
        />
      )}
    </div>
  );
};

export default UserManagement;
