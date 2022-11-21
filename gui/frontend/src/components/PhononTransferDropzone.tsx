import React from 'react';
import { useTranslation } from 'react-i18next';

export const PhononTransferDropzone = () => {
  const { t } = useTranslation();

  return (
    <div className="text-xl px-6 py-8 mb-8 flex flex-col gap-y-2 items-center border-gray-400 text-gray-500 border-dashed border-4 rounded-md text-center font-bandeins-sans">
      {t('Drag-n-drop Phonons from another card here to stage a transfer.')}
    </div>
  );
};
