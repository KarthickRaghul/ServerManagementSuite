// components/server/ResourceOptimization/ResourceCard.tsx
import React from 'react';
import { FaPlay, FaSpinner } from 'react-icons/fa';
import './resourceCard.css';

interface ResourceCardProps {
  icon: React.ReactNode;
  title: string;
  description: string;
  stats: string[];
  type: 'cleanup' | 'restart';
  onAction: () => void;
  loading: boolean;
  disabled: boolean;
}

const ResourceCard: React.FC<ResourceCardProps> = ({ 
  icon, 
  title, 
  description, 
  stats, 
  type, 
  onAction,
  loading,
  disabled
}) => {
  const getCardColor = (type: string) => {
    switch (type) {
      case 'cleanup':
        return 'linear-gradient(135deg, #ef4444, #dc2626)';
      case 'restart':
        return 'linear-gradient(135deg, #3b82f6, #2563eb)';
      default:
        return 'linear-gradient(135deg, #10b981, #059669)';
    }
  };

  return (
    <div className="resource-optimization-card">
      <div className="resource-optimization-header">
        <div 
          className="resource-optimization-icon"
          style={{ background: getCardColor(type) }}
        >
          {icon}
        </div>
        <div className="resource-optimization-info">
          <h4 className="resource-optimization-title">{title}</h4>
          <p className="resource-optimization-description">{description}</p>
        </div>
      </div>
      
      <div className="resource-optimization-stats">
        <div className="resource-optimization-stat">
          <span className="resource-optimization-stat-value">{stats[0]}</span>
          <span className="resource-optimization-stat-label">Size/Count</span>
        </div>
        <div className="resource-optimization-stat">
          <span className="resource-optimization-stat-value">{stats[1]}</span>
          <span className="resource-optimization-stat-label">Status</span>
        </div>
      </div>
      
      <button 
        className="resource-optimization-action-btn"
        onClick={onAction}
        disabled={disabled || loading}
      >
        {loading ? (
          <FaSpinner className="resource-optimization-action-icon spinning" />
        ) : (
          <FaPlay className="resource-optimization-action-icon" />
        )}
        {loading ? 'Cleaning...' : 'Clean Now'}
      </button>
    </div>
  );
};

export default ResourceCard;
