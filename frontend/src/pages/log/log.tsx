// pages/logs/logs.tsx
import React from 'react';
import Sidebar from "../../components/common/sidebar/sidebar";
import Header from "../../components/common/header/header";
import LogFilter from '../../components/server/log/LogFilter';
import LogTable from '../../components/server/log/LogTable';
import LogVolumeCard from '../../components/server/log/LogVolumeCard';
import ErrorSummaryCard from '../../components/server/log/ErrorSummaryCard';
import SuccessRequestCard from '../../components/server/log/SuccessRequestCard';
import LogExportPanel from '../../components/server/log/LogExportPanel';
import { LogProvider } from '../../context/LogContext';
import './log.css';

const Logs: React.FC = () => {
  return (
    <>
      <Header />
      <div className="container">
        <Sidebar />
        <div className="content">
          <LogProvider>
            <div className="server-logs-main-container">
              <div className="server-logs-filter-section">
                <LogFilter />
              </div>

              <div className="server-logs-content-section">
                <div className="server-logs-left-panel">
                  <LogTable />
                  <LogVolumeCard />
                </div>

                <div className="server-logs-right-panel">
                  <ErrorSummaryCard />
                  <SuccessRequestCard />
                  <LogExportPanel />
                </div>
              </div>
            </div>
          </LogProvider>
        </div>
      </div>
    </>
  );
};

export default Logs;
