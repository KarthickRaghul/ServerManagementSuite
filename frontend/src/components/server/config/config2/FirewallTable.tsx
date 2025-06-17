// components/server/config2/FirewallTable.tsx
import React, { useState } from 'react';
import { FaShieldAlt, FaTrash, FaPlus, FaLock, FaUnlock } from 'react-icons/fa';
import { useConfig2 } from '../../../../hooks/server/useConfig2';
import FirewallRuleModal from './FirewallRuleModal';
import './FirewallTable.css';

const FirewallTable: React.FC = () => {
  const [showModal, setShowModal] = useState(false);
  const [selectedChain, setSelectedChain] = useState('All');
  const { firewallData, loading, updateFirewallRule, fetchFirewallRules } = useConfig2();

  const handleAddRule = async (ruleData: {
    action: string;
    rule: string;
    protocol: string;
    port: string;
    source?: string;
    destination?: string;
  }) => {
    const success = await updateFirewallRule(ruleData);
    if (success) {
      await fetchFirewallRules();
      return true;
    }
    return false;
  };

  const handleDeleteRule = async (rule: { 
    target: string; 
    protocol: string; 
    options: string; 
    source: string; 
    destination: string; 
  }) => {
    if (window.confirm(`Are you sure you want to delete this firewall rule?`)) {
      const success = await updateFirewallRule({
        action: 'delete',
        rule: rule.target.toLowerCase(),
        protocol: rule.protocol,
        port: extractPortFromOptions(rule.options),
        source: rule.source,
        destination: rule.destination
      });
      
      if (success) {
        await fetchFirewallRules();
      }
    }
  };

  const extractPortFromOptions = (options: string): string => {
    // Extract port from various iptables option formats
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
    
    return 'Any';
  };

  const getTargetColor = (target: string) => {
    switch (target.toUpperCase()) {
      case 'ACCEPT':
        return 'firewall-target-accept';
      case 'DROP':
        return 'firewall-target-drop';
      case 'REJECT':
        return 'firewall-target-reject';
      default:
        return 'firewall-target-default';
    }
  };

  const getPolicyColor = (policy: string) => {
    switch (policy.toUpperCase()) {
      case 'ACCEPT':
        return 'firewall-policy-accept';
      case 'DROP':
        return 'firewall-policy-drop';
      case 'REJECT':
        return 'firewall-policy-reject';
      default:
        return 'firewall-policy-default';
    }
  };

  // Get filtered rules based on selected chain
  const getFilteredRules = () => {
    if (!firewallData) return [];
    
    if (selectedChain === 'All') {
      // Combine all rules from all chains
      const allRules: {
        chainName: string;
        chainPolicy: string;
        number: number;
        target: string;
        protocol: string;
        source: string;
        destination: string;
        port: string;
        options: string;
      }[] = [];
      firewallData.chains.forEach(chain => {
        chain.rules.forEach(rule => {
          allRules.push({
            ...rule,
            chainName: chain.name,
            chainPolicy: chain.policy,
            port: extractPortFromOptions(rule.options)
          });
        });
      });
      return allRules;
    } else {
      // Return rules from selected chain
      const selectedChainData = firewallData.chains.find(chain => chain.name === selectedChain);
      return selectedChainData ? selectedChainData.rules.map(rule => ({
        ...rule,
        chainName: selectedChainData.name,
        chainPolicy: selectedChainData.policy,
        port: extractPortFromOptions(rule.options)
      })) : [];
    }
  };

  const getTotalRulesCount = () => {
    if (!firewallData) return 0;
    return firewallData.chains.reduce((total, chain) => total + chain.rules.length, 0);
  };

  const filteredRules = getFilteredRules();
  const selectedChainData = firewallData?.chains.find(chain => chain.name === selectedChain);

  return (
    <div className="firewall-table-card">
      <div className="firewall-table-header">
        <div className="firewall-table-title-section">
          <div className="firewall-table-icon-wrapper">
            <FaShieldAlt className="firewall-table-icon" />
          </div>
          <div>
            <h3 className="firewall-table-title">Firewall Management</h3>
            <p className="firewall-table-description">Manage iptables firewall rules</p>
          </div>
        </div>
        <div className="firewall-table-controls">
          <div className="firewall-status-indicator">
            {firewallData?.active ? (
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
            disabled={loading}
          >
            <FaPlus className="firewall-table-add-icon" />
            Add Rule
          </button>
        </div>
      </div>

      <div className="firewall-table-content">
        {/* Chain Selector */}
        <div className="firewall-chain-selector">
          <label className="firewall-chain-label">Filter by Chain:</label>
          <select 
            className="firewall-chain-select"
            value={selectedChain}
            onChange={(e) => setSelectedChain(e.target.value)}
            disabled={loading}
          >
            <option value="All">All Chains ({getTotalRulesCount()} rules)</option>
            {firewallData?.chains.map(chain => (
              <option key={chain.name} value={chain.name}>
                {chain.name} ({chain.rules.length} rules)
              </option>
            ))}
          </select>
          {selectedChain !== 'All' && selectedChainData && (
            <span className={`firewall-chain-policy ${getPolicyColor(selectedChainData.policy)}`}>
              Policy: {selectedChainData.policy}
            </span>
          )}
          {selectedChain === 'All' && (
            <span className="firewall-all-chains-indicator">
              Showing rules from all chains
            </span>
          )}
        </div>

        {loading ? (
          <div className="firewall-table-loading">Loading firewall rules...</div>
        ) : (
          <div className="firewall-table-container">
            {filteredRules.length === 0 ? (
              <div className="firewall-table-empty">
                <FaShieldAlt className="firewall-table-empty-icon" />
                <p>
                  {selectedChain === 'All' 
                    ? 'No firewall rules configured' 
                    : `No rules configured for ${selectedChain} chain`
                  }
                </p>
              </div>
            ) : (
              <table className="firewall-table">
                <thead>
                  <tr>
                    {selectedChain === 'All' && <th>Chain</th>}
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
                    <tr key={`${rule.chainName}-${rule.number}-${index}`} className="firewall-table-row">
                      {selectedChain === 'All' && (
                        <td>
                          <span className="firewall-chain-badge">
                            {rule.chainName}
                          </span>
                        </td>
                      )}
                      <td className="firewall-rule-number">{rule.number}</td>
                      <td>
                        <span className={`firewall-target-badge ${getTargetColor(rule.target)}`}>
                          {rule.target}
                        </span>
                      </td>
                      <td className="firewall-protocol">{rule.protocol}</td>
                      <td className="firewall-source">{rule.source}</td>
                      <td className="firewall-destination">{rule.destination}</td>
                      <td className="firewall-port">{rule.port}</td>
                      <td>
                        <button 
                          className="firewall-table-delete-btn" 
                          onClick={() => handleDeleteRule(rule)}
                          disabled={loading}
                          title="Delete Rule"
                        >
                          <FaTrash />
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        )}
      </div>

      <FirewallRuleModal
        isOpen={showModal}
        onClose={() => setShowModal(false)}
        onAddRule={handleAddRule}
        isLoading={loading}
      />
    </div>
  );
};

export default FirewallTable;
