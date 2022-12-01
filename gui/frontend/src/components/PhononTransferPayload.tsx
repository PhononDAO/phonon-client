import React from 'react';
import { PhononCard } from '../interfaces/interfaces';
import { Phonon } from './Phonon';

export const PhononTransferPayload: React.FC<{
  card: PhononCard;
}> = ({ card }) => {
  return (
    <div className={'overflow-scroll gap-2 grid w-full'}>
      {card.OutgoingTransferProposal?.Phonons?.map((phonon, key) => (
        <Phonon
          key={key}
          phonon={phonon}
          card={card}
          isProposed={true}
          showAction={true}
        />
      ))}
    </div>
  );
};
