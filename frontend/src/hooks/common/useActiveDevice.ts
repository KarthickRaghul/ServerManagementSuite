// hooks/common/useActiveDevice.ts
import { useAppContext } from "../../context/AppContext";

export const useActiveDevice = () => {
  const {
    activeDevice,
    devices,
    devicesLoading: loading,
    devicesError: error,
    updateActiveDevice,
  } = useAppContext();

  const getHostForRequest = () => {
    return activeDevice ? { host: activeDevice.ip } : {};
  };

  return {
    activeDevice,
    devices,
    loading,
    error,
    updateActiveDevice,
    getHostForRequest,
  };
};
