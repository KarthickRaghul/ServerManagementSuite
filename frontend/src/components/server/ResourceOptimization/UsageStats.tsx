// components/server/ResourceOptimization/UsageStats.tsx
import React from 'react';
import './usageStats.css';

interface UsageStatsProps {
  label: string;
  value: number;
  type: 'memory' | 'cpu' | 'disk';
  loading?: boolean;
}

const UsageStats: React.FC<UsageStatsProps> = ({ label, value, type, loading = false }) => {
  const getCircleColor = (type: string, value: number) => {
    if (value > 80) return '#ef4444'; // Red for high usage
    if (value > 60) return '#f59e0b'; // Orange for medium usage
    return '#22c55e'; // Green for low usage
  };

  const getGradientColor = (type: string) => {
    switch (type) {
      case 'memory':
        return 'linear-gradient(135deg, #8b5cf6, #a855f7)';
      case 'cpu':
        return 'linear-gradient(135deg, #3b82f6, #60a5fa)';
      case 'disk':
        return 'linear-gradient(135deg, #f59e0b, #fbbf24)';
      default:
        return 'linear-gradient(135deg, #22c55e, #4ade80)';
    }
  };

  const circleColor = getCircleColor(type, value);
  const gradientColor = getGradientColor(type);

  if (loading) {
    return (
      <div className="resource-usage-stat">
        <div 
          className="resource-usage-circle loading"
          style={{ background: gradientColor }}
        >
          <div className="resource-usage-inner">
            <span className="resource-usage-loading">...</span>
          </div>
        </div>
        <p className="resource-usage-label">{label}</p>
      </div>
    );
  }

  return (
    <div className="resource-usage-stat">
      <div 
        className="resource-usage-circle"
        style={{
          background: gradientColor,
          boxShadow: `0 8px 20px ${circleColor}40`
        }}
      >
        <div className="resource-usage-inner">
          <span className="resource-usage-value" style={{ color: circleColor }}>
            {value.toFixed(1)}%
          </span>
        </div>
      </div>
      <p className="resource-usage-label">{label}</p>
    </div>
  );
};

export default UsageStats;
