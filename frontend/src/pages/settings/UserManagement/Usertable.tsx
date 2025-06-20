// pages/settings/UserManagement/Usertable.tsx
import React from 'react';
import { FaTrash, FaUser, FaUserShield, FaSpinner } from 'react-icons/fa';

interface User {
  id: string;
  name: string; // This is the username from backend
  email: string;
  role: 'admin' | 'viewer';
}

interface Props {
  users: User[];
  onDelete: (username: string) => void;
  deleting: string | null;
  loading: boolean;
  canDeleteUser: (username: string, role: string) => boolean;
}

const UserTable: React.FC<Props> = ({ 
  users, 
  onDelete, 
  deleting, 
  loading,
  canDeleteUser 
}) => {
  const getRoleBadgeClass = (role: string): string => {
    return role === 'admin' ? 'settings-usermgmt-role-admin' : 'settings-usermgmt-role-viewer';
  };

  const getRoleIcon = (role: string) => {
    return role === 'admin' ? <FaUserShield /> : <FaUser />;
  };

  if (loading && users.length === 0) {
    return (
      <div className="settings-usermgmt-table-container">
        <div className="settings-usermgmt-loading">
          <FaSpinner className="spinning" />
          <p>Loading users...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="settings-usermgmt-table-container">
      {users.length === 0 ? (
        <div className="settings-usermgmt-empty">
          <div className="settings-usermgmt-empty-icon">👥</div>
          <p>No users found. Add your first user to get started.</p>
        </div>
      ) : (
        <table className="settings-usermgmt-table">
          <thead>
            <tr>
              <th>S.No</th>
              <th>Username</th>
              <th>Email Address</th>
              <th>Role</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {users.map((user, index) => {
              const isDeletable = canDeleteUser(user.name, user.role);
              const isDeleting = deleting === user.name;
              
              return (
                <tr key={user.id}>
                  <td>{index + 1}</td>
                  <td>
                    <div className="settings-usertable-username-cell">
                      {getRoleIcon(user.role)}
                      <span>{user.name}</span>
                    </div>
                  </td>
                  <td>{user.email}</td>
                  <td>
                    <span className={`settings-usermgmt-role-badge ${getRoleBadgeClass(user.role)}`}>
                      {user.role}
                    </span>
                  </td>
                  <td>
                    <div className="settings-usermgmt-actions">
                      <button 
                        className={`settings-usermgmt-btn-delete ${!isDeletable ? 'disabled' : ''}`}
                        onClick={() => isDeletable && onDelete(user.name)}
                        disabled={!isDeletable || isDeleting || loading}
                        title={
                          !isDeletable 
                            ? 'Cannot delete the last admin user' 
                            : `Delete ${user.name}`
                        }
                      >
                        {isDeleting ? <FaSpinner className="spinning" /> : <FaTrash />}
                      </button>
                      {!isDeletable && (
                        <span className="settings-usertable-protected-indicator" title="Protected admin user">
                          🔒
                        </span>
                      )}
                    </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      )}
    </div>
  );
};

export default UserTable;
