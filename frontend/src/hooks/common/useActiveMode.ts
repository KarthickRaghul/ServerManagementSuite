// hooks/useActiveMode.ts
import { useState, useEffect } from 'react';

export type ModeType = 'server' | 'network';

export const useActiveMode = () => {
  const [activeMode, setActiveMode] = useState<ModeType>(() => {
    // Initialize from localStorage on first load
    const storedMode = localStorage.getItem('active_mode') as ModeType | null;
    return (storedMode === 'server' || storedMode === 'network') ? storedMode : 'server';
  });

  // Sync to localStorage whenever state changes
  useEffect(() => {
    localStorage.setItem('active_mode', activeMode);
    console.log('Mode state changed to:', activeMode); // Debug log
  }, [activeMode]);

  const updateActiveMode = (mode: ModeType) => {
    console.log('Updating mode from', activeMode, 'to', mode);
    setActiveMode(mode); // This will trigger re-renders automatically
  };

  return { activeMode, updateActiveMode };
};
