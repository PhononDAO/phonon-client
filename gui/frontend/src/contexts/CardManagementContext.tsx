import { createContext, useState, ReactNode, useEffect } from 'react';
import { usePhononCards } from '../hooks/usePhononCards';
import localStorage from '../utils/localStorage';

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
    addCardsToState,
    removeCardsFromState,
    resetCardsInState,
    getCardById,
    getCardPairingCode,
    addPhononsToCardState,
    removePhononsFromCardState,
    resetPhononsOnCardState,
    addPhononsToCardTransferState,
    removePhononsFromCardTransferState,
    resetPhononsOnCardTransferState,
    updateCardTransferStatusState,
  ] = usePhononCards(localStorage.getPhononCards() ?? []);

  const defaultContext = {
    isLoadingPhononCards,
    isCardsMini,
    setIsCardsMini,
    setIsLoadingPhononCards,
    phononCards,
    addCardsToState,
    removeCardsFromState,
    resetCardsInState,
    getCardById,
    getCardPairingCode,
    addPhononsToCardState,
    removePhononsFromCardState,
    resetPhononsOnCardState,
    addPhononsToCardTransferState,
    removePhononsFromCardTransferState,
    resetPhononsOnCardTransferState,
    updateCardTransferStatusState,
  };

  /**
   * Whenever `phononCards` data changes, it is persisted to `localStorage`
   */
  useEffect(() => {
    localStorage.setPhononCards(phononCards);
  }, [phononCards]);

  return (
    <CardManagementContext.Provider value={{ ...defaultContext, ...overrides }}>
      {children}
    </CardManagementContext.Provider>
  );
};
