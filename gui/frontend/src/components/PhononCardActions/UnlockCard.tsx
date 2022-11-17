/* eslint-disable @typescript-eslint/no-unsafe-return */
import { PhononCard as Card } from '../../classes/PhononCard';
import { IonIcon } from '@ionic/react';
import { lockClosed } from 'ionicons/icons';
import { useContext } from 'react';
import { CardManagementContext } from '../../contexts/CardManagementContext';

export const UnlockCard: React.FC<{
  card: Card;
}> = ({ card }) => {
  const { isCardsMini, addPhononCardsToState } = useContext(
    CardManagementContext
  );

  return (
    <>
      <button
        onClick={() => {
          card.AttemptUnlock = true;
          addPhononCardsToState([card]);
        }}
        className="w-full z-50 text-amber-400 hover:text-amber-300 text-center"
      >
        <IonIcon
          className={
            'p-16 duration-150 ' +
            (isCardsMini && !card.InTray ? 'text-4xl' : 'text-6xl')
          }
          icon={lockClosed}
        />
      </button>
    </>
  );
};
