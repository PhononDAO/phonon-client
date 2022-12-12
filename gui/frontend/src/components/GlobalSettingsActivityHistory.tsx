import { Stack } from '@chakra-ui/react';
import { useTranslation } from 'react-i18next';
import localStorage from '../utils/localStorage';
import { DateTime } from 'luxon';
import { IonIcon } from '@ionic/react';
import { checkmarkCircle, closeCircle } from 'ionicons/icons';

export const GlobalSettingsActivityHistory: React.FC = () => {
  const { t } = useTranslation();
  const theme = {
    success: {
      styles: 'text-green-600 bg-green-200',
      icon: checkmarkCircle,
    },
    error: {
      styles: 'text-red-600 bg-red-200',
      icon: closeCircle,
    },
  };

  return (
    <>
      {localStorage.getActivityHistory().length > 0 && (
        <div className="mb-2">
          {t(
            'The following is your activity history, with the most recent first:'
          )}
        </div>
      )}
      <Stack className="h-144 overflow-scroll" spacing={3}>
        {localStorage
          .getActivityHistory()
          .reverse()
          .map((historyItem, key) => (
            <div
              key={key}
              className={
                'flex  items-center gap-x-2 px-4 py-2 rounded ' +
                String(theme[historyItem.type].styles)
              }
            >
              <IonIcon size="large" icon={theme[historyItem.type].icon} />
              <div className="w-full flex justify-between items-center gap-x-4">
                {historyItem.message}
                <span className="text-gray-400 text-sm whitespace-nowrap">
                  {DateTime.fromISO(historyItem.datetime).toRFC2822()}
                </span>
              </div>
            </div>
          ))}
        {localStorage.getActivityHistory().length === 0 && (
          <span className="text-xl text-gray-500 italic">
            {t('No activity history yet.')}
          </span>
        )}
      </Stack>
    </>
  );
};
