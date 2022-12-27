import { IonIcon } from '@ionic/react';
import { send } from 'ionicons/icons';
import { Pluralize } from 'pluralize-react';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { CardManagementContext } from '../contexts/CardManagementContext';
import { PhononCard } from '../interfaces/interfaces';
import { IncomingPhononTransferButton } from './IncomingPhononTransferButton';

export const IncomingTransferNotice: React.FC<{
  card: PhononCard;
}> = ({ card }) => {
  const { t } = useTranslation();
  const { getCardById } = useContext(CardManagementContext);

  const sourceCard = getCardById(
    card?.IncomingTransferProposal?.Phonons[0].SourceCardId
  );

  return (
    <div className="flex gap-x-2 justify-between px-4 py-2 mb-4 w-full bg-white border-2 border-blue-600 rounded-md text-blue-600 font-bandeins-sans uppercase">
      <div className="flex gap-x-2 items-center">
        <div className="animate-bounce">
          <IonIcon icon={send} className="mt-2 -rotate-30 text-xl" />
        </div>
        <span>
          {t('Incoming transfer request: ')}
          <Pluralize
            count={card?.IncomingTransferProposal?.Phonons?.length}
            singular="phonon"
            className="font-bandeins-sans-bold text-xl"
          />
        </span>
      </div>
      <IncomingPhononTransferButton
        sourceCard={getCardById(sourceCard?.CardId)}
        destinationCard={card}
      />
    </div>
  );
};
