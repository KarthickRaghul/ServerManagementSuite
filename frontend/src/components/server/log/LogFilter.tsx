// components/server/log/LogFilter.tsx
import React from 'react';
import { FaFilter, FaSearch, FaCalendarAlt, FaSync, FaTrash } from 'react-icons/fa';
import { useLogContext } from '../../../context/LogContext';
import './LogFilter.css';

const LogFilter: React.FC = () => {
  const { 
    filters, 
    updateFilters, 
    fetchLogs, 
    loading, 
    getUniqueApplications, 
    clearFilters 
  } = useLogContext();
  
  const uniqueApplications = getUniqueApplications();

  const handleFilterChange = (field: string, value: string | number) => {
    updateFilters({ [field]: value });
  };

  const handleRefresh = () => {
    fetchLogs();
  };

  const handleClearFilters = () => {
    clearFilters();
  };

  return (
    <div className="server-logs-filter-container">
      <div className="server-logs-filter-header">
        <div className="server-logs-filter-title">
          <div className="server-logs-filter-icon-wrapper">
            <FaFilter className="server-logs-filter-icon" />
          </div>
          <h3>Log Filters</h3>
        </div>
        <div className="server-logs-filter-actions">
          <button 
            className="server-logs-refresh-btn"
            onClick={handleRefresh}
            disabled={loading}
            title="Refresh Logs"
          >
            <FaSync className={`server-logs-refresh-icon ${loading ? 'spinning' : ''}`} />
          </button>
          <button 
            className="server-logs-clear-btn"
            onClick={handleClearFilters}
            disabled={loading}
            title="Clear All Filters"
          >
            <FaTrash />
          </button>
        </div>
      </div>

      <div className="server-logs-filter-fields">
        <div className="server-logs-filter-group">
          <label className="server-logs-filter-label">Search</label>
          <div className="server-logs-search-wrapper">
            <FaSearch className="server-logs-search-icon" />
            <input
              type="text"
              className="server-logs-filter-search"
              placeholder="Search logs..."
              value={filters.search}
              onChange={(e) => handleFilterChange('search', e.target.value)}
              disabled={loading}
            />
          </div>
        </div>

        <div className="server-logs-filter-group">
          <label className="server-logs-filter-label">Log Level</label>
          <select
            className="server-logs-filter-select"
            value={filters.level}
            onChange={(e) => handleFilterChange('level', e.target.value)}
            disabled={loading}
          >
            <option value="">All Levels</option>
            <option value="error">Error</option>
            <option value="warning">Warning</option>
            <option value="info">Info</option>
          </select>
        </div>

        <div className="server-logs-filter-group">
          <label className="server-logs-filter-label">Application</label>
          <select
            className="server-logs-filter-select"
            value={filters.application}
            onChange={(e) => handleFilterChange('application', e.target.value)}
            disabled={loading}
          >
            <option value="">All Applications</option>
            {uniqueApplications.map(app => (
              <option key={app} value={app}>{app}</option>
            ))}
          </select>
        </div>

        <div className="server-logs-filter-group">
          <label className="server-logs-filter-label">Lines</label>
          <select
            className="server-logs-filter-select"
            value={filters.lines}
            onChange={(e) => handleFilterChange('lines', parseInt(e.target.value))}
            disabled={loading}
          >
            <option value={50}>50 lines</option>
            <option value={100}>100 lines</option>
            <option value={200}>200 lines</option>
            <option value={500}>500 lines</option>
            <option value={1000}>1000 lines</option>
          </select>
        </div>

        <div className="server-logs-filter-group">
          <label className="server-logs-filter-label">Date Filter</label>
          <div className="server-logs-date-wrapper">
            <FaCalendarAlt className="server-logs-date-icon" />
            <input
              type="date"
              className="server-logs-filter-date"
              value={filters.date}
              onChange={(e) => handleFilterChange('date', e.target.value)}
              disabled={loading}
              title="Filter logs from this date"
            />
          </div>
        </div>

        <div className="server-logs-filter-group">
          <label className="server-logs-filter-label">Time Filter</label>
          <div className="server-logs-date-wrapper">
            <FaCalendarAlt className="server-logs-date-icon" />
            <input
              type="time"
              className="server-logs-filter-date"
              value={filters.time}
              onChange={(e) => handleFilterChange('time', e.target.value)}
              disabled={loading}
              title="Filter logs from this time"
            />
          </div>
        </div>
      </div>
    </div>
  );
};

export default LogFilter;
