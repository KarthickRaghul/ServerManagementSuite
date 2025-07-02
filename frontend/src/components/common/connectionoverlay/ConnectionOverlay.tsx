// components/common/connectionoverlay/ConnectionOverlay.tsx
import React from "react";
import { useConnectionOverlay } from "../../../context/ConnectionOverlayContext";
import { useRole } from "../../../hooks/auth/useRole";
import {
  FaSpinner,
  FaExclamationTriangle,
  FaServer,
  FaCog,
  FaRedo,
  FaTimes,
  FaSignOutAlt,
} from "react-icons/fa";
import { useNavigate } from "react-router-dom";
import AuthService from "../../../auth/auth";
import "./ConnectionOverlay.css";

const ConnectionOverlay: React.FC = () => {
  const { state, hide, checkConnection } = useConnectionOverlay();
  const { isAdmin } = useRole();
  const navigate = useNavigate();

  if (!state.visible) return null;

  const handleRetry = () => {
    if (state.deviceIp) {
      checkConnection(state.deviceIp, state.isInitialCheck);
    }
  };

  const handleGoToConfig = () => {
    navigate("/");
    hide();
  };

  // ✅ NEW: Handle logout for viewers who can't access device management
  const handleLogout = async () => {
    try {
      await AuthService.logout();
      // This will redirect to login page
    } catch (error) {
      console.error("Logout failed:", error);
      // Force redirect even if logout API fails
      window.location.href = "/login";
    }
  };

  // ✅ NEW: Force close overlay (for edge cases)
  const handleForceClose = () => {
    hide();
  };

  return (
    <div className="conn-overlay-backdrop">
      <div className="conn-overlay-modal">
        {/* ✅ NEW: Close button (top-right) */}
        <button
          className="conn-overlay-close-btn"
          onClick={handleForceClose}
          title="Close overlay"
        >
          <FaTimes />
        </button>

        {state.loading ? (
          <>
            <FaSpinner className="conn-overlay-spinner spinning" />
            <h2>
              {state.isInitialCheck
                ? "Checking server connection..."
                : "Connecting to server..."}
            </h2>
            <p>
              Verifying connection to <strong>{state.deviceIp}</strong>
            </p>
            <div className="conn-overlay-loading-dots">
              <span></span>
              <span></span>
              <span></span>
            </div>
          </>
        ) : state.error ? (
          <>
            <FaExclamationTriangle className="conn-overlay-error-icon" />
            <h2>Connection Failed</h2>
            <p className="conn-overlay-error-message">{state.error}</p>
            <div className="conn-overlay-error-details">
              <h3>Possible Solutions:</h3>
              <ul className="conn-overlay-tips">
                <li>
                  <strong>Network:</strong> Check if the server is online and
                  accessible
                </li>
                <li>
                  <strong>Firewall:</strong> Ensure firewall allows connections
                  (especially on Windows)
                </li>
                <li>
                  <strong>Registration:</strong> Verify the device is registered
                  with the correct key
                </li>
                <li>
                  <strong>IP Address:</strong> Confirm the IP address is correct
                </li>
                <li>
                  <strong>Port:</strong> Check if the client service is running
                  on the correct port
                </li>
              </ul>
            </div>
            <div className="conn-overlay-actions">
              <button
                className="conn-overlay-btn conn-overlay-btn-retry"
                onClick={handleRetry}
              >
                <FaRedo className="conn-overlay-btn-icon" />
                Retry Connection
              </button>

              {/* ✅ Role-based action buttons */}
              {isAdmin ? (
                <button
                  className="conn-overlay-btn conn-overlay-btn-config"
                  onClick={handleGoToConfig}
                >
                  <FaCog className="conn-overlay-btn-icon" />
                  Go to Device Management
                </button>
              ) : (
                // ✅ NEW: Logout option for viewers
                <button
                  className="conn-overlay-btn conn-overlay-btn-logout"
                  onClick={handleLogout}
                >
                  <FaSignOutAlt className="conn-overlay-btn-icon" />
                  Logout
                </button>
              )}
            </div>
            -{/* ✅ NEW: Additional help text for viewers */}
            {!isAdmin && (
              <div className="conn-overlay-viewer-help">
                <p className="conn-overlay-help-text">
                  <strong>Note:</strong> As a viewer, you cannot manage devices.
                  Please contact your administrator to fix server connection
                  issues, or logout to connect to a different server.
                </p>
              </div>
            )}
          </>
        ) : (
          <>
            <FaServer className="conn-overlay-success-icon" />
            <h2>Connected Successfully!</h2>
            <p>
              Connection established with <strong>{state.deviceIp}</strong>
            </p>
            <div className="conn-overlay-success-checkmark">✓</div>
          </>
        )}
      </div>
    </div>
  );
};

export default ConnectionOverlay;
