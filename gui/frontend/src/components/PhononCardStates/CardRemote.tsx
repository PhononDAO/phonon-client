import { Button } from '@chakra-ui/button';
import { IonIcon } from '@ionic/react';
import { cloud, flashOff } from 'ionicons/icons';
import { useTranslation } from 'react-i18next';

export const CardRemote: React.FC<{
  card?;
  unpair;
}> = ({ card, unpair }) => {
  const { t } = useTranslation();

  return (
    <div className="absolute z-40 w-full h-full p-2">
      <div className="flex space-x-2 font-bandeins-sans-bold uppercase text-xl">
        <span className="text-white">PHONON</span>
      </div>
      <div className="flex flex-col gap-y-2 py-4 px-2 items-center justify-center text-xl">
        <IonIcon icon={cloud} className="text-white text-5xl" />
        <div>
          <span className="block text-center text-white text-xl">
            {t('Paired Remote Card')}
          </span>
        </div>
      </div>
      <div className="absolute bottom-0 left-0 w-full">
        <div className="bg-white z-50 px-2 py-2 flex justify-between items-center">
          <div className="font-noto-sans-mono text-blue-900 text-lg">
            04e0d5eb884a73c0
          </div>
          <Button
            leftIcon={<IonIcon icon={flashOff} />}
            size="xs"
            colorScheme="red"
            onClick={unpair}
          >
            {t('Unpair')}
          </Button>
        </div>
      </div>
    </div>
  );
};
