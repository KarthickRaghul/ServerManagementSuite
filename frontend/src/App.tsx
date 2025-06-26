// App.tsx
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import { AppProvider } from "./context/AppContext";
import { NotificationProvider } from "./context/NotificationContext";
import { ConnectionOverlayProvider } from "./context/ConnectionOverlayContext";
import Login from "./pages/login/Login";
import Config from "./pages/config/config";
import Health from "./pages/health/health";
import Logs from "./pages/log/log";
import Alert from "./pages/alert/alert";
import Resource from "./pages/resource/resource";
import Settings from "./pages/settings/SettingsPage";
import NotFound from "./pages/notfound/notfound";
import ProtectedRoute from "./components/auth/ProtectedRoute";
import RoleProtectedRoute from "./components/auth/RoleProtectedRoute";
import ConnectionProtectedRoute from "./components/auth/ConnectionProtectedRoute";
import NotificationContainer from "./components/common/notification/NotificationContainer";
import ConnectionOverlay from "./components/common/connectionoverlay/ConnectionOverlay";
import "./App.css";
import React from "react";

export default function App() {
  return (
    <NotificationProvider>
      <AppProvider>
        <ConnectionOverlayProvider>
          <Router>
            <Routes>
              {/* Public route - Login page */}
              <Route path="/login" element={<Login />} />

              {/* Admin only routes */}
              <Route
                path="/"
                element={
                  <ProtectedRoute>
                    <RoleProtectedRoute allowedRoles={["admin"]}>
                      <Config />
                    </RoleProtectedRoute>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/resource"
                element={
                  <ProtectedRoute>
                    <RoleProtectedRoute allowedRoles={["admin"]}>
                      <ConnectionProtectedRoute>
                        <Resource />
                      </ConnectionProtectedRoute>
                    </RoleProtectedRoute>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/settings"
                element={
                  <ProtectedRoute>
                    <RoleProtectedRoute allowedRoles={["admin"]}>
                      <ConnectionProtectedRoute>
                        <Settings />
                      </ConnectionProtectedRoute>
                    </RoleProtectedRoute>
                  </ProtectedRoute>
                }
              />

              {/* Routes accessible by both admin and viewer */}
              <Route
                path="/health"
                element={
                  <ProtectedRoute>
                    <RoleProtectedRoute allowedRoles={["admin", "viewer"]}>
                      <ConnectionProtectedRoute>
                        <Health />
                      </ConnectionProtectedRoute>
                    </RoleProtectedRoute>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/log"
                element={
                  <ProtectedRoute>
                    <RoleProtectedRoute allowedRoles={["admin", "viewer"]}>
                      <ConnectionProtectedRoute>
                        <Logs />
                      </ConnectionProtectedRoute>
                    </RoleProtectedRoute>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/alert"
                element={
                  <ProtectedRoute>
                    <RoleProtectedRoute allowedRoles={["admin", "viewer"]}>
                      <ConnectionProtectedRoute>
                        <Alert />
                      </ConnectionProtectedRoute>
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
            {/* Connection overlay - renders on top of everything */}
            <ConnectionOverlay />
          </Router>
        </ConnectionOverlayProvider>
      </AppProvider>
    </NotificationProvider>
  );
}
