import React, { useState, useEffect } from "react";
import type { ReactNode } from "react";
import "./modalwrapper.css";
import { FaTimes } from "react-icons/fa";

interface Props {
  title: string;
  children: ReactNode;
  onClose: () => void;
}

const ModalWrapper: React.FC<Props> = ({ title, children, onClose }) => {
  const [isVisible, setIsVisible] = useState(false);
  const [isClosing, setIsClosing] = useState(false);

  useEffect(() => {
    // Trigger opening animation
    setIsVisible(true);
    
    // Prevent body scroll
    document.body.style.overflow = 'hidden';
    
    return () => {
      document.body.style.overflow = 'unset';
    };
  }, []);

  const handleClose = () => {
    setIsClosing(true);
    setTimeout(() => {
      onClose();
    }, 300); // Match CSS transition duration
  };

  const handleOverlayClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      handleClose();
    }
  };

  return (
    <div 
      className={`config1-modal-overlay ${isVisible ? 'config1-modal-visible' : ''} ${isClosing ? 'config1-modal-closing' : ''}`}
      onClick={handleOverlayClick}
    >
      <div className={`config1-modal-container ${isVisible ? 'config1-modal-visible' : ''} ${isClosing ? 'config1-modal-closing' : ''}`}>
        <div className="config1-modal-header-section">
          <h2 className="config1-modal-title">{title}</h2>
          <button className="config1-modal-close-btn" onClick={handleClose}>
            <FaTimes />
          </button>
        </div>
        <div className="config1-modal-content-wrapper">{children}</div>
      </div>
    </div>
  );
};

export default ModalWrapper;
