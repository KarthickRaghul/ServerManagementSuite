import React, { useState, useEffect } from "react";
import "./login.css";
import { FaUser, FaLock, FaShieldAlt } from "react-icons/fa";
import img from "../../assets/img"; // Adjust the path as necessary

const Login: React.FC = () => {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const [rememberMe, setRememberMe] = useState(false);

  // Get MAC address on component mount
  useEffect(() => {
    // Load remembered username if exists
    const rememberedUser = localStorage.getItem("remember_user");
    if (rememberedUser) {
      setUsername(rememberedUser);
      setRememberMe(true);
    }
  }, []);

  const handleLogin = async (): Promise<void> => {
    setError(null);
    setSuccess(false);
    setLoading(true);

    try {
      // Get backend URL from environment variables
      const backendUrl = import.meta.env.VITE_BACKEND_URL;
      const loginEndpoint = `${backendUrl}/api/auth/login`;

      const response = await fetch(loginEndpoint, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include", // Important: This allows cookies to be set by backend
        body: JSON.stringify({ 
          username: username.trim(), 
          password: password.trim() 
        }),
      });

      const data = await response.json();

      if (response.ok && data.status === "ok") {
        // Store access token in localStorage only
        localStorage.setItem("access_token", data.access_token);
        
        // Refresh token is automatically stored in httpOnly cookies by the backend
        // No need to manually handle refresh token here

        // Handle remember me functionality
        if (rememberMe) {
          localStorage.setItem("remember_user", username);
        } else {
          localStorage.removeItem("remember_user");
        }

        setSuccess(true);

        // Redirect after short delay
        setTimeout(() => {
          window.location.href = "/"; // Navigate to dashboard or main page
        }, 1500);

      } else if (response.status === 401) {
        setError("Invalid username or password.");
      } else if (response.status === 403) {
        setError("Device not authorized. Please contact administrator.");
      } else {
        setError(data.message || "Something went wrong. Please try again.");
      }
    } catch (err) {
      console.error('Login error:', err);
      setError("Network error. Check your connection and try again.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-page-wrapper">
      <div className="login-page-background">

        <div className="login-page-container">
          <div className="login-page-card">
            <div className="login-page-header">
              <div className="login-page-icon">
                 <img src={img.citbif} alt="" />
              </div>
              <h1 className="login-page-title">SMS</h1>
              <p className="login-page-subtitle">Server Management Suite</p>
            </div>

            <form
              className="login-page-form"
              onSubmit={(e) => {
                e.preventDefault();
                handleLogin();
              }}
            >
              <div className="login-page-input-section">
                <label htmlFor="username" className="login-page-label">
                  Username
                </label>
                <div className="login-page-input-group">
                  <span className="login-page-input-icon">
                    <FaUser />
                  </span>
                  <input
                    type="text"
                    id="username"
                    className="login-page-input"
                    aria-label="Username"
                    placeholder="Enter your username"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    autoFocus
                    required
                  />
                </div>
              </div>

              <div className="login-page-input-section">
                <label htmlFor="password" className="login-page-label">
                  Password
                </label>
                <div className="login-page-input-group">
                  <span className="login-page-input-icon">
                    <FaLock />
                  </span>
                  <input
                    type="password"
                    id="password"
                    className="login-page-input"
                    aria-label="Password"
                    placeholder="Enter your password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    required
                  />
                </div>
              </div>

              <div className="login-page-options-row">
                <label className="login-page-checkbox-label">
                  <input 
                    type="checkbox" 
                    className="login-page-checkbox"
                    checked={rememberMe}
                    onChange={(e) => setRememberMe(e.target.checked)}
                  />
                  <span className="login-page-checkbox-text">Remember me</span>
                </label>
              </div>

              <button 
                type="submit" 
                className="login-page-submit-btn"
                disabled={loading || !username.trim() || !password.trim()}
              >
                {loading ? (
                  <>
                    <span className="login-page-loading-spinner"></span>
                    Signing In...
                  </>
                ) : (
                  "Sign In"
                )}
              </button>
            </form>

            {/* Status Messages */}
            {error && (
              <div className="login-page-error-message">
                <span className="login-page-error-icon">⚠️</span>
                {error}
              </div>
            )}
            {success && (
              <div className="login-page-success-message">
                <span className="login-page-success-icon">✅</span>
                Login successful! Redirecting...
              </div>
            )}

            <div className="login-page-footer">
              <p className="login-page-copyright">
                © 2025 SNSMS Network Management Suite. All rights reserved.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Login;
