import React, { useContext } from 'react';
import { useDrop } from 'react-dnd';
import { useTranslation } from 'react-i18next';
import { PhononCard, Phonon } from '../interfaces/interfaces';
import { CardManagementContext } from '../contexts/CardManagementContext';

export const PhononTransferDropzone: React.FC<{ card: PhononCard }> = ({
  card,
}) => {
  const { t } = useTranslation();
  const { phononCards, addPhononsToCardTransferState } = useContext(
    CardManagementContext
  );

  const otherPhononCards = phononCards.filter(
    (thisCard: PhononCard) => thisCard.CardId !== card.CardId && thisCard.InTray
  );

  const [{ isOver }, drop] = useDrop(() => ({
    accept: otherPhononCards.map((card: PhononCard) => 'Phonon-' + card.CardId),
    drop: (item: Phonon, monitor) => {
      const phonon = monitor.getItem();

      // let's add the phonon to a transfer proposal for the destination card
      addPhononsToCardTransferState(card, [phonon], 'OutgoingTransferProposal');
    },
    collect: (monitor) => ({
      isOver: monitor.isOver(),
      canDrop: monitor.canDrop(),
    }),
  }));

  return (
    <div
      ref={drop}
      className={
        'text-xl w-full py-6 flex flex-col gap-y-2 items-center text-center rounded-md font-bandeins-sans' +
        (isOver ? ' bg-green-200' : '')
      }
    >
      {t('Drag-n-drop Phonons from another card here to stage a transfer.')}
    </div>
  );
};
