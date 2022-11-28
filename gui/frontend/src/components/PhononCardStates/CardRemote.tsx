import { Button } from '@chakra-ui/button';
import { IonIcon } from '@ionic/react';
import { cloud, flashOff } from 'ionicons/icons';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { PhononCard } from '../../interfaces/interfaces';
import { CardManagementContext } from '../../contexts/CardManagementContext';

export const CardRemote: React.FC<{
  unpair?;
  isMini?: boolean;
  showActions?: boolean;
}> = ({ unpair = false, isMini = false, showActions = true }) => {
  const { t } = useTranslation();
  const { phononCards } = useContext(CardManagementContext);

  // get the remote card, assuming it exists
  const remoteCards = phononCards.filter((card: PhononCard) => card.IsRemote);

  return (
    remoteCards.length > 0 && (
      <div className="absolute z-40 w-full h-full p-2">
        <div
          className={
            'flex space-x-2 font-bandeins-sans-bold uppercase' +
            (isMini ? 'text-sm' : ' text-xl')
          }
        >
          <span className="text-white">PHONON</span>
        </div>
        <div className="flex flex-col gap-y-2 py-4 px-2 items-center justify-center text-xl">
          <IonIcon
            icon={cloud}
            className={'text-white' + (isMini ? ' text-3xl' : ' text-5xl')}
          />
          <div>
            <span
              className={
                'block text-center text-white' +
                (isMini ? ' text-xs' : ' text-xl')
              }
            >
              {t('Paired Remote Card')}
            </span>
          </div>
        </div>
        <div className="absolute bottom-0 left-0 w-full">
          <div
            className={
              'bg-white z-50 px-2 flex justify-between items-center' +
              (isMini ? ' p-px' : ' py-2')
            }
          >
            <div
              className={
                'font-noto-sans-mono text-blue-900' +
                (isMini ? ' text-md' : ' text-base')
              }
            >
              {remoteCards[0].CardId}
            </div>
            {showActions && (
              <Button
                leftIcon={<IonIcon icon={flashOff} />}
                size="xs"
                colorScheme="red"
                onClick={unpair}
              >
                {t('Unpair')}
              </Button>
            )}
          </div>
        </div>
      </div>
    )
  );
};
