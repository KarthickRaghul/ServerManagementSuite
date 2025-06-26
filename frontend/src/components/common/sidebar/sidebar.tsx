// components/common/sidebar/sidebar.tsx
import { useNavigate, useLocation } from "react-router-dom";
import "./sidebar.css";
import icons from "../../../assets/icons";
import { useRole } from "../../../hooks";
import { useHealthMetrics } from "../../../hooks/server/useHealthMetrics"; // ✅ Fix 1: Correct import path
import { useAppContext } from "../../../context/AppContext"; // ✅ Fix 2: Add missing import
import { useNotification } from "../../../context/NotificationContext"; // ✅ Fix 3: Add missing import
import {
  formatPercentage,
  formatBytes,
  getChangeType,
  getMetricIcon,
  calculatePreviousValue,
  calculateChange,
} from "../../../utils/metricsUtils";
import { useEffect, useState } from "react";
// ✅ Fix 4: Remove unused React import
import { FaSync, FaExclamationTriangle } from "react-icons/fa";

interface Metric {
  icon: string;
  value: string;
  label: string;
  change: string;
  changeType: "positive" | "negative" | "neutral";
}

interface MenuItem {
  label: string;
  icon: keyof typeof icons;
  path: string;
  count?: number;
  alert?: boolean;
  roles: ("admin" | "viewer")[];
}

const allMenuItems: MenuItem[] = [
  { label: "Configuration", icon: "config", path: "/", roles: ["admin"] },
  {
    label: "Health",
    icon: "health",
    path: "/health",
    roles: ["admin", "viewer"],
  },
  {
    label: "Monitoring & Alerts",
    icon: "reports",
    path: "/alert",
    alert: true,
    roles: ["admin", "viewer"],
  },
  {
    label: "Resource Optimization",
    icon: "resource",
    path: "/resource",
    roles: ["admin"],
  },
  {
    label: "Logging Systems",
    icon: "logg",
    path: "/log",
    roles: ["admin", "viewer"],
  },
];

