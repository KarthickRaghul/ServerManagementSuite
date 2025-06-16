// components/common/notification/Notification.tsx
import React, { useEffect, useState } from 'react';
import { FaCheckCircle, FaExclamationCircle, FaInfoCircle, FaTimes } from 'react-icons/fa';
import './Notification.css';
import type { Notification as NotificationType } from '../../../context/NotificationContext';

interface NotificationProps {
  notification: NotificationType;
  onRemove: (id: string) => void;
}

const NotificationComponent: React.FC<NotificationProps> = ({ notification, onRemove }) => {
  const [isVisible, setIsVisible] = useState(false);
  const [isRemoving, setIsRemoving] = useState(false);

  useEffect(() => {
    // Trigger entrance animation
    setTimeout(() => setIsVisible(true), 10);
  }, []);

  const handleRemove = () => {
    setIsRemoving(true);
    setTimeout(() => {
      onRemove(notification.id);
    }, 300); // Match the exit animation duration
  };

  const getIcon = () => {
    switch (notification.type) {
      case 'success':
        return <FaCheckCircle className="notification-icon success" />;
      case 'error':
        return <FaExclamationCircle className="notification-icon error" />;
      case 'warning':
        return <FaExclamationCircle className="notification-icon warning" />;
      case 'info':
        return <FaInfoCircle className="notification-icon info" />;
      default:
        return <FaInfoCircle className="notification-icon info" />;
    }
  };

  return (
    <div 
      className={`notification ${notification.type} ${isVisible ? 'visible' : ''} ${isRemoving ? 'removing' : ''}`}
    >
      <div className="notification-content">
        {getIcon()}
        <div className="notification-text">
          <div className="notification-title">{notification.title}</div>
          <div className="notification-message">{notification.message}</div>
        </div>
        <button className="notification-close" onClick={handleRemove}>
          <FaTimes />
        </button>
      </div>
    </div>
  );
};

export default NotificationComponent;
