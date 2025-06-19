// components/common/header/header.tsx
import React, { useState, useRef, useEffect } from "react";
import "./header.css";
import { FaBell, FaServer, FaSignOutAlt, FaUser } from "react-icons/fa";
import { IoIosArrowDown } from "react-icons/io";
import { useNavigate, useLocation } from "react-router-dom";
import { useAuth } from "../../../hooks";
import { useRole } from "../../../hooks";
import type { Device } from "../../../types/app";
import { useAppContext } from "../../../context/AppContext";
import { useConnectionOverlay } from "../../../context/ConnectionOverlayContext";

const Header: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { logout } = useAuth();
  const { isAdmin, userInfo } = useRole();
  const { 
    activeDevice, 
    updateActiveDevice,
    devices,
    devicesLoading,
    devicesError,
    refreshDevices
  } = useAppContext();
  
  const { checkConnection } = useConnectionOverlay();
  
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const [userDropdownOpen, setUserDropdownOpen] = useState(false);

  const dropdownRef = useRef<HTMLDivElement>(null);
  const userDropdownRef = useRef<HTMLDivElement>(null);

  const getPageTitle = (pathname: string): string => {
    const routeTitles: { [key: string]: string } = {
      "/": "Configuration",
      "/login": "Login",
      "/health": "System Health",
      "/logs": "Logging Systems",
      "/alert": "Monitoring & Alerts",
      "/resource": "Resource Optimization",
      "/settings": "Settings",
    };

    return routeTitles[pathname] || "SNSMS";
  };

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setDropdownOpen(false);
      }
      if (userDropdownRef.current && !userDropdownRef.current.contains(event.target as Node)) {
        setUserDropdownOpen(false);
      }
    };

    if (dropdownOpen || userDropdownOpen) {
      document.addEventListener("mousedown", handleClickOutside);
    }

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [dropdownOpen, userDropdownOpen]);

  const handleDeviceSelect = async (device: Device) => {
    updateActiveDevice(device);
    setDropdownOpen(false);
    
    // Trigger connection check when device is manually selected
    if (device.ip) {
      await checkConnection(device.ip, false);
    }
  };

  const handleLogout = async () => {
    try {
      await logout();
      setUserDropdownOpen(false);
    } catch (error) {
      console.error('Logout failed:', error);
    }
  };

  const handleRefreshDevices = async () => {
    await refreshDevices();
  };

  const getDeviceDisplayText = () => {
    if (devicesLoading) return 'Loading...';
    if (devicesError) return 'Error loading devices';
    if (devices.length === 0) return 'No devices found';
    if (activeDevice) return activeDevice.tag;
    return 'Select device';
  };

  const getUserDisplayName = () => {
    return userInfo.username || (isAdmin ? 'Admin' : 'Viewer');
  };

  return (
    <div className="header-component-container">
      <div className="header-component-left">
        <div className="header-component-logo-circle">
          <span className="header-component-logo-letter">S</span>
          <span className="header-component-status-indicator" />
        </div>
        <div className="header-component-brand-info">
          <div className="header-component-brand-title">SMS</div>
          <div className="header-component-brand-subtitle">Server Management Suite</div>
        </div>

        <div className="header-component-page-title-section">
          <div className="header-component-page-title">{getPageTitle(location.pathname)}</div>
        </div>
      </div>

      <div className="header-component-right">
        {/* Device Dropdown - Only for Admin */}
        {isAdmin && (
          <div className="header-component-server-dropdown-wrapper" ref={dropdownRef}>
            <div className="header-component-server-dropdown" onClick={() => setDropdownOpen(!dropdownOpen)}>
              {getDeviceDisplayText()}
              <IoIosArrowDown
                className={`header-component-dropdown-icon ${dropdownOpen ? "header-component-rotated" : ""}`}
              />
            </div>
            {dropdownOpen && (
              <div className="header-component-server-dropdown-menu">
                {devicesLoading && (
                  <div className="header-component-server-dropdown-item">Loading devices...</div>
                )}
                {devicesError && (
                  <div className="header-component-server-dropdown-item">
                    <span>Error: {devicesError}</span>
                    <button onClick={handleRefreshDevices} style={{ marginLeft: '8px', fontSize: '12px' }}>
                      Retry
                    </button>
                  </div>
                )}
                {!devicesLoading && !devicesError && devices.length === 0 && (
                  <div className="header-component-server-dropdown-item">
                    No server devices registered
                  </div>
                )}
                {!devicesLoading && !devicesError && devices.map((device) => (
                  <div
                    key={device.id}
                    className={`header-component-server-dropdown-item ${
                      activeDevice?.id === device.id ? 'header-component-active-device' : ''
                    }`}
                    onClick={() => handleDeviceSelect(device)}
                  >
                    {device.tag} ({device.ip})
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* Server Status Indicator */}
        <div className="header-component-server-status">
          <FaServer className="header-component-server-icon" />
          <span className="header-component-server-text">Server Mode</span>
        </div>

        {/* Alerts */}
        <div className="header-component-alert-icon" onClick={() => navigate("/alert")}>
          <FaBell className="header-component-bell-icon" />
          <span className="header-component-alert-text">Alerts</span>
        </div>

        {/* User Dropdown */}
        <div className="header-component-user-dropdown-wrapper" ref={userDropdownRef}>
          <div className="header-component-user-dropdown" onClick={() => setUserDropdownOpen(!userDropdownOpen)}>
            <FaUser className="header-component-user-icon" />
            <span className="header-component-user-text">{getUserDisplayName()}</span>
            <IoIosArrowDown
              className={`header-component-dropdown-icon ${userDropdownOpen ? "header-component-rotated" : ""}`}
            />
          </div>
          {userDropdownOpen && (
            <div className="header-component-user-dropdown-menu">
              {/* Profile Settings - Only for Admin */}
              {isAdmin && (
                <>
                  <div 
                    className="header-component-user-dropdown-item"
                    onClick={() => {
                      navigate("/settings");
                      setUserDropdownOpen(false);
                    }}
                  >
                    <FaUser className="header-component-dropdown-item-icon" />
                    Profile Settings
                  </div>
                  <div className="header-component-dropdown-divider"></div>
                </>
              )}
              <div 
                className="header-component-user-dropdown-item header-component-logout-item"
                onClick={handleLogout}
              >
                <FaSignOutAlt className="header-component-dropdown-item-icon" />
                Logout
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default Header;
