// pages/settings/UserManagement/Usertable.tsx
import React from 'react';
import { FaTrash } from 'react-icons/fa';

interface User {
  id: string;
  name: string;
  email: string;
  role: 'admin' | 'viewer';
}

interface Props {
  users: User[];
  onDelete: (username: string) => void;
  deleting: string | null;
  loading: boolean;
}

const UserTable: React.FC<Props> = ({ users, onDelete, deleting, loading }) => {
  const getRoleBadgeClass = (role: string): string => {
    return role === 'admin' ? 'settings-usermgmt-role-admin' : 'settings-usermgmt-role-viewer';
  };

  if (loading) {
    return (
      <div className="settings-usermgmt-table-container">
        <div className="settings-usermgmt-loading">Loading users...</div>
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
            {users.map((user, index) => (
              <tr key={user.id}>
                <td>{index + 1}</td>
                <td>{user.name}</td>
                <td>{user.email}</td>
                <td>
                  <span className={`settings-usermgmt-role-badge ${getRoleBadgeClass(user.role)}`}>
                    {user.role}
                  </span>
                </td>
                <td>
                  <div className="settings-usermgmt-actions">
                    <button 
                      className="settings-usermgmt-btn-delete"
                      onClick={() => onDelete(user.name)}
                      disabled={deleting === user.name}
                      title="Delete User"
                    >
                      {deleting === user.name ? '...' : <FaTrash />}
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
};

export default UserTable;
