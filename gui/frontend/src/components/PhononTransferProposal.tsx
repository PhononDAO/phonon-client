import React from 'react';
import { useTranslation } from 'react-i18next';
import { PhononCard } from '../interfaces/interfaces';
import { PhononTransferDropzone } from './PhononTransferDropzone';
import { PhononTransferPayload } from './PhononTransferPayload';
import { SendPhononTransferButton } from './SendPhononTransferButton';

export const PhononTransferProposal: React.FC<{
  card: PhononCard;
}> = ({ card }) => {
  const { t } = useTranslation();

  return (
    <>
      {card.IncomingTransferProposal &&
        card.IncomingTransferProposal.length > 0 && (
          <div className="flex justify-end mb-2">
            <SendPhononTransferButton card={card} />
          </div>
        )}
      <div className="text-xl px-6 py-2 mb-8 flex flex-col gap-y-2 items-center border-gray-400 text-gray-500 border-dashed border-4 rounded-md text-center font-bandeins-sans">
        <PhononTransferPayload card={card} />
        <PhononTransferDropzone card={card} />
      </div>
    </>
  );
};
