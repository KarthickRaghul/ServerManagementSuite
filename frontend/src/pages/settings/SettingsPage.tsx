// pages/settings/SettingsPage.tsx
import React from 'react';
import Sidebar from '../../components/common/sidebar/sidebar';
import Header from '../../components/common/header/header';
import UserManagement from './UserManagement/UserManagement';
import './settings.css';

const SettingsPage: React.FC = () => {
  return (
    <>
      <Header />
      <div className="container">
        <Sidebar />
        <div className="content">
          <div className="settings-main-content">
            <div className="settings-header-section">
              <h1 className="settings-page-title">Settings</h1>
              <p className="settings-page-subtitle">Manage user accounts and system configuration</p>
            </div>
            <UserManagement />
          </div>
        </div>
      </div>
    </>
  );
};

export default SettingsPage;
