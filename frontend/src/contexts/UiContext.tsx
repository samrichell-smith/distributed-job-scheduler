import React, { createContext, useContext, useState } from 'react';

type UiContextValue = {
  screenshotMode: boolean;
  setScreenshotMode: (v: boolean) => void;
};

const UiContext = createContext<UiContextValue | undefined>(undefined);

export const UiProvider = ({ children }: { children: React.ReactNode }) => {
  const [screenshotMode, setScreenshotMode] = useState(false);
  return (
    <UiContext.Provider value={{ screenshotMode, setScreenshotMode }}>
      {children}
    </UiContext.Provider>
  );
};

export function useUi() {
  const ctx = useContext(UiContext);
  if (!ctx) throw new Error('useUi must be used within UiProvider');
  return ctx;
}

export default UiContext;
