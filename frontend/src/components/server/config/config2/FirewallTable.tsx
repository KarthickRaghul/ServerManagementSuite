// components/server/config2/FirewallTable.tsx
import React, { useState } from "react";
import {
  FaShieldAlt,
  FaTrash,
  FaPlus,
  FaLock,
  FaUnlock,
  FaToggleOn,
  FaToggleOff,
  FaSpinner,
  FaExclamationTriangle,
} from "react-icons/fa";
import { useConfig2 } from "../../../../hooks/server/useConfig2";
import FirewallRuleModal from "./FirewallRuleModal";
import "./FirewallTable.css";

interface WindowsFirewallRule {
  Name: string;
  DisplayName: string;
  Direction: "Inbound" | "Outbound";
  Action: "Allow" | "Block";
  Enabled: "True" | "False";
  Profile: "Public" | "Private" | "Domain" | "Any";
}

interface LinuxFirewallRule {
  chain: string;
  number: number;
  target: string;
  protocol: string;
  source: string;
  destination: string;
  options: string;
  chainName?: string;
  chainPolicy?: string;
  port?: string;
}

interface LinuxFirewallData {
  type: "iptables";
  chains: Array<{
    name: string;
    policy: string;
    rules: LinuxFirewallRule[];
  }>;
  active: boolean;
}

