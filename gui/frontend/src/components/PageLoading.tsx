import { useTranslation } from 'react-i18next';

export const PageLoading = () => {
  const { t } = useTranslation();

  return (
    <div className="bg-black w-full font-bandeins-sans-bold uppercase text-4xl text-white p-6 h-screen relative flex items-center justify-center">
      <div className="animate-bounce">
        <img
          className="w-12 inline mr-4"
          src="/assets/images/phonon-logo.png"
        />
        {t('Loading Phonon Manager...')}
      </div>
    </div>
  );
};
