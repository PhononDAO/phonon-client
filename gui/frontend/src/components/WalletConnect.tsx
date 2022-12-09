import { Button } from '@chakra-ui/react';
import { useEthers } from '@usedapp/core';
import { useTranslation } from 'react-i18next';
import { abbreviateHash } from '../utils/formatting';

export const WalletConnect: React.FC = () => {
  const { t } = useTranslation();
  const { activateBrowserWallet, deactivate, account } = useEthers();

  return (
    <div>
      {!account && (
        <Button colorScheme="purple" onClick={activateBrowserWallet}>
          {t('Connect Wallet')}
        </Button>
      )}
      {account && (
        <div className="text-right">
          <div className="text-white text-xs mb-1">
            {t('Connected as ') + abbreviateHash(account)}
          </div>
          <Button size="xs" colorScheme="gray" onClick={deactivate}>
            {t('Disconnect Wallet')}
          </Button>
        </div>
      )}
    </div>
  );
};
