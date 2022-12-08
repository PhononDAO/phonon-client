import { useTranslation } from 'react-i18next';
import { WalletConnect } from './WalletConnect';

export const Header = () => {
  const { t } = useTranslation();

  return (
    <header className="px-6 py-4 flex justify-between">
      <span className="text-4xl text-white font-bandeins-sans-bold">
        {t('PHONON MANAGER')}
      </span>
      <WalletConnect />
    </header>
  );
};
