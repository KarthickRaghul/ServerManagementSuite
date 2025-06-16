// pages/resource/resource.tsx
import React from 'react';
import Sidebar from "../../components/common/sidebar/sidebar";
import Resource from "../../components/server/ResourceOptimization/Resource";
import Header from "../../components/common/header/header";
import './resource.css';

function ResourcePage() {
  return (
    <>
      <Header />
      <div className="container">
        <Sidebar />
        <div className="content">
          <div className="resource-main-content">
            <Resource />
          </div>
        </div>
      </div>
    </>
  );
}

export default ResourcePage;
