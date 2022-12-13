/* eslint-disable @typescript-eslint/no-unsafe-return */
import { Phonon } from '../interfaces/interfaces';
import { IonIcon } from '@ionic/react';
import {
  closeCircle,
  helpCircle,
  shieldCheckmark,
  syncCircle,
} from 'ionicons/icons';
import { useTranslation } from 'react-i18next';
import { Button, Tooltip } from '@chakra-ui/react';
import { useState } from 'react';

export const PhononValidationStatus: React.FC<{
  phonon: Phonon;
}> = ({ phonon }) => {
  const { t } = useTranslation();
  const [status, setStatus] = useState<string>(phonon.ValidationStatus);

  const statusProperties = {
    unvalidated: {
      textColor: 'text-yellow-600',
      icon: helpCircle,
      label: t('Unvalidated'),
    },
    validating: {
      textColor: 'text-blue-500',
      icon: syncCircle,
      label: t('Validating'),
    },
    valid: {
      textColor: 'text-green-500',
      icon: shieldCheckmark,
      label: t('Valid'),
    },
    error: {
      textColor: 'text-red-500',
      icon: closeCircle,
      label: t('Error'),
    },
  };

  const startValidation = () => {
    // queue the validation
    setStatus('validating');

    // simulate validation
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const promise = new Promise((resolve) => {
      setTimeout(() => {
        resolve('validating');
      }, 3000);
    }).then(() => {
      // let's randomw determine the validation result
      setStatus(Math.round(Math.random()) ? 'valid' : 'error');
    });
  };

  return (
    <span
      className={
        'flex gap-x-2 items-center justify-end ' +
        String(statusProperties[status].textColor)
      }
    >
      {status === 'error' ? (
        <>
          <Tooltip
            hasArrow
            label={t('Phonon includes a currency type that is unsupported.')}
            aria-label={t('Phonon error message')}
            bg="red.600"
            fontSize="lg"
          >
            <span className="flex gap-x-2 items-center">
              <IonIcon
                icon={statusProperties[status].icon}
                className="text-xl"
              />
              <span>{statusProperties[status].label}</span>
            </span>
          </Tooltip>
          <Button size="xs" colorScheme="blue" onClick={startValidation}>
            {t('Re-validate')}
          </Button>
        </>
      ) : (
        <>
          <IonIcon
            icon={statusProperties[status].icon}
            className={
              'text-xl ' + (status === 'validating' ? 'animate-spin' : '')
            }
          />
          <span>{statusProperties[status].label}</span>
          {status === 'unvalidated' && (
            <Button size="xs" colorScheme="blue" onClick={startValidation}>
              {t('Validate')}
            </Button>
          )}
        </>
      )}
    </span>
  );
};
