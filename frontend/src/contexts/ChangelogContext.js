import React, { createContext, useContext, useState } from 'react';

const ChangelogContext = createContext();

export const useChangelog = () => {
  const context = useContext(ChangelogContext);
  if (!context) {
    throw new Error('useChangelog must be used within a ChangelogProvider');
  }
  return context;
};

export const ChangelogProvider = ({ children }) => {
  const [manualTrigger, setManualTrigger] = useState(false);

  const openChangelog = () => {
    setManualTrigger(true);
  };

  const closeChangelog = () => {
    setManualTrigger(false);
  };

  return (
    <ChangelogContext.Provider value={{ 
      manualTrigger, 
      openChangelog, 
      closeChangelog 
    }}>
      {children}
    </ChangelogContext.Provider>
  );
};