// pages/config/config.tsx
import { useState, useEffect } from "react";
import Sidebar from "../../components/common/sidebar/sidebar";
import Header from "../../components/common/header/header";
import { useAppContext } from "../../context/AppContext"; // Use this instead of useActiveMode

// Server Components
import ServerConfig1 from "../../components/server/config/config1/config1";
import ServerConfig2 from "../../components/server/config/config2/config2";

// Network Components  
import NetworkConfig from "../../components/network/config/config";

import "./config.css";

function Config() {
  const [activeTab, setActiveTab] = useState<"general" | "advanced">("general");
  const { activeMode } = useAppContext(); // This will now trigger re-renders automatically

  // Reset tab when mode changes (optional)
  useEffect(() => {
    setActiveTab("general");
  }, [activeMode]);

  const handleTabClick = (tab: "general" | "advanced") => {
    setActiveTab(tab);
  };

  const renderServerComponents = () => (
    <>
      <div className="config-main-container">
        <div className="config-main-top-bar">
          <div className="config-main-feature-tabs">
            <button
              className={`config-main-feature-tab ${activeTab === "general" ? "config-main-feature-tab-active" : ""}`}
              onClick={() => handleTabClick("general")}
            >
              <span className="config-main-tab-icon">⚙️</span> General Features
            </button>
            <button
              className={`config-main-feature-tab ${activeTab === "advanced" ? "config-main-feature-tab-active" : ""}`}
              onClick={() => handleTabClick("advanced")}
            >
              <span className="config-main-tab-icon">🛡️</span> Advanced Features
            </button>
          </div>
        </div>
      </div>

      <div className="config-main-tab-content">
        {activeTab === "general" && <ServerConfig1 />}
        {activeTab === "advanced" && <ServerConfig2 />}
      </div>
    </>
  );

  const renderNetworkComponents = () => (
    <div className="config-main-tab-content">
      <NetworkConfig />
    </div>
  );

  return (
    <>
      <Header />
      <div className="container">
        <Sidebar />
        <div className="content">
          <div className="config-main-content">
            {activeMode === "server" ? renderServerComponents() : renderNetworkComponents()}
          </div>
        </div>
      </div>
    </>
  );
}

export default Config;
