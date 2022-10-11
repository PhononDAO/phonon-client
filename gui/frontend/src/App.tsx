import React, { Suspense } from 'react';
import { useTranslation } from 'react-i18next';
import { NotificationTray } from './components/NotificationTray';
import { Stage } from './components/Stage';

const App = () => {
  const { t } = useTranslation();

  return (
    <Suspense fallback="loading">
      <div className="w-100 relative">
        <NotificationTray />
        <header className="text-4xl">{t('PHONON MANAGER GOES HERE')}</header>
        <Stage />
      </div>
    </Suspense>
  );
};

export default App;
