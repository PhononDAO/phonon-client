import React from 'react';
import { PhononCard } from '../interfaces/interfaces';
import { Phonon } from './Phonon';

export const PhononTransferPayload: React.FC<{
  destinationCard: PhononCard;
}> = ({ destinationCard }) => {
  return (
    <div className={'overflow-scroll gap-2 grid w-full'}>
      {destinationCard.OutgoingTransferProposal?.Phonons?.map((phonon, key) => (
        <Phonon
          key={key}
          phonon={phonon}
          destinationCard={destinationCard}
          isProposed={true}
          showAction={true}
        />
      ))}
    </div>
  );
};
