import { Phonon as iPhonon, PhononCard } from '../interfaces/interfaces';
import { IonIcon } from '@ionic/react';
import {
  helpCircle,
  syncCircle,
  shieldCheckmark,
  closeCircle,
} from 'ionicons/icons';
import { Phonon } from './Phonon';
import { useTranslation } from 'react-i18next';

export const PhononValidator: React.FC<{
  card: PhononCard;
  phonon: iPhonon;
  isProposed?: boolean;
  showAction?: boolean;
  isTransferred?: boolean;
}> = ({
  phonon,
  card,
  isProposed = false,
  showAction = false,
  isTransferred = false,
}) => {
  const { t } = useTranslation();

  return (
    <div className="flex bg-gray-200 rounded-full">
      {!isTransferred && (
        <div className="gap-y-2 flex items-center text-xs uppercase px-4">
          {phonon.ValidationStatus === 'unvalidated' && (
            <>
              <IonIcon icon={helpCircle} className="text-yellow-600 text-2xl" />
              <span className="ml-2 text-yellow-600">{t('Unvalidated')}</span>
            </>
          )}
          {phonon.ValidationStatus === 'validating' && (
            <>
              <IonIcon
                icon={syncCircle}
                className="text-blue-500 text-2xl animate-spin"
              />
              <span className="ml-2 text-blue-500">{t('Validating')}</span>
            </>
          )}
          {phonon.ValidationStatus === 'valid' && (
            <>
              <IonIcon
                icon={shieldCheckmark}
                className="text-green-500 text-2xl"
              />
              <span className="ml-2 text-green-500">{t('Valid')}</span>
            </>
          )}
          {phonon.ValidationStatus === 'not_valid' && (
            <>
              <IonIcon icon={closeCircle} className="text-red-500 text-2xl" />
              <span className="ml-2 text-red-500">{t('Not Valid')}</span>
            </>
          )}
        </div>
      )}
      <Phonon phonon={phonon} isProposed={isProposed} showAction={showAction} />
    </div>
  );
};
