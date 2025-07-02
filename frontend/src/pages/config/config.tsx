// pages/config/config.tsx
import React from "react";
import { useState } from "react";
import Sidebar from "../../components/common/sidebar/sidebar";
import Header from "../../components/common/header/header";

// Server Components
import ServerConfig1 from "../../components/server/config/config1/config1";
import ServerConfig2 from "../../components/server/config/config2/config2";

import "./config.css";

function Config() {
  const [activeTab, setActiveTab] = useState<"general" | "advanced">("general");

  const handleTabClick = (tab: "general" | "advanced") => {
    setActiveTab(tab);
  };

  return (
    <>
      <Header />
      <div className="container">
        <Sidebar />
        <div className="content">
          <div className="config-main-content">
            <div className="config-main-container">
              <div className="config-main-top-bar">
                <div className="config-main-feature-tabs">
                  <button
                    className={`config-main-feature-tab ${activeTab === "general" ? "config-main-feature-tab-active" : ""}`}
                    onClick={() => handleTabClick("general")}
                  >
                    <span className="config-main-tab-icon">⚙️</span> General
                    Features
                  </button>
                  <button
                    className={`config-main-feature-tab ${activeTab === "advanced" ? "config-main-feature-tab-active" : ""}`}
                    onClick={() => handleTabClick("advanced")}
                  >
                    <span className="config-main-tab-icon">🛡️</span> Advanced
                    Features
                  </button>
                </div>
              </div>
            </div>

            <div className="config-main-tab-content">
              {activeTab === "general" && <ServerConfig1 />}
              {activeTab === "advanced" && <ServerConfig2 />}
            </div>
          </div>
        </div>
      </div>
    </>
  );
}

export default Config;
