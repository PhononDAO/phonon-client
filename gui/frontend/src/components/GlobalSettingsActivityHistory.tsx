import { Button, Stack } from '@chakra-ui/react';
import { useTranslation } from 'react-i18next';
import localStorage from '../utils/localStorage';
import { DateTime } from 'luxon';
import { IonIcon } from '@ionic/react';
import { checkmarkCircle, closeCircle } from 'ionicons/icons';
import { useState } from 'react';

export const GlobalSettingsActivityHistory: React.FC = () => {
  const { t } = useTranslation();
  const [history, setHistory] = useState(localStorage.getActivityHistory());

  const languageOptions = {
    'en-US': t('English'),
    'es-MX': t('Spanish'),
  };

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

  const clearHistory = () => {
    localStorage.setActivityHistory([]);
    setHistory([]);
  };

  return (
    <>
      {history.length > 0 && (
        <div className="mb-2 flex justify-between items-center">
          {t(
            'The following is your activity history, with the most recent first:'
          )}
          <Button size="sm" onClick={clearHistory}>
            {t('Clear activity history')}
          </Button>
        </div>
      )}
      <Stack className="h-144 overflow-scroll" spacing={3}>
        {history.reverse().map((historyItem, key) =>
          historyItem.type === 'seperator' ? (
            <div key={key} className="relative flex py-2 items-center">
              <div className="flex-grow border-t border-gray-300"></div>
              <span className="flex-shrink mx-4 text-gray-400 text-sm">
                {t(
                  historyItem.message,
                  Object.keys(historyItem.data).reduce((acc, key) => {
                    acc[key] = languageOptions[historyItem.data[key]];
                    return acc;
                  }, {})
                )}
              </span>
              <div className="flex-grow border-t border-gray-300"></div>
            </div>
          ) : (
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
          )
        )}
        {history.length === 0 && (
          <span className="text-xl text-gray-500 italic">
            {t('No activity history yet.')}
          </span>
        )}
      </Stack>
    </>
  );
};
