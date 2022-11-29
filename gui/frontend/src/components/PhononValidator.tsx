import { Phonon as iPhonon, PhononCard } from '../interfaces/interfaces';
import { abbreviateHash, fromDecimals } from '../utils/formatting';
import { ChainIDTag } from './ChainIDTag';
import { CURRENCIES } from '../constants/Currencies';
import { useDrag } from 'react-dnd';
import { IonIcon } from '@ionic/react';
import { helpCircle } from 'ionicons/icons';
import { useContext } from 'react';
import { CardManagementContext } from '../contexts/CardManagementContext';
import { Phonon } from './Phonon';
import { useTranslation } from 'react-i18next';

interface DropResult {
  name: string;
  type: string;
}

export const PhononValidator: React.FC<{
  card: PhononCard;
  phonon: iPhonon;
  isProposed?: boolean;
  showAction?: boolean;
}> = ({ phonon, card, isProposed = false, showAction = false }) => {
  const { t } = useTranslation();

  return (
    <div className="flex bg-gray-200 rounded-full">
      <div className=" gap-y-2 text-yellow-600 flex items-center text-xs uppercase px-4">
        <IonIcon icon={helpCircle} className="text-2xl" />
        {t('Unvalidated')}
      </div>
      <Phonon
        card={card}
        phonon={phonon}
        isProposed={isProposed}
        showAction={showAction}
      />
    </div>
  );
};