const Sidebar = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { role, isAdmin } = useRole();
  const { activeDevice } = useAppContext(); // ✅ Fix 5: Get activeDevice
  const { addNotification } = useNotification(); // ✅ Fix 6: Get notification function
  const { metrics, healthData, loading, error, refreshMetrics } =
    useHealthMetrics();

  // ✅ Fix 7: Add refreshing state
  const [refreshing, setRefreshing] = useState(false);

  // Store previous values for change calculation
  const [previousMetrics, setPreviousMetrics] = useState<{
    cpu: number;
    ram: number;
    disk: number;
    network: number;
  } | null>(null);

  // Update previous metrics when new data comes in
  useEffect(() => {
    if (metrics && !previousMetrics) {
      // Initialize previous metrics with slight variations for demo
      setPreviousMetrics({
        cpu: calculatePreviousValue(metrics.cpu),
        ram: calculatePreviousValue(metrics.ram),
        disk: calculatePreviousValue(metrics.disk),
        network: calculatePreviousValue(metrics.network),
      });
    }
  }, [metrics, previousMetrics]);

  // ✅ Fix 8: Enhanced refresh handler with proper error handling
  const handleRefresh = async () => {
    if (!activeDevice) {
      addNotification({
        title: "No Device Selected",
        message: "Please select a device to refresh metrics",
        type: "warning",
        duration: 3000,
      });
      return;
    }

    if (refreshing || loading) {
      return; // Prevent multiple simultaneous refreshes
    }

    setRefreshing(true);
    try {
      await refreshMetrics();
      addNotification({
        title: "Metrics Refreshed",
        message: "System metrics have been updated successfully",
        type: "success",
        duration: 2000,
      });

      // Reset previous metrics to recalculate changes
      setPreviousMetrics(null);
    } catch (err) {
      console.error("Failed to refresh metrics:", err);
      addNotification({
        title: "Refresh Failed",
        message: "Failed to refresh system metrics",
        type: "error",
        duration: 4000,
      });
    } finally {
      setRefreshing(false);
    }
  };

  // Filter menu items based on user role only (no mode filtering needed)
  const menuItems = allMenuItems.filter(
    (item) => role && item.roles.includes(role),
  );

  // Generate dynamic metrics based on health data
  const getDynamicMetrics = (): Metric[] => {
    if (loading && !healthData) {
      return [
        {
          icon: "⏳",
          value: "Loading...",
          label: "CPU Usage",
          change: "0%",
          changeType: "neutral",
        },
        {
          icon: "⏳",
          value: "Loading...",
          label: "RAM Usage",
          change: "0%",
          changeType: "neutral",
        },
        {
          icon: "⏳",
          value: "Loading...",
          label: "Disk Usage",
          change: "0%",
          changeType: "neutral",
        },
        {
          icon: "⏳",
          value: "Loading...",
          label: "Network I/O",
          change: "0%",
          changeType: "neutral",
        },
      ];
    }

    if (error && !healthData) {
      return [
        {
          icon: "❌",
          value: "Error",
          label: "CPU Usage",
          change: "0%",
          changeType: "negative",
        },
        {
          icon: "❌",
          value: "Error",
          label: "RAM Usage",
          change: "0%",
          changeType: "negative",
        },
        {
          icon: "❌",
          value: "Error",
          label: "Disk Usage",
          change: "0%",
          changeType: "negative",
        },
        {
          icon: "❌",
          value: "Error",
          label: "Network I/O",
          change: "0%",
          changeType: "negative",
        },
      ];
    }

    if (!metrics || !healthData || !activeDevice) {
      return [
        {
          icon: "📊",
          value: "N/A",
          label: "CPU Usage",
          change: "0%",
          changeType: "neutral",
        },
        {
          icon: "📊",
          value: "N/A",
          label: "RAM Usage",
          change: "0%",
          changeType: "neutral",
        },
        {
          icon: "📊",
          value: "N/A",
          label: "Disk Usage",
          change: "0%",
          changeType: "neutral",
        },
        {
          icon: "📊",
          value: "N/A",
          label: "Network I/O",
          change: "0%",
          changeType: "neutral",
        },
      ];
    }

    const cpuChange = previousMetrics
      ? calculateChange(metrics.cpu, previousMetrics.cpu)
      : "+0%";
    const ramChange = previousMetrics
      ? calculateChange(metrics.ram, previousMetrics.ram)
      : "+0%";
    const diskChange = previousMetrics
      ? calculateChange(metrics.disk, previousMetrics.disk)
      : "+0%";
    const networkChange = previousMetrics
      ? calculateChange(metrics.network, previousMetrics.network)
      : "+0%";

    return [
      {
        icon: getMetricIcon("cpu"),
        value: formatPercentage(metrics.cpu),
        label: "CPU Usage",
        change: cpuChange,
        changeType: getChangeType(metrics.cpu, { warning: 70, critical: 90 }),
      },
      {
        icon: getMetricIcon("ram"),
        value: formatPercentage(metrics.ram),
        label: "RAM Usage",
        change: ramChange,
        changeType: getChangeType(metrics.ram, { warning: 80, critical: 95 }),
      },
      {
        icon: getMetricIcon("disk"),
        value: formatPercentage(metrics.disk),
        label: "Disk Usage",
        change: diskChange,
        changeType: getChangeType(metrics.disk, { warning: 85, critical: 95 }),
      },
      {
        icon: getMetricIcon("network"),
        value: formatBytes(metrics.network * 1024 * 1024), // Convert MB to bytes for formatting
        label: "Network I/O",
        change: networkChange,
        changeType: "neutral", // Network I/O is generally neutral
      },
    ];
  };

  const dynamicMetrics = getDynamicMetrics();

  return (
    <div className="sidebar-container">
      {/* ✅ Fix 9: Enhanced metrics header */}
      <div className="metrics-header">
        <button
          className={`metrics-refresh-btn ${refreshing || loading ? "refreshing" : ""}`}
          onClick={handleRefresh}
          disabled={refreshing || loading || !activeDevice}
          title={!activeDevice ? "Select a device first" : "Refresh metrics"}
        >
          <FaSync
            className={`refresh-icon ${refreshing || loading ? "spinning" : ""}`}
          />
          {refreshing ? "Refreshing..." : "Refresh"}
        </button>
      </div>

      <div className="metrics-section">
        {dynamicMetrics.map((metric, index) => (
          <div key={index} className={`metric-card metric-${index}`}>
            <div className="metric-change" data-type={metric.changeType}>
              {metric.change}
            </div>
            <div className="metric-value">{metric.value}</div>
            <div className="metric-label">{metric.label}</div>
          </div>
        ))}

        {/* ✅ Fix 10: Enhanced error indicator */}
        {error && (
          <div className="metrics-error">
            <FaExclamationTriangle className="error-icon" />
            <div className="error-content">
              <span className="error-title">Metrics Error</span>
              <span className="error-message">{error}</span>
              <button
                className="error-retry-btn"
                onClick={handleRefresh}
                disabled={refreshing || loading}
              >
                Retry
              </button>
            </div>
          </div>
        )}

        {/* ✅ Fix 11: No device selected state */}
        {!activeDevice && (
          <div className="no-device-message">
            <span>Select a device to view metrics</span>
          </div>
        )}
      </div>

      <div className="menu-section">
        {menuItems.map((item, index) => {
          const isActive = location.pathname === item.path;

          return (
            <div
              key={index}
              className={`menu-item ${isActive ? "active-blue" : ""}`}
              onClick={() => navigate(item.path)}
            >
              <img
                src={icons[item.icon]}
                alt={item.label}
                className="menu-icon"
              />
              <span className="menu-label">{item.label}</span>
              {item.count !== undefined && (
                <span className={`menu-count ${item.alert ? "alert" : ""}`}>
                  {item.count}
                </span>
              )}
            </div>
          );
        })}

        {/* Settings - Only for Admin */}
        {isAdmin && (
          <div
            className={`menu-item settings-button ${location.pathname === "/settings" ? "active-blue" : ""}`}
            onClick={() => navigate("/settings")}
          >
            <img src={icons.settings} alt="Settings" className="menu-icon" />
            <span className="menu-label">Settings</span>
          </div>
        )}
      </div>
    </div>
  );
};

export default Sidebar;
