import React from 'react';
import { useTranslation, withTranslation } from 'react-i18next';

function App(i18n) {
  const { t } = useTranslation();

  const changeLanguage = async (language) => {
    await i18n.changeLanguage(language);
  };

  changeLanguage('fr-FR').catch((err) => {
    console.log(err);
  });

  return (
    <div className="App">
      <header className="text-4xl">{t('PHONON MANAGER GOES HERE')}</header>
    </div>
  );
}

export default withTranslation()(App);
