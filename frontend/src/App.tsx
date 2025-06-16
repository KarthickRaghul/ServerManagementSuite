// App.tsx
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import { AppProvider } from "./context/AppContext";
import { NotificationProvider } from "./context/NotificationContext";
import Login from "./pages/login/Login";
import Config from "./pages/config/config";
import Health from "./pages/health/health";
import Log from "./pages/log/log";
import Backup from "./pages/backup/backup";
import Alert from "./pages/alert/alert";
import Resource from "./pages/resource/resource";
import Settings from "./pages/settings/SettingsPage";
import NotFound from "./pages/notfound/notfound";
import ProtectedRoute from "./components/auth/ProtectedRoute";
import RoleProtectedRoute from "./components/auth/RoleProtectedRoute";
import NotificationContainer from "./components/common/notification/NotificationContainer";
import "./App.css";

export default function App() {
  return (
    <NotificationProvider>
      <AppProvider>
        <Router>
          <Routes>
            {/* Public route - Login page */}
            <Route path="/login" element={<Login />} />
            
            {/* Admin only routes */}
            <Route 
              path="/" 
              element={
                <ProtectedRoute>
                  <RoleProtectedRoute allowedRoles={['admin']}>
                    <Config />
                  </RoleProtectedRoute>
                </ProtectedRoute>
              } 
            />
            <Route 
              path="/backup" 
              element={
                <ProtectedRoute>
                  <RoleProtectedRoute allowedRoles={['admin']}>
                    <Backup />
                  </RoleProtectedRoute>
                </ProtectedRoute>
              } 
            />
            <Route 
              path="/resource" 
              element={
                <ProtectedRoute>
                  <RoleProtectedRoute allowedRoles={['admin']}>
                    <Resource />
                  </RoleProtectedRoute>
                </ProtectedRoute>
              } 
            />
            <Route 
              path="/settings" 
              element={
                <ProtectedRoute>
                  <RoleProtectedRoute allowedRoles={['admin']}>
                    <Settings />
                  </RoleProtectedRoute>
                </ProtectedRoute>
              } 
            />
            
            {/* Routes accessible by both admin and viewer */}
            <Route 
              path="/health" 
              element={
                <ProtectedRoute>
                  <RoleProtectedRoute allowedRoles={['admin', 'viewer']}>
                    <Health />
                  </RoleProtectedRoute>
                </ProtectedRoute>
              } 
            />
            <Route 
              path="/log" 
              element={
                <ProtectedRoute>
                  <RoleProtectedRoute allowedRoles={['admin', 'viewer']}>
                    <Log />
                  </RoleProtectedRoute>
                </ProtectedRoute>
              } 
            />
            <Route 
              path="/alert" 
              element={
                <ProtectedRoute>
                  <RoleProtectedRoute allowedRoles={['admin', 'viewer']}>
                    <Alert />
                  </RoleProtectedRoute>
                </ProtectedRoute>
              } 
            />
            
            {/* 404 page */}
            <Route 
              path="*" 
              element={
                <ProtectedRoute>
                  <NotFound />
                </ProtectedRoute>
              } 
            />
          </Routes>
          <NotificationContainer />
        </Router>
      </AppProvider>
    </NotificationProvider>
  );
}
