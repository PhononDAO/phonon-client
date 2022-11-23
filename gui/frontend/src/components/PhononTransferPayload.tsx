import React from 'react';
import { useTranslation } from 'react-i18next';
import { PhononCard } from '../interfaces/interfaces';
import { Phonon } from './Phonon';

export const PhononTransferPayload: React.FC<{
  card: PhononCard;
}> = ({ card }) => {
  const { t } = useTranslation();

  return (
    <div className={'overflow-scroll gap-2 grid w-full'}>
      {card.IncomingTransferProposal?.map((phonon, key) => (
        <Phonon key={key} phonon={phonon} card={card} isProposed={true} />
      ))}
    </div>
  );
};
