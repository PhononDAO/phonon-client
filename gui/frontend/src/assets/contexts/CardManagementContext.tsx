import { createContext, useState, ReactNode } from 'react';
import { usePhononCards } from '../../hooks/usePhononCards';

export const CardManagementContext = createContext(undefined);

export const CardManagementContextProvider = ({
  children,
  overrides,
}: {
  children: ReactNode;
  overrides?: { [key: string]: any };
}) => {
  const [isLoadingPhononCards, setIsLoadingPhononCards] = useState(false);
  const [isCardsMini, setIsCardsMini] = useState<boolean>(false);

  const [
    phononCards,
    addPhononCardsToState,
    removePhononCardsFromState,
    resetPhononCardsInState,
  ] = usePhononCards([]);

  const defaultContext = {
    isLoadingPhononCards,
    isCardsMini,
    setIsCardsMini,
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
