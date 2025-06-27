// components/common/networkoperation/NetworkOperationOverlay.tsx
import React from "react";
import { 
  FaSpinner, 
  FaExclamationTriangle, 
  FaCheckCircle,
  FaRedo,
  FaPlus
} from "react-icons/fa";
import { useNetworkOperation } from "../../../context/NetworkOperationContext";
import { useConnectionOverlay } from "../../../context/ConnectionOverlayContext";
import { useAppContext } from "../../../context/AppContext";
import "./NetworkOperationOverlay.css";

const NetworkOperationOverlay: React.FC = () => {
  const { state, hide } = useNetworkOperation();
  const { checkConnection } = useConnectionOverlay();
  const { activeDevice } = useAppContext();

  if (!state.visible) return null;

  const handleRetryConnection = () => {
    if (activeDevice?.ip) {
      hide();
      checkConnection(activeDevice.ip);
    }
  };

  const handleGoToDeviceManagement = () => {
    hide();
    window.location.href = "/"; // Navigate to device management
  };

  const renderInterfaceRestartContent = () => (
    <div className="network-operation-content">
      <div className="network-operation-icon-wrapper restart">
        {state.loading ? (
          <FaSpinner className="spinning" />
        ) : (
          <FaCheckCircle />
        )}
      </div>
      
      <h3 className="network-operation-title">
        {state.loading ? "Restarting Network Interface" : "Interface Restart Complete"}
      </h3>
      
      <p className="network-operation-message">{state.message}</p>
      
      {state.loading && (
        <div className="network-operation-countdown">
          <div className="countdown-circle">
            <span className="countdown-number">{state.countdown}</span>
          </div>
          <p>Connection will be restored in {state.countdown} seconds</p>
        </div>
      )}

      {!state.loading && (
        <div className="network-operation-actions">
          <button 
            className="btn btn-primary"
            onClick={handleRetryConnection}
          >
            <FaRedo /> Test Connection
          </button>
        </div>
      )}
    </div>
  );

  const renderNetworkConfigContent = () => (
    <div className="network-operation-content">
      <div className="network-operation-icon-wrapper warning">
        <FaExclamationTriangle />
      </div>
      
      <h3 className="network-operation-title">Network Configuration Changed</h3>
      
      <p className="network-operation-message">{state.message}</p>
      
      <div className="network-operation-info-box">
        <h4>What to do next:</h4>
        <ul>
          <li><strong>If IP address was NOT changed:</strong> Click "Test Connection" to verify connectivity</li>
          <li><strong>If IP address was changed:</strong> Click "Manage Devices" to re-register with new IP</li>
        </ul>
      </div>

      <div className="network-operation-actions">
        <button 
          className="btn btn-secondary"
          onClick={handleRetryConnection}
        >
          <FaRedo /> Test Connection
        </button>
        <button 
          className="btn btn-primary"
          onClick={handleGoToDeviceManagement}
        >
          <FaPlus /> Manage Devices
        </button>
      </div>
    </div>
  );

  return (
    <div className="network-operation-overlay">
      <div className="network-operation-overlay-backdrop" onClick={!state.loading ? hide : undefined}>
        <div className="network-operation-overlay-container" onClick={(e) => e.stopPropagation()}>
          {state.operationType === "interface_restart" 
            ? renderInterfaceRestartContent() 
            : renderNetworkConfigContent()
          }
          
          {!state.loading && (
            <button 
              className="network-operation-close-btn"
              onClick={hide}
              title="Close"
            >
              ×
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default NetworkOperationOverlay;
