// components/server/config2/config2.tsx
import React from 'react';
import NetworkBasicInfo from './NetworkBasicInfo';
import NetworkInterfaceManager from './NetworkInterfaceManager';
import './config2.css';

const Config2: React.FC = () => {
  return (
    <div className="network-config2-main-container">
      <div className="network-config2-header">
        <h1 className="network-config2-title">Network Configuration</h1>
        <p className="network-config2-subtitle">Manage network settings and interface configurations</p>
      </div>
      
      <div className="network-config2-cards-grid">
        <NetworkBasicInfo />
        <NetworkInterfaceManager />
      </div>
    </div>
  );
};

export default Config2;
