// pages/health/health.tsx
import React from 'react';
import Sidebar from "../../components/common/sidebar/sidebar";
import HealthDashboard from "../../components/server/health/HealthDashboard";
import Header from "../../components/common/header/header";
import './health.css';

function HealthPage() {
  return (
    <>
      <Header />
      <div className="container">
        <Sidebar />
        <div className="content">
          <div className="health-main-content">
            <HealthDashboard />
          </div>
        </div>
      </div>
    </>
  );
}

export default HealthPage;
