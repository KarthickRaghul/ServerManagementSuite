// components/server/config/config1/components/servermanagement.tsx
import React, { useState } from "react";
import "./servermanagement.css";
import ModalWrapper from "./modalwrapper";
import AccessTokenModal from "../../../../common/accesstoken/AccessTokenModal";
import { FaServer, FaTrash, FaPlus } from "react-icons/fa";
import { useServerManagement } from "../../../../../hooks/server/useServerManagement";
import { useNotification } from "../../../../../context/NotificationContext";

interface ServerFormData {
  ip: string;
  tag: string;
  os: string;
}

interface AccessTokenData {
  token: string;
  deviceInfo: {
    tag: string;
    ip: string;
    os: string;
  };
}

const ServerManagement: React.FC = () => {
  const [showModal, setShowModal] = useState(false);
  const [showTokenModal, setShowTokenModal] = useState(false);
  const [accessTokenData, setAccessTokenData] =
    useState<AccessTokenData | null>(null);
  const [newServer, setNewServer] = useState<ServerFormData>({
    ip: "",
    tag: "",
    os: "linux",
  });
  const [submitError, setSubmitError] = useState<string | null>(null);

  const {
    servers,
    loading,
    error,
    creating,
    deleting,
    createServer,
    deleteServer,
  } = useServerManagement();

  const { addNotification } = useNotification();

  const addServer = async () => {
    if (!newServer.ip.trim() || !newServer.tag.trim()) {
      setSubmitError("IP address and tag are required");
      return;
    }

    // Basic IP validation
    const ipRegex =
      /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
    if (!ipRegex.test(newServer.ip.trim())) {
      setSubmitError("Please enter a valid IP address");
      return;
    }

    setSubmitError(null);

    try {
      const result = await createServer({
        ip: newServer.ip.trim(),
        tag: newServer.tag.trim(),
        os: newServer.os,
      });

      if (result && typeof result === "object" && "access_token" in result) {
        // Show access token modal
        setAccessTokenData({
          token: result.access_token,
          deviceInfo: {
            tag: newServer.tag.trim(),
            ip: newServer.ip.trim(),
            os: newServer.os,
          },
        });
        setShowTokenModal(true);

        addNotification({
          title: "Server Created",
          message: `Server "${newServer.tag}" has been successfully created`,
          type: "success",
          duration: 4000,
        });

        // Reset form
        setNewServer({ ip: "", tag: "", os: "linux" });
        setSubmitError(null);
        setShowModal(false); // ✅ Close modal on success
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to create server";
      setSubmitError(errorMessage);
      addNotification({
        title: "Creation Failed",
        message: errorMessage,
        type: "error",
        duration: 5000,
      });
    }
  };
  const handleDeleteServer = async (ip: string, tag: string) => {
    if (
      window.confirm(`Are you sure you want to delete server "${tag}" (${ip})?`)
    ) {
      try {
        await deleteServer(ip);

        addNotification({
          title: "Server Deleted",
          message: `Server "${tag}" has been successfully deleted`,
          type: "success",
          duration: 4000,
        });
      } catch (err) {
        const errorMessage =
          err instanceof Error ? err.message : "Failed to delete server";
        addNotification({
          title: "Deletion Failed",
          message: errorMessage,
          type: "error",
          duration: 5000,
        });
      }
    }
  };

  const handleClose = () => {
    setShowModal(false);
    setNewServer({ ip: "", tag: "", os: "linux" });
    setSubmitError(null);
  };

  const handleCloseTokenModal = () => {
    setShowTokenModal(false);
    setAccessTokenData(null);
  };

  const formatDate = (dateString: string): string => {
    try {
      return new Date(dateString).toLocaleDateString("en-US", {
        year: "numeric",
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      });
    } catch {
      return dateString;
    }
  };

  const getOSBadgeClass = (os: string): string => {
    switch (os.toLowerCase()) {
      case "windows":
        return "config1-servermgmt-os-windows";
      case "linux":
        return "config1-servermgmt-os-linux";
      default:
        return "config1-servermgmt-os-default";
    }
  };

  return (
    <>
      <div
        className="config1-servermgmt-card-container"
        onClick={() => setShowModal(true)}
      >
        <div className="config1-servermgmt-icon-wrapper">
          <FaServer size={20} color="white" />
        </div>
        <h3>Server Management</h3>
        <p>Manage server infrastructure and monitoring across your network</p>
        {loading && (
          <div className="config1-servermgmt-loading">Loading servers...</div>
        )}
        {error && (
          <div className="config1-servermgmt-error">Error loading servers</div>
        )}
      </div>

      {showModal && (
        <ModalWrapper title="Server Management" onClose={handleClose}>
          <div className="config1-servermgmt-modal-content">
            <p className="config1-servermgmt-subtitle">
              Manage server infrastructure, monitoring, and configuration
            </p>

            {/* Show loading error */}
            {error && (
              <div className="config1-servermgmt-error-banner">
                <p>Error loading servers: {error}</p>
              </div>
            )}

            {/* Show submit error */}
            {submitError && (
              <div className="config1-servermgmt-error-banner">
                <p>{submitError}</p>
              </div>
            )}

            <div className="config1-servermgmt-input-section">
              <div className="config1-servermgmt-input-row">
                <div className="config1-servermgmt-input-group">
                  <label className="config1-servermgmt-input-label">
                    IP Address
                  </label>
                  <input
                    type="text"
                    className="config1-servermgmt-input"
                    placeholder="e.g., 192.168.1.10"
                    value={newServer.ip}
                    onChange={(e) =>
                      setNewServer({ ...newServer, ip: e.target.value })
                    }
                    disabled={creating}
                  />
                </div>
                <div className="config1-servermgmt-input-group">
                  <label className="config1-servermgmt-input-label">
                    Tag/Name
                  </label>
                  <input
                    type="text"
                    className="config1-servermgmt-input"
                    placeholder="e.g., Web Server 01"
                    value={newServer.tag}
                    onChange={(e) =>
                      setNewServer({ ...newServer, tag: e.target.value })
                    }
                    disabled={creating}
                  />
                </div>
                <div className="config1-servermgmt-input-group">
                  <label className="config1-servermgmt-input-label">
                    Operating System
                  </label>
                  <select
                    className="config1-servermgmt-select"
                    value={newServer.os}
                    onChange={(e) =>
                      setNewServer({ ...newServer, os: e.target.value })
                    }
                    disabled={creating}
                  >
                    <option value="linux">Linux</option>
                    <option value="windows">Windows</option>
                  </select>
                </div>
                <button
                  className="config1-servermgmt-btn-add"
                  onClick={addServer}
                  disabled={
                    !newServer.ip.trim() || !newServer.tag.trim() || creating
                  }
                >
                  <FaPlus className="config1-servermgmt-btn-icon" />
                  {creating ? "Adding..." : "Add Server"}
                </button>
              </div>
            </div>

            <div className="config1-servermgmt-table-container">
              {loading ? (
                <div className="config1-servermgmt-loading-state">
                  <div className="config1-servermgmt-loading-spinner">
                    Loading servers...
                  </div>
                </div>
              ) : (
                <table className="config1-servermgmt-table">
                  <thead>
                    <tr>
                      <th>IP Address</th>
                      <th>Tag/Name</th>
                      <th>Operating System</th>
                      <th>Created At</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {servers.length === 0 ? (
                      <tr>
                        <td colSpan={5}>
                          <div className="config1-servermgmt-empty-state">
                            <div className="config1-servermgmt-empty-icon">
                              🖥️
                            </div>
                            <p>
                              No servers found. Add your first server above.
                            </p>
                          </div>
                        </td>
                      </tr>
                    ) : (
                      servers.map((server) => (
                        <tr key={server.id}>
                          <td>{server.ip}</td>
                          <td>{server.tag}</td>
                          <td>
                            <span
                              className={`config1-servermgmt-os-badge ${getOSBadgeClass(server.os)}`}
                            >
                              {server.os.charAt(0).toUpperCase() +
                                server.os.slice(1)}
                            </span>
                          </td>
                          <td>{formatDate(server.created_at)}</td>
                          <td>
                            <div className="config1-servermgmt-action-buttons">
                              <button
                                className="config1-servermgmt-btn-delete"
                                onClick={() =>
                                  handleDeleteServer(server.ip, server.tag)
                                }
                                disabled={deleting === server.ip}
                                title="Delete Server"
                              >
                                {deleting === server.ip ? "..." : <FaTrash />}
                              </button>
                            </div>
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              )}
            </div>
          </div>
        </ModalWrapper>
      )}

      {/* Access Token Modal */}
      {accessTokenData && (
        <AccessTokenModal
          isOpen={showTokenModal}
          onClose={handleCloseTokenModal}
          accessToken={accessTokenData.token}
          deviceInfo={accessTokenData.deviceInfo}
        />
      )}
    </>
  );
};

export default ServerManagement;
