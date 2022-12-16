import { Phonon as iPhonon } from '../interfaces/interfaces';
import { IonIcon } from '@ionic/react';
import {
  helpCircle,
  syncCircle,
  shieldCheckmark,
  closeCircle,
} from 'ionicons/icons';
import { Phonon } from './Phonon';
import { useTranslation } from 'react-i18next';
import { Tooltip } from '@chakra-ui/react';

export const PhononValidator: React.FC<{
  phonon: iPhonon;
  isProposed?: boolean;
  showAction?: boolean;
  isTransferred?: boolean;
}> = ({
  phonon,
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
          {phonon.ValidationStatus === 'error' && (
            <>
              <Tooltip
                hasArrow
                label={t(
                  'Phonon includes a currency type that is unsupported.'
                )}
                aria-label={t('Phonon error message')}
                bg="red.600"
                fontSize="lg"
              >
                <span className="text-red-500 whitespace-nowrap flex gap-x-2 items-center">
                  <div className="relative h-6 mr-6">
                    <IonIcon
                      icon={closeCircle}
                      className="absolute ext-red-500 text-2xl animate-ping"
                    />
                    <IonIcon
                      icon={closeCircle}
                      className="absolute text-red-500 text-2xl"
                    />
                  </div>
                  {t('Validation Error')}
                </span>
              </Tooltip>
            </>
          )}
        </div>
      )}
      <Phonon phonon={phonon} isProposed={isProposed} showAction={showAction} />
    </div>
  );
};
