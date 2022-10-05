import React, { Suspense, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Toaster } from 'react-hot-toast';
import { Stage } from './components/Stage';

const App = () => {
  const { t, i18n } = useTranslation();

  // const changeLanguage = async (language) => {
  //   return await i18n.changeLanguage(language);
  // };

  // useEffect(() => {
  //   changeLanguage('fr-FR').catch((err) => {
  //     console.log(err);
  //   });
  // }, []);

  return (
    <Suspense fallback="loading">
      <div className="w-100 relative">
        <Toaster
          position="top-right"
          toastOptions={{
            duration: 80000,
            success: {
              className: 'border border-green-300',
            },
            error: {
              className: 'border border-red-300',
            },
          }}
        />
        <header className="text-4xl">{t('PHONON MANAGER GOES HERE')}</header>
        <Stage />
      </div>
    </Suspense>
  );
};

export default App;
