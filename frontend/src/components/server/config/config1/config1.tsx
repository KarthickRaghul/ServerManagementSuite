import React from "react";
import "./config1.css";
import ServerConfiguration from "./components/serverconfiguration";
import CommandExecution from "./components/commandexecution";
import UserManagement from "./components/servermanagement";
import SecurityManagement from "./components/securitymanagement";
import ServerOverview from "./components/serveroverview";
import { FaServer } from "react-icons/fa";

const Config1: React.FC = () => {
  return (
    <div className="config1-main-container">
      {/* Title & Description */}
      <div className="config1-header-section">
        <div className="config1-icon-wrapper">
          <FaServer />
        </div>
        <h1>Server Configuration Management</h1>
        <p>
          Comprehensive server administration, user management, and security
          configuration in one unified platform
        </p>
      </div>

      {/* Four Horizontal Cards */}
      <div className="config1-cards-row-horizontal">
        <ServerConfiguration />
        <CommandExecution />
        <UserManagement />
        <SecurityManagement />
      </div>

      {/* Overview */}
      <div className="config1-overview-wrapper">
        <ServerOverview />
      </div>
    </div>
  );
};

export default Config1;
