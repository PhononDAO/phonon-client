import { IonIcon } from '@ionic/react';
import { send } from 'ionicons/icons';
import { Pluralize } from 'pluralize-react';
import { useTranslation } from 'react-i18next';
import { Phonon, PhononCard } from '../interfaces/interfaces';
import { IncomingPhononTransferButton } from './IncomingPhononTransferButton';

export const IncomingTransferNotice: React.FC<{
  card: PhononCard;
}> = ({ card }) => {
  const { t } = useTranslation();

  const aPhonon = {
    Address: '0x7Ab7050217C76d729fa542161ca59Cb28484e0fa',
    ChainID: 43114,
    Denomination: '5008000000000000000',
    CurrencyType: 3,
    SourceCardId: '04e0d5eb884a73ce',
  } as Phonon;

  const bPhonon = {
    Address: '0x7Ab7050217C76d729fa542161ca59Cb28484bf8e',
    ChainID: 137,
    Denomination: '50600000000000000',
    CurrencyType: 2,
    SourceCardId: '04e0d5eb884a73cf',
  } as Phonon;

  const sourceCard = {
    CardId: '05d3d5ebcf4aa32c',
    IsRemote: false,
    InTray: true,
    IncomingTransferProposal: [aPhonon, bPhonon],
  } as PhononCard;

  return (
    <div className="flex gap-x-2 justify-between px-4 py-2 mb-4 w-full bg-white border-2 border-blue-600 rounded-md text-blue-600 font-bandeins-sans uppercase">
      <div className="flex gap-x-2 items-center">
        <div className="animate-bounce">
          <IonIcon icon={send} className="mt-2 -rotate-30 text-xl" />
        </div>
        <span>
          {t('Incoming transfer request: ')}
          <Pluralize
            count={sourceCard.IncomingTransferProposal.length}
            singular="phonon"
            className="font-bandeins-sans-bold text-xl"
          />
        </span>
      </div>
      <IncomingPhononTransferButton
        destinationCard={card}
        sourceCard={sourceCard}
      />
    </div>
  );
};
