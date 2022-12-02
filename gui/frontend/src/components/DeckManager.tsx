import { CardDeck } from './CardDeck';
import filter from 'lodash/filter';
import { useContext, useEffect } from 'react';
import { CardManagementContext } from '../contexts/CardManagementContext';
import { PhononCard } from '../interfaces/interfaces';
import { maxTrays } from '../constants/Constants';

export const DeckManager = () => {
  const { phononCards, setIsCardsMini } = useContext(CardManagementContext);
  const cardsInTrays = filter(phononCards, (p: PhononCard) => p.InTray);

  useEffect(() => {
    setIsCardsMini(cardsInTrays.length > 0);
  }, [cardsInTrays, setIsCardsMini]);

  return (
    <div className="grid grid-cols-2 gap-x-12 sticky top-2">
      {Array(Math.min(Number(cardsInTrays.length + 1), maxTrays))
        .fill(null)
        .map((id, key) => (
          <CardDeck
            key={key}
            card={cardsInTrays[key]}
            canHaveRemote={key > 0}
          />
        ))}
    </div>
  );
};
