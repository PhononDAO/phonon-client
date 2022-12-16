import React, { Suspense, useEffect } from 'react';
import { NotificationTray } from './components/NotificationTray';
import { Header } from './components/Header';
import { Stage } from './components/Stage';
import { PageLoading } from './components/PageLoading';
import { Mainnet, DAppProvider, Config, Goerli } from '@usedapp/core';
import { useTranslation } from 'react-i18next';
import localStorage from './utils/localStorage';
import 'console.history';
import { GlobalSettings } from './interfaces/interfaces';

const config: Config = {
  readOnlyChainId: Mainnet.chainId,
  readOnlyUrls: {
    [Mainnet.chainId]:
      'https://eth-mainnet.g.alchemy.com/v2/1vxQLStZLLkPSYFNv6xfG74858Zkw_v2',
    [Goerli.chainId]:
      'https://eth-goerli.g.alchemy.com/v2/ZoCdZaZgBRbzeRgXizrpHQRG7oDTQ1Kq',
  },
};

const App = () => {
  const { i18n, t } = useTranslation();
  const defaultSettings: GlobalSettings =
    localStorage.getConfigurableSettings();

  const changeLanguage = async (language) => {
    return await i18n.changeLanguage(language);
  };

  useEffect(() => {
    changeLanguage(defaultSettings.defaultLanguage)
      .then(() => {
        document.title = t('PHONON MANAGER');
      })
      .catch((err) => {
        console.log(err);
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <DAppProvider config={config}>
      <Suspense fallback={<PageLoading />}>
        <NotificationTray />
        <div className="w-full overflow-scroll flex flex-col h-screen bg-black relative">
          <Header />
          <Stage />
        </div>
      </Suspense>
    </DAppProvider>
  );
};

export default App;
