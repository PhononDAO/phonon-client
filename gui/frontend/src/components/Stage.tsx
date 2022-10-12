import { useTranslation } from 'react-i18next';

export const Stage = () => {
  const { t } = useTranslation();

  return (
    <main className="bg-zinc-900 font-bandeins-sans text-lg text-white px-6 py-4 flex-grow">
      {t('STAGE HERE')}
    </main>
  );
};
