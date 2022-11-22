import React from 'react';
import { useTranslation } from 'react-i18next';
import { PhononCard } from '../classes/PhononCard';
import { PhononTransferDropzone } from './PhononTransferDropzone';

export const PhononTransferProposal: React.FC<{ card: PhononCard }> = ({
  card,
}) => {
  const { t } = useTranslation();

  return (
    <div className="text-xl px-6 py-2 mb-8 flex flex-col gap-y-2 items-center border-gray-400 text-gray-500 border-dashed border-4 rounded-md text-center font-bandeins-sans">
      <PhononTransferDropzone card={card} />
    </div>
  );
};
