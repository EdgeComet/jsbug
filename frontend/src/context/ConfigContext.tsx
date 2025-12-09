import { createContext, useContext, useState, ReactNode } from 'react';
import type { AppConfig, PanelConfig } from '../types/config';

const defaultPanelConfig: PanelConfig = {
  jsEnabled: true,
  userAgent: 'chrome-mobile',
  timeout: 15,
  waitFor: 'networkIdle',
  blocking: {
    imagesMedia: false,
    css: false,
    trackingScripts: true, // Always checked
  },
};

const defaultConfig: AppConfig = {
  left: {
    ...defaultPanelConfig,
    jsEnabled: true,
    userAgent: 'chrome-mobile',
    timeout: 15,
    waitFor: 'networkIdle',
  },
  right: {
    ...defaultPanelConfig,
    jsEnabled: false,
    userAgent: 'chrome-mobile',
    timeout: 10,
    waitFor: 'load',
  },
};

interface ConfigContextType {
  config: AppConfig;
  setConfig: (config: AppConfig) => void;
  updateLeftConfig: (updates: Partial<PanelConfig>) => void;
  updateRightConfig: (updates: Partial<PanelConfig>) => void;
}

const ConfigContext = createContext<ConfigContextType | undefined>(undefined);

interface ConfigProviderProps {
  children: ReactNode;
}

export function ConfigProvider({ children }: ConfigProviderProps) {
  const [config, setConfig] = useState<AppConfig>(defaultConfig);

  const updateLeftConfig = (updates: Partial<PanelConfig>) => {
    setConfig(prev => ({
      ...prev,
      left: { ...prev.left, ...updates },
    }));
  };

  const updateRightConfig = (updates: Partial<PanelConfig>) => {
    setConfig(prev => ({
      ...prev,
      right: { ...prev.right, ...updates },
    }));
  };

  return (
    <ConfigContext.Provider value={{ config, setConfig, updateLeftConfig, updateRightConfig }}>
      {children}
    </ConfigContext.Provider>
  );
}

export function useConfig() {
  const context = useContext(ConfigContext);
  if (context === undefined) {
    throw new Error('useConfig must be used within a ConfigProvider');
  }
  return context;
}

export { defaultConfig };
