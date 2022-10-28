import { createContext, useState, ReactNode } from 'react';
import { useRecords } from '../../hooks/useRecords';

export const CardManagementContext = createContext(undefined);

export const CardManagementContextProvider = ({
  children,
  overrides,
}: {
  children: ReactNode;
  overrides?: { [key: string]: any };
}) => {
  const [isLoadingPhononCards, setIsLoadingPhononCards] = useState(false);

  const [
    phononCards,
    addPhononCardsToState,
    removePhononCardsFromState,
    resetPhononCardsInState,
  ] = useRecords([]);

  const defaultContext = {
    isLoadingPhononCards,
    setIsLoadingPhononCards,
    phononCards,
    addPhononCardsToState,
    removePhononCardsFromState,
    resetPhononCardsInState,
  };

  return (
    <CardManagementContext.Provider value={{ ...defaultContext, ...overrides }}>
      {children}
    </CardManagementContext.Provider>
  );
};
