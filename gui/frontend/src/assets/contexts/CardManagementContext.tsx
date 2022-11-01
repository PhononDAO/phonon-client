import { createContext, useState, ReactNode } from 'react';
import { useRecords } from '../../hooks/useRecords';
import { PhononCard } from '../../interfaces/interfaces';

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
  const [deckOneCard, setDeckOneCard] = useState<PhononCard | null>(null);
  const [deckTwoCard, setDeckTwoCard] = useState<PhononCard | null>(null);

  const [
    phononCards,
    addPhononCardsToState,
    removePhononCardsFromState,
    resetPhononCardsInState,
  ] = useRecords([]);

  const defaultContext = {
    isLoadingPhononCards,
    isCardsMini,
    setIsCardsMini,
    setIsLoadingPhononCards,
    phononCards,
    addPhononCardsToState,
    removePhononCardsFromState,
    resetPhononCardsInState,
    deckOneCard,
    setDeckOneCard,
    deckTwoCard,
    setDeckTwoCard,
  };

  return (
    <CardManagementContext.Provider value={{ ...defaultContext, ...overrides }}>
      {children}
    </CardManagementContext.Provider>
  );
};
