import { CardDeck } from './CardDeck';
import { useContext } from 'react';
import { CardManagementContext } from '../assets/contexts/CardManagementContext';

export const DeckManager = () => {
  const { deckOneCard, deckTwoCard, setDeckOneCard, setDeckTwoCard } =
    useContext(CardManagementContext);

  return (
    <div className="grid grid-cols-2 gap-x-12 sticky top-2">
      <CardDeck card={deckOneCard} setDeckCard={setDeckOneCard} />
      {deckOneCard && (
        <CardDeck
          card={deckTwoCard}
          canHaveRemote={true}
          setDeckCard={setDeckTwoCard}
        />
      )}
    </div>
  );
};
