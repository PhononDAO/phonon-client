import React, { Suspense, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

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
      <div className="App">
        <header className="text-4xl">{t('PHONON MANAGER GOES HERE')}</header>
      </div>
    </Suspense>
  );
};

export default App;