const FirewallTable: React.FC = () => {
  const [showModal, setShowModal] = useState(false);
  const [selectedChain, setSelectedChain] = useState("All");
  const { firewallData, loading, error, updateFirewallRule } = useConfig2();

  // ✅ Enhanced firewall type detection
  const firewallType: "windows" | "linux" =
    Array.isArray(firewallData) &&
    firewallData.length > 0 &&
    "Name" in firewallData[0]
      ? "windows"
      : "linux";

  const handleAddRule = async (ruleData: any) => {
    try {
      const success = await updateFirewallRule(ruleData);
      if (success) {
        setShowModal(false);
        return true;
      }
      return false;
    } catch (err) {
      // Error notification is already handled by the hook
      return false;
    }
  };

  // ✅ Enhanced delete handling for both Windows and Linux
  const handleDeleteRule = async (
    rule: WindowsFirewallRule | LinuxFirewallRule,
  ) => {
    const confirmMessage =
      firewallType === "windows"
        ? `Are you sure you want to delete the rule "${(rule as WindowsFirewallRule).DisplayName || (rule as WindowsFirewallRule).Name}"?`
        : `Are you sure you want to delete this firewall rule?`;

    if (window.confirm(confirmMessage)) {
      let deleteData: any;

      if (firewallType === "windows") {
        const windowsRule = rule as WindowsFirewallRule;
        // ✅ Windows delete request
        deleteData = {
          action: "delete",
          name: windowsRule.Name, // Use the rule name for deletion
        };
      } else {
        const linuxRule = rule as LinuxFirewallRule;
        deleteData = {
          action: "delete",
          rule: linuxRule.target.toLowerCase(),
          protocol: linuxRule.protocol,
          port: extractPortFromOptions(linuxRule.options),
          source:
            linuxRule.source !== "0.0.0.0/0" ? linuxRule.source : undefined,
          destination:
            linuxRule.destination !== "0.0.0.0/0"
              ? linuxRule.destination
              : undefined,
        };
      }

      console.log("🔍 Deleting firewall rule:", deleteData);

      try {
        const success = await updateFirewallRule(deleteData);
        if (success) {
          // Firewall rules will be automatically refreshed by the hook
          return;
        }
      } catch (err) {
        // Error notification is already handled by the hook
        console.error("Failed to delete firewall rule:", err);
      }
    }
  };

  // ✅ Windows rule toggle functionality
  const handleToggleRule = async (rule: WindowsFirewallRule) => {
    if (firewallType === "windows") {
      const newEnabled = rule.Enabled === "True" ? "False" : "True";
      try {
        const success = await updateFirewallRule({
          action: "toggle",
          name: rule.Name,
          enabled: newEnabled,
        });

        if (success) {
          // Firewall rules will be automatically refreshed by the hook
          return;
        }
      } catch (err) {
        // Error notification is already handled by the hook
        console.error("Failed to toggle firewall rule:", err);
      }
    }
  };

  const extractPortFromOptions = (options: string): string => {
    const dptMatch = options.match(/dpt:(\d+)/);
    const sptMatch = options.match(/spt:(\d+)/);
    const portMatch = options.match(/--dport (\d+)/);
    const sportMatch = options.match(/--sport (\d+)/);
    const rangeMatch = options.match(/dpts:(\d+:\d+)/);

    if (dptMatch) return dptMatch[1];
    if (sptMatch) return sptMatch[1];
    if (portMatch) return portMatch[1];
    if (sportMatch) return sportMatch[1];
    if (rangeMatch) return rangeMatch[1];

    return "Any";
  };

  const getActionColor = (action: string): string => {
    switch (action.toUpperCase()) {
      case "ALLOW":
      case "ACCEPT":
        return "firewall-target-accept";
      case "BLOCK":
      case "DROP":
        return "firewall-target-drop";
      case "REJECT":
        return "firewall-target-reject";
      default:
        return "firewall-target-default";
    }
  };

  // ✅ Enhanced Windows table rendering
  const renderWindowsTable = () => {
    const windowsRules = Array.isArray(firewallData)
      ? (firewallData as WindowsFirewallRule[])
      : [];

    return (
      <div className="firewall-table-wrapper">
        <table className="firewall-table">
          <thead>
            <tr>
              <th>Name</th>
              <th>Display Name</th>
              <th>Direction</th>
              <th>Action</th>
              <th>Enabled</th>
              <th>Profile</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {windowsRules.map((rule, index) => (
              <tr key={`${rule.Name}-${index}`} className="firewall-table-row">
                <td className="firewall-rule-name">{rule.Name}</td>
                <td className="firewall-display-name">{rule.DisplayName}</td>
                <td>
                  <span
                    className={`firewall-direction-badge ${rule.Direction.toLowerCase()}`}
                  >
                    {rule.Direction}
                  </span>
                </td>
                <td>
                  <span
                    className={`firewall-target-badge ${getActionColor(rule.Action)}`}
                  >
                    {rule.Action}
                  </span>
                </td>
                <td>
                  <button
                    className={`firewall-toggle-btn ${rule.Enabled === "True" ? "enabled" : "disabled"}`}
                    onClick={() => handleToggleRule(rule)}
                    disabled={loading.updating}
                    title={`${rule.Enabled === "True" ? "Disable" : "Enable"} rule`}
                  >
                    {loading.updating ? (
                      <FaSpinner className="spinning" />
                    ) : rule.Enabled === "True" ? (
                      <FaToggleOn />
                    ) : (
                      <FaToggleOff />
                    )}
                    {rule.Enabled === "True" ? "Enabled" : "Disabled"}
                  </button>
                </td>
                <td className="firewall-profile">{rule.Profile}</td>
                <td className="firewall-table-actions">
                  <button
                    className="firewall-table-delete-btn"
                    onClick={() => handleDeleteRule(rule)}
                    disabled={loading.updating}
                    title="Delete Rule"
                  >
                    {loading.updating ? (
                      <FaSpinner className="spinning" />
                    ) : (
                      <FaTrash />
                    )}
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    );
  };

  const renderLinuxTable = () => {
    const linuxData = firewallData as LinuxFirewallData;

    const getFilteredRules = (): LinuxFirewallRule[] => {
      if (!linuxData?.chains) return [];

      if (selectedChain === "All") {
        const allRules: LinuxFirewallRule[] = [];
        linuxData.chains.forEach((chain) => {
          chain.rules.forEach((rule) => {
            allRules.push({
              ...rule,
              chainName: chain.name,
              chainPolicy: chain.policy,
              port: extractPortFromOptions(rule.options),
            });
          });
        });
        return allRules;
      } else {
        const selectedChainData = linuxData.chains.find(
          (chain) => chain.name === selectedChain,
        );
        return selectedChainData
          ? selectedChainData.rules.map((rule) => ({
              ...rule,
              chainName: selectedChainData.name,
              chainPolicy: selectedChainData.policy,
              port: extractPortFromOptions(rule.options),
            }))
          : [];
      }
    };

    const filteredRules = getFilteredRules();

    return (
      <>
        <div className="firewall-chain-selector">
          <label className="firewall-chain-label">Filter by Chain:</label>
          <select
            className="firewall-chain-select"
            value={selectedChain}
            onChange={(e) => setSelectedChain(e.target.value)}
            disabled={loading.updating}
          >
            <option value="All">
              All Chains (
              {linuxData?.chains?.reduce(
                (total, chain) => total + chain.rules.length,
                0,
              ) || 0}{" "}
              rules)
            </option>
            {linuxData?.chains?.map((chain) => (
              <option key={chain.name} value={chain.name}>
                {chain.name} ({chain.rules.length} rules)
              </option>
            ))}
          </select>
        </div>

        <div className="firewall-table-wrapper">
          <table className="firewall-table">
            <thead>
              <tr>
                {selectedChain === "All" && <th>Chain</th>}
                <th>Rule #</th>
                <th>Target</th>
                <th>Protocol</th>
                <th>Source</th>
                <th>Destination</th>
                <th>Port</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredRules.map((rule, index) => (
                <tr
                  key={`${rule.chainName}-${rule.number}-${index}`}
                  className="firewall-table-row"
                >
                  {selectedChain === "All" && (
                    <td>
                      <span className="firewall-chain-badge">
                        {rule.chainName}
                      </span>
                    </td>
                  )}
                  <td className="firewall-rule-number">{rule.number}</td>
                  <td>
                    <span
                      className={`firewall-target-badge ${getActionColor(rule.target)}`}
                    >
                      {rule.target}
                    </span>
                  </td>
                  <td className="firewall-protocol">{rule.protocol}</td>
                  <td className="firewall-source">{rule.source}</td>
                  <td className="firewall-destination">{rule.destination}</td>
                  <td className="firewall-port">{rule.port}</td>
                  <td className="firewall-table-actions">
                    <button
                      className="firewall-table-delete-btn"
                      onClick={() => handleDeleteRule(rule)}
                      disabled={loading.updating}
                      title="Delete Rule"
                    >
                      {loading.updating ? (
                        <FaSpinner className="spinning" />
                      ) : (
                        <FaTrash />
                      )}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </>
    );
  };

  const getRulesCount = (): number => {
    if (firewallType === "windows") {
      return Array.isArray(firewallData) ? firewallData.length : 0;
    } else {
      const linuxData = firewallData as LinuxFirewallData;
      return (
        linuxData?.chains?.reduce(
          (total, chain) => total + chain.rules.length,
          0,
        ) || 0
      );
    }
  };

  // ✅ Enhanced loading state
  if (loading.firewallData && !firewallData) {
    return (
      <div className="firewall-table-card">
        <div className="firewall-table-loading">
          <div className="firewall-table-loading-spinner">
            <FaSpinner className="spinning" />
          </div>
          <p>Loading firewall rules...</p>
        </div>
      </div>
    );
  }

  // ✅ Enhanced error state
  if (error && !firewallData) {
    return (
      <div className="firewall-table-card">
        <div className="firewall-table-error">
          <div className="firewall-table-error-icon">
            <FaExclamationTriangle />
          </div>
          <div className="firewall-table-error-content">
            <h4>Failed to Load Firewall Rules</h4>
            <p>{error}</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="firewall-table-card">
      <div className="firewall-table-header">
        <div className="firewall-table-title-section">
          <div className="firewall-table-icon-wrapper">
            <FaShieldAlt className="firewall-table-icon" />
          </div>
          <div>
            <h3 className="firewall-table-title">
              {firewallType === "windows"
                ? "Windows Firewall"
                : "iptables Firewall"}{" "}
              Management
            </h3>
            <p className="firewall-table-description">
              Manage{" "}
              {firewallType === "windows"
                ? "Windows Defender Firewall"
                : "iptables"}{" "}
              rules ({getRulesCount()} rules)
            </p>
          </div>
        </div>
        <div className="firewall-table-controls">
          <div className="firewall-status-indicator">
            {(firewallData as LinuxFirewallData)?.active !== false ? (
              <span className="firewall-status-active">
                <FaLock className="firewall-status-icon" />
                Active
              </span>
            ) : (
              <span className="firewall-status-inactive">
                <FaUnlock className="firewall-status-icon" />
                Inactive
              </span>
            )}
          </div>
          <button
            className="firewall-table-add-btn"
            onClick={() => setShowModal(true)}
            disabled={loading.updating}
            title="Add new firewall rule"
          >
            {loading.updating ? (
              <FaSpinner className="spinning" />
            ) : (
              <FaPlus className="firewall-table-add-icon" />
            )}
            Add Rule
          </button>
        </div>
      </div>

      <div className="firewall-table-content">
        <div className="firewall-table-container">
          {getRulesCount() === 0 ? (
            <div className="firewall-table-empty">
              <FaShieldAlt className="firewall-table-empty-icon" />
              <p>No firewall rules configured</p>
              <small>Click "Add Rule" to create your first firewall rule</small>
            </div>
          ) : firewallType === "windows" ? (
            renderWindowsTable()
          ) : (
            renderLinuxTable()
          )}
        </div>

        {/* ✅ Loading indicator when updating */}
        {loading.updating && (
          <div className="firewall-table-updating">
            <FaSpinner className="spinning" />
            <span>Updating firewall rules...</span>
          </div>
        )}
      </div>

      <FirewallRuleModal
        isOpen={showModal}
        onClose={() => setShowModal(false)}
        onAddRule={handleAddRule}
        isLoading={loading.updating}
        firewallType={firewallType}
      />
    </div>
  );
};

export default FirewallTable;
