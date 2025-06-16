// utils/metricsUtils.ts
export const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };
  
  export const formatPercentage = (value: number): string => {
    return `${Math.round(value * 100) / 100}%`;
  };
  
  export const getChangeType = (value: number, threshold: { warning: number; critical: number }): 'positive' | 'negative' => {
    if (value >= threshold.critical) return 'negative';
    if (value >= threshold.warning) return 'negative';
    return 'positive';
  };
  
  export const getMetricIcon = (type: string): string => {
    const icons: { [key: string]: string } = {
      cpu: '🖥️',
      ram: '💾',
      disk: '💽',
      network: '🌐'
    };
    return icons[type] || '📊';
  };
  
  export const calculatePreviousValue = (current: number): number => {
    // Simulate previous value for change calculation
    // In a real app, you'd store historical data
    const variation = (Math.random() - 0.5) * 10; // Random variation between -5% and +5%
    return Math.max(0, current + variation);
  };
  
  export const calculateChange = (current: number, previous: number): string => {
    if (previous === 0) return '+0%';
    
    const change = ((current - previous) / previous) * 100;
    const sign = change >= 0 ? '+' : '';
    return `${sign}${Math.round(change * 100) / 100}%`;
  };
  