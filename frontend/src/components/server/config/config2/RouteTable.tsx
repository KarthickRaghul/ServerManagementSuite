// components/server/config2/RouteTable.tsx
import React, { useState } from 'react';
import { FaRoute, FaTrash, FaPlus } from 'react-icons/fa';
import { useConfig2 } from '../../../../hooks/server/useConfig2';
import { useNotification } from '../../../../context/NotificationContext';
import RouteModal from './RouteModal';
import './RouteTable.css';

const RouteTable: React.FC = () => {
  const [showModal, setShowModal] = useState(false);
  const { routeTable, loading, updateRoute, fetchRouteTable } = useConfig2();
  const { addNotification } = useNotification();

  const handleAddRoute = async (routeData: {
    action: string;
    destination: string;
    gateway: string;
    interface?: string;
    metric?: string;
  }) => {
    try {
      const success = await updateRoute(routeData);
      if (success) {
        addNotification({
          title: 'Route Added',
          message: 'Route has been added successfully',
          type: 'success',
          duration: 3000
        });
        await fetchRouteTable();
        return true;
      }
      return false;
    } catch (err) {
      addNotification({
        title: 'Add Route Failed',
        message: err instanceof Error ? err.message : 'Failed to add route',
        type: 'error',
        duration: 5000
      });
      return false;
    }
  };

  const handleDeleteRoute = async (route: { destination: string; gateway: string; interface?: string; metric?: string }) => {
    if (window.confirm(`Are you sure you want to delete route to ${route.destination}?`)) {
      try {
        const success = await updateRoute({
          action: 'delete',
          destination: route.destination,
          gateway: route.gateway,
          interface: route.interface,
          metric: route.metric
        });
        
        if (success) {
          addNotification({
            title: 'Route Deleted',
            message: 'Route has been deleted successfully',
            type: 'success',
            duration: 3000
          });
          await fetchRouteTable();
        }
      } catch (err) {
        addNotification({
          title: 'Delete Route Failed',
          message: err instanceof Error ? err.message : 'Failed to delete route',
          type: 'error',
          duration: 5000
        });
      }
    }
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
                    <th>Metric</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {routeTable.map((route, index) => (
                    <tr key={index} className="route-table-row">
                      <td className="route-table-destination">{route.destination}</td>
                      <td className="route-table-gateway">{route.gateway}</td>
                      <td className="route-table-interface">{route.interface || 'N/A'}</td>
                      <td className="route-table-metric">{route.metric || 'N/A'}</td>
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
