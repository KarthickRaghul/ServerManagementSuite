import React from "react";
import { useNavigate } from "react-router-dom";
import "./notfound.css";

const NotFound: React.FC = () => {
  const navigate = useNavigate();

  const handleGoHome = () => {
    navigate("/");
  };

  const handleGoBack = () => {
    navigate(-1);
  };

  return (
    <div className="notfound-container">
      <div className="notfound-content">
        <div className="notfound-animation">
          <div className="notfound-circle"></div>
          <div className="notfound-circle"></div>
          <div className="notfound-circle"></div>
        </div>

        <div className="notfound-text">
          <h1 className="notfound-title">404</h1>
          <h2 className="notfound-subtitle">Page Not Found</h2>
          <p className="notfound-description">
            The page you're looking for doesn't exist or has been moved.
            <br />
            Please check the URL or navigate back to safety.
          </p>
        </div>

        <div className="notfound-actions">
          <button
            className="notfound-btn notfound-btn-primary"
            onClick={handleGoHome}
          >
            <span className="btn-icon">🏠</span>
            Go Home
          </button>
          <button
            className="notfound-btn notfound-btn-secondary"
            onClick={handleGoBack}
          >
            <span className="btn-icon">←</span>
            Go Back
          </button>
        </div>

        <div className="notfound-footer">
          <p>Lost in the digital void? Our navigation will guide you back.</p>
        </div>
      </div>
    </div>
  );
};

export default NotFound;
