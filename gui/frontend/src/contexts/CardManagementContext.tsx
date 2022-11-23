import { createContext, useState, ReactNode } from 'react';
import { usePhononCards } from '../hooks/usePhononCards';
import { usePhonons } from '../hooks/usePhonons';

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

  const [
    phononsOnCards,
    addCardPhononsToState,
    removeCardPhononsFromState,
    resetCardPhononsInState,
  ] = usePhonons({});

  const defaultContext = {
    isLoadingPhononCards,
    isCardsMini,
    setIsCardsMini,
    setIsLoadingPhononCards,
    phononCards,
    addPhononCardsToState,
    removePhononCardsFromState,
    resetPhononCardsInState,
    phononsOnCards,
    addCardPhononsToState,
    removeCardPhononsFromState,
    resetCardPhononsInState,
  };

  return (
    <CardManagementContext.Provider value={{ ...defaultContext, ...overrides }}>
      {children}
    </CardManagementContext.Provider>
  );
};
