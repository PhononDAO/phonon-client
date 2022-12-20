import { useTranslation } from 'react-i18next';
import { useDrop } from 'react-dnd';
import { Button } from '@chakra-ui/react';
import { Card } from './Card';
import { PhononCard } from '../interfaces/interfaces';
import { useContext, useState } from 'react';
import { IonIcon } from '@ionic/react';
import { cloudDownload } from 'ionicons/icons';
import { CardManagementContext } from '../contexts/CardManagementContext';
import { CardPairing } from './CardPairing';
import { CardRemote } from './PhononCardStates/CardRemote';

export const CardTray: React.FC<{
  card: PhononCard;
  canHaveRemote?: boolean;
}> = ({ card = null, canHaveRemote = false }) => {
  const { t } = useTranslation();
  const {
    phononCards,
    addCardsToState,
    resetPhononsOnCardTransferState,
    removeCardsFromState,
  } = useContext(CardManagementContext);
  const [showPairingOptions, setShowPairingOptions] = useState(false);

  const [{ isOver, canDrop }, drop] = useDrop(() => ({
    accept: ['PhononCard'],
    drop: (item: PhononCard, monitor) => {
      const itemCard = monitor.getItem();

      if (itemCard.IsLocked) {
        itemCard.AttemptUnlock = true;
        itemCard.FutureAction = 'InTray';
      } else {
        itemCard.InTray = true;
      }
      addCardsToState([itemCard]);
    },
    collect: (monitor) => ({
      isOver: monitor.isOver(),
      canDrop: monitor.canDrop(),
    }),
  }));

  const unpair = () => {
    setShowPairingOptions(false);

    resetPhononsOnCardTransferState(
      phononCards.filter((card: PhononCard) => card.IsRemote)[0],
      'OutgoingTransferProposal'
    );

    removeCardsFromState(
      phononCards.filter((card: PhononCard) => card.IsRemote)
    );
  };

  // only show card if not a mock card or if mock cards are enabled
  return card?.InTray ? (
    <>
      <div className="w-80 h-52">
        {card?.IsRemote ? (
          <div className="flip-card-back w-full h-full absolute rounded-lg shadow-sm shadow-zinc-600 hover:shadow-md hover:shadow-zinc-500/60 bg-phonon-card-blue bg-cover bg-no-repeat overflow-hidden">
            <CardRemote unpair={unpair} />
          </div>
        ) : (
          <Card card={card} forceLarge={true} />
        )}
      </div>
    </>
  ) : (
    <>
      {showPairingOptions ? (
        <CardPairing setShowPairingOptions={setShowPairingOptions} />
      ) : (
        <div
          ref={drop}
          className={
            'w-80 h-52 rounded-lg border border-4 overflow-hidden flex flex-col gap-y-2 items-center justify-center text-xl transition-all ' +
            (isOver && canDrop
              ? 'border-green-500 bg-green-200'
              : 'border-dashed border-white bg-phonon-card-gray bg-cover bg-no-repeat')
          }
        >
          <div className="text-white ">{t('Drop a card here')}</div>
          {canHaveRemote && (
            <>
              <div>
                <span className="block text-center text-white ">{t('OR')}</span>
              </div>
              <Button
                leftIcon={<IonIcon icon={cloudDownload} />}
                size="md"
                className="uppercase"
                onClick={() => {
                  setShowPairingOptions(!showPairingOptions);
                }}
              >
                {t('Pair Remote Card')}
              </Button>
            </>
          )}
        </div>
      )}
    </>
  );
};
