import React, { Suspense } from 'react';
import { NotificationTray } from './components/NotificationTray';
import { Header } from './components/Header';
import { Stage } from './components/Stage';
import { PageLoading } from './components/PageLoading';
import { Mainnet, DAppProvider, Config, Goerli } from '@usedapp/core';
import 'console.history';

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
