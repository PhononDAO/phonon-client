import React from 'react';
import { PhononCard } from '../interfaces/interfaces';
import { IncomingTransferNotice } from './IncomingTransferNotice';
import { PhononTransferDropzone } from './PhononTransferDropzone';
import { PhononTransferPayload } from './PhononTransferPayload';
import { OutgoingPhononTransferButton } from './OutgoingPhononTransferButton';

export const PhononTransferProposal: React.FC<{
  card: PhononCard;
}> = ({ card }) => {
  return (
    <>
      {true && <IncomingTransferNotice card={card} />}
      {card.IncomingTransferProposal &&
        card.IncomingTransferProposal.length > 0 && (
          <div className="flex justify-end mb-2">
            <OutgoingPhononTransferButton destinationCard={card} />
          </div>
        )}
      <div className="text-xl px-6 py-2 mb-8 flex flex-col gap-y-2 items-center border-gray-400 text-gray-500 border-dashed border-4 rounded-md text-center font-bandeins-sans">
        <PhononTransferPayload card={card} />
        <PhononTransferDropzone card={card} />
      </div>
    </>
  );
};
