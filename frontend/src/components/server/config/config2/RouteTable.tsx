// components/server/config2/RouteTable.tsx
import React, { useState } from 'react';
import { FaRoute, FaTrash, FaPlus, FaFlag } from 'react-icons/fa';
import { useConfig2 } from '../../../../hooks/server/useConfig2';
import RouteModal from './RouteModal';
import './RouteTable.css';

const RouteTable: React.FC = () => {
  const [showModal, setShowModal] = useState(false);
  const { routeTable, loading, updateRoute, fetchRouteTable } = useConfig2();

  const handleAddRoute = async (routeData: {
    action: string;
    destination: string;
    gateway: string;
    interface?: string;
    metric?: string;
  }) => {
    const success = await updateRoute(routeData);
    if (success) {
      await fetchRouteTable();
      return true;
    }
    return false;
  };

  interface Route {
    destination: string;
    gateway: string;
    iface?: string;
    metric?: string;
    genmask?: string;
    flags?: string;
  }

  const handleDeleteRoute = async (route: Route) => {
    if (window.confirm(`Are you sure you want to delete route to ${route.destination}?`)) {
      const success = await updateRoute({
        action: 'delete',
        destination: route.destination,
        gateway: route.gateway === '*' ? '0.0.0.0' : route.gateway,
        interface: route.iface,
        metric: route.metric
      });
      
      if (success) {
        await fetchRouteTable();
      }
    }
  };

  const formatDestination = (destination: string, genmask: string) => {
    if (destination === '0.0.0.0' && genmask === '0.0.0.0') {
      return 'default';
    }
    
    const cidr = netmaskToCidr(genmask);
    return cidr ? `${destination}/${cidr}` : destination;
  };

  const netmaskToCidr = (netmask: string): number | null => {
    const netmaskMap: { [key: string]: number } = {
      '255.255.255.255': 32,
      '255.255.255.254': 31,
      '255.255.255.252': 30,
      '255.255.255.248': 29,
      '255.255.255.240': 28,
      '255.255.255.224': 27,
      '255.255.255.192': 26,
      '255.255.255.128': 25,
      '255.255.255.0': 24,
      '255.255.254.0': 23,
      '255.255.252.0': 22,
      '255.255.248.0': 21,
      '255.255.240.0': 20,
      '255.255.224.0': 19,
      '255.255.192.0': 18,
      '255.255.128.0': 17,
      '255.255.0.0': 16,
      '255.254.0.0': 15,
      '255.252.0.0': 14,
      '255.248.0.0': 13,
      '255.240.0.0': 12,
      '255.224.0.0': 11,
      '255.192.0.0': 10,
      '255.128.0.0': 9,
      '255.0.0.0': 8,
      '254.0.0.0': 7,
      '252.0.0.0': 6,
      '248.0.0.0': 5,
      '240.0.0.0': 4,
      '224.0.0.0': 3,
      '192.0.0.0': 2,
      '128.0.0.0': 1,
      '0.0.0.0': 0
    };
    return netmaskMap[netmask] ?? null;
  };

  const getFlagDescription = (flags: string) => {
    const flagMap: { [key: string]: string } = {
      'U': 'Up',
      'G': 'Gateway',
      'H': 'Host',
      'R': 'Reinstate',
      'D': 'Dynamic',
      'M': 'Modified',
      'A': 'Addrconf',
      'C': 'Cache',
      '!': 'Reject'
    };
    
    return flags.split('').map(flag => flagMap[flag] || flag).join(', ');
  };

  return (
    <div className="route-table-card">
      <div className="route-table-header">
        <div className="route-table-title-section">
          <div className="route-table-icon-wrapper">
            <FaRoute className="route-table-icon" />
          </div>
          <div>
            <h3 className="route-table-title">Route Table</h3>
            <p className="route-table-description">Manage network routing</p>
          </div>
        </div>
        <button 
          className="route-table-add-btn"
          onClick={() => setShowModal(true)}
          disabled={loading}
        >
          <FaPlus className="route-table-add-icon" />
          Add Route
        </button>
      </div>

      <div className="route-table-content">
        {loading ? (
          <div className="route-table-loading">Loading routes...</div>
        ) : (
          <div className="route-table-container">
            {routeTable.length === 0 ? (
              <div className="route-table-empty">
                <FaRoute className="route-table-empty-icon" />
                <p>No routes configured</p>
              </div>
            ) : (
              <table className="route-table">
                <thead>
                  <tr>
                    <th>Destination</th>
                    <th>Gateway</th>
                    <th>Interface</th>
                    <th>Flags</th>
                    <th>Metric</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {routeTable.map((route, index) => (
                    <tr key={index} className="route-table-row">
                      <td className="route-table-destination">
                        {formatDestination(route.destination, route.genmask)}
                      </td>
                      <td className="route-table-gateway">
                        {route.gateway === '*' ? 'Direct' : route.gateway}
                      </td>
                      <td className="route-table-interface">{route.iface}</td>
                      <td className="route-table-flags">
                        <span 
                          className="route-flags-badge" 
                          title={getFlagDescription(route.flags)}
                        >
                          <FaFlag className="route-flags-icon" />
                          {route.flags}
                        </span>
                      </td>
                      <td className="route-table-metric">{route.metric}</td>
                      <td>
                        <button 
                          className="route-table-delete-btn" 
                          onClick={() => handleDeleteRoute(route)}
                          disabled={loading}
                          title="Delete Route"
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

      <RouteModal
        isOpen={showModal}
        onClose={() => setShowModal(false)}
        onAddRoute={handleAddRoute}
        isLoading={loading}
      />
    </div>
  );
};

export default RouteTable;
