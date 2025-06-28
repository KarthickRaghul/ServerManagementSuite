// components/server/config2/NetworkInterfaceManager.tsx
import React from "react";
import {
  FaCog,
  FaEthernet,
  FaWifi,
  FaSync,
  FaClock,
  FaSpinner,
  FaExclamationTriangle,
} from "react-icons/fa";
import { FiRefreshCw } from "react-icons/fi";
import { useConfig2 } from "../../../../hooks/server/useConfig2";
import "./NetworkInterfaceManager.css";

interface NetworkInterface {
  id: string;
  name: string;
  status: string; // active/inactive
  power: string; // on/off
  type: "wifi" | "ethernet";
}

const NetworkInterfaceManager: React.FC = () => {
  const { networkBasics, loading, error, updateInterface, restartInterface } =
    useConfig2();

  // ✅ Updated to handle power field for enable/disable
  const handleToggle = async (iface: string, action: "enable" | "disable") => {
    try {
      const success = await updateInterface(iface, action);
      if (success) {
        // Success notification is already handled by the hook
        return;
      }
    } catch (err) {
      // Error notification is already handled by the hook
      console.error(`Failed to ${action} interface:`, err);
    }
  };

  const handleRestart = async () => {
    try {
      const success = await restartInterface();
      if (success) {
        // Success notification is already handled by the hook
        return;
      }
    } catch (err) {
      // Error notification is already handled by the hook
      console.error("Failed to restart interface:", err);
    }
  };

  // ✅ Updated to properly map all three fields
  const getInterfaceArray = (): NetworkInterface[] => {
    if (!networkBasics?.interface) return [];
    return Object.entries(networkBasics.interface).map(([key, value]) => ({
      id: key,
      name: value.mode,
      status: value.status, // active/inactive
      power: value.power, // on/off
      type:
        value.mode.includes("wlan") || value.mode.includes("wifi")
          ? "wifi"
          : "ethernet",
    }));
  };

  const interfaces = getInterfaceArray();

  // ✅ Enhanced loading state
  if (loading.networkBasics && !networkBasics) {
    return (
      <div className="network-interface-manager-card">
        <div className="network-interface-manager-loading">
          <div className="network-interface-manager-loading-spinner">
          </div>
          <p>Loading interface information...</p>
        </div>
      </div>
    );
  }

  // ✅ Enhanced error state
  if (error && !networkBasics) {
    return (
      <div className="network-interface-manager-card">
        <div className="network-interface-manager-error">
          <div className="network-interface-manager-error-icon">
            <FaExclamationTriangle />
          </div>
          <div className="network-interface-manager-error-content">
            <h4>Failed to Load Interface Information</h4>
            <p>{error}</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="network-interface-manager-card">
      <div className="network-interface-manager-header">
        <div className="network-interface-manager-title-section">
          <div className="network-interface-manager-icon-wrapper">
            <FaCog className="network-interface-manager-icon" />
          </div>
          <div>
            <h3 className="network-interface-manager-title">
              Interface Manager
            </h3>
            <p className="network-interface-manager-description">
              Control network interfaces
            </p>
          </div>
        </div>
      </div>

      <div className="network-interface-manager-content">
        {/* Interface List */}
        <div className="network-interface-manager-interfaces-section">
          <h4 className="network-interface-manager-section-title">
            Network Interfaces
          </h4>
          <div className="network-interface-manager-interfaces-list">
            {interfaces.length === 0 ? (
              <div className="network-interface-manager-no-interfaces">
                <FaEthernet className="network-interface-manager-no-interfaces-icon" />
                <p>No network interfaces found</p>
              </div>
            ) : (
              interfaces.map((iface) => (
                <div
                  className="network-interface-manager-interface-item"
                  key={iface.id}
                >
                  <div className="network-interface-manager-interface-info">
                    <div className="network-interface-manager-interface-icon-wrapper">
                      {iface.type === "wifi" ? (
                        <FaWifi className="network-interface-manager-interface-icon" />
                      ) : (
                        <FaEthernet className="network-interface-manager-interface-icon" />
                      )}
                    </div>
                    <div className="network-interface-manager-interface-details">
                      <span className="network-interface-manager-interface-name">
                        {iface.name}
                      </span>
                      <div className="network-interface-manager-interface-status-group">
                        {/* ✅ Show both power and status */}
                        <span
                          className={`network-interface-manager-interface-power ${iface.power}`}
                        >
                          {iface.power === "on" ? "Powered On" : "Powered Off"}
                        </span>
                        <span
                          className={`network-interface-manager-interface-status ${iface.status}`}
                        >
                          {iface.status === "active" ? "Active" : "Inactive"}
                        </span>
                      </div>
                    </div>
                  </div>
                  <div className="network-interface-manager-interface-controls">
                    {/* ✅ Enable/Disable based on power field */}
                    <button
                      className={`network-interface-manager-control-btn ${iface.power === "on" ? "enabled" : "enable"}`}
                      onClick={() => handleToggle(iface.name, "enable")}
                      disabled={loading.updating || iface.power === "on"}
                      title={
                        iface.power === "on"
                          ? "Interface is already enabled"
                          : "Enable interface"
                      }
                    >
                      {loading.updating ? (
                        <FaSpinner className="spinning" />
                      ) : (
                        "Enable"
                      )}
                    </button>
                    <button
                      className={`network-interface-manager-control-btn ${iface.power === "off" ? "disabled" : "disable"}`}
                      onClick={() => handleToggle(iface.name, "disable")}
                      disabled={loading.updating || iface.power === "off"}
                      title={
                        iface.power === "off"
                          ? "Interface is already disabled"
                          : "Disable interface"
                      }
                    >
                      {loading.updating ? (
                        <FaSpinner className="spinning" />
                      ) : (
                        "Disable"
                      )}
                    </button>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>

        {/* ✅ Enhanced Status Section with interface summary */}
        <div className="network-interface-manager-status-section">
          <div className="network-interface-manager-status-grid">
            <div className="network-interface-manager-status-item">
              <div className="network-interface-manager-status-icon-wrapper status">
                <FaSync />
              </div>
              <div className="network-interface-manager-status-content">
                <label>Network Status</label>
                <span
                  className={`network-interface-manager-status-value ${networkBasics?.ip_address ? "online" : "offline"}`}
                >
                  {networkBasics?.ip_address ? "Online" : "Offline"}
                </span>
              </div>
            </div>
            <div className="network-interface-manager-status-item">
              <div className="network-interface-manager-status-icon-wrapper uptime">
                <FaClock />
              </div>
              <div className="network-interface-manager-status-content">
                <label>Uptime</label>
                <span className="network-interface-manager-status-value">
                  {networkBasics?.uptime || "N/A"}
                </span>
              </div>
            </div>
            <div className="network-interface-manager-status-item">
              <div className="network-interface-manager-status-icon-wrapper interfaces">
                <FaEthernet />
              </div>
              <div className="network-interface-manager-status-content">
                <label>Active Interfaces</label>
                <span className="network-interface-manager-status-value">
                  {interfaces.filter((i) => i.status === "active").length} of{" "}
                  {interfaces.length}
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* ✅ Enhanced Restart Button */}
        <button
          className={`network-interface-manager-restart-btn ${loading.updating ? "loading" : ""}`}
          onClick={handleRestart}
          disabled={loading.updating}
          title="Restart all network interfaces"
        >
          <FiRefreshCw
            className={`network-interface-manager-restart-icon ${loading.updating ? "spinning" : ""}`}
          />
          {loading.updating
            ? "Restarting Network..."
            : "Restart Network Service"}
        </button>

        {/* ✅ Loading indicator when updating */}
        {loading.updating && (
          <div className="network-interface-manager-updating">
            <FaSpinner className="spinning" />
            <span>Updating interface configuration...</span>
          </div>
        )}
      </div>
    </div>
  );
};

export default NetworkInterfaceManager;
