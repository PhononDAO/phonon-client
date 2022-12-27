import { useTranslation } from 'react-i18next';

export const Header = () => {
  const { t } = useTranslation();

  return (
    <header className="text-4xl px-6 py-4 text-white font-bandeins-sans-bold">
      {t('PHONON MANAGER')}
    </header>
  );
};
