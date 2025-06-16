// components/common/notification/NotificationContainer.tsx
import React from 'react';
import { useNotification } from '../../../context/NotificationContext';
import NotificationComponent from './Notification';
import './NotificationContainer.css';

const NotificationContainer: React.FC = () => {
  const { notifications, removeNotification } = useNotification();

  return (
    <div className="notification-container">
      {notifications.map(notification => (
        <NotificationComponent
          key={notification.id}
          notification={notification}
          onRemove={removeNotification}
        />
      ))}
    </div>
  );
};

export default NotificationContainer;
