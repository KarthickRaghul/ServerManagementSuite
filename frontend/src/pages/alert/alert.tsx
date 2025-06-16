// pages/alert/alert.tsx
import Sidebar from "../../components/common/sidebar/sidebar";
import AlertDashboard from "../../components/server/alert/AlertDashboard";
import Header from "../../components/common/header/header";
import './alert.css';

function AlertPage() {
  return (
    <>
      <Header />
      <div className="container">
        <Sidebar />
        <div className="content">
          <div className="alert-main-content">
            <AlertDashboard />
          </div>
        </div>
      </div>
    </>
  );
}

export default AlertPage;
