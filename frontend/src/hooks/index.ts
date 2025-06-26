// hooks/index.ts
// Auth hooks
export { useAuth } from "./auth/useAuth";
export { useRole } from "./auth/useRole";

// Server hooks
export { useServerConfiguration } from "./server/useServerConfiguration";
export { useServerManagement } from "./server/useServerManagement";
export { useServerOverview } from "./server/useServerOverview";
export { useCommandExecution } from "./server/useCommandExecution";
export { useSecurityManagement } from "./server/useSecurityManagement";
export { useHealthMetrics } from "./server/useHealthMetrics";

// Common hooks
export { useActiveDevice } from "./common/useActiveDevice";
export { useActiveMode } from "./common/useActiveMode";
