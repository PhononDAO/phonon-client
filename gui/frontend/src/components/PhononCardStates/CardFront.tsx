import { UnlockCard } from '../PhononCardActions/UnlockCard';
import { useTranslation } from 'react-i18next';
import { useContext } from 'react';
import { CardManagementContext } from '../../contexts/CardManagementContext';

export const CardFront: React.FC<{
  card;
  isMini?: boolean;
}> = ({ card, isMini = false }) => {
  const { t } = useTranslation();
  const { isCardsMini } = useContext(CardManagementContext);

  const showCardsMini = isMini || isCardsMini;

  return (
    <div className="absolute w-full h-full">
      <div className="absolute w-full h-full p-2 font-noto-sans-mono">
        <div
          className={'text-white ' + (showCardsMini ? 'text-md' : 'text-lg')}
        >
          {card.VanityName ? card.VanityName : card.CardId}
        </div>
        {card.VanityName && (
          <div className="text-xxs text-gray-400">{card.CardId}</div>
        )}
      </div>
      <img
        className={
          'absolute ' +
          (showCardsMini ? 'w-16 right-10 bottom-3' : 'w-24 right-12 bottom-4')
        }
        src="/assets/images/phonon-logo.png"
      />
      <div className="absolute h-full relative flex items-center">
        <div
          className={
            'absolute text-center font-bandeins-sans-bold text-white uppercase rotate-90 ' +
            (showCardsMini
              ? 'text-3xl -right-[46px]'
              : 'text-phonon-card -right-[74px]')
          }
        >
          PHONON
        </div>
        {<UnlockCard card={card} />}
      </div>
      {card.IsMock && (
        <div
          className={
            'absolute uppercase rotate-30 font-bandeins-sans-bold text-center text-white bg-red-600 py-px ' +
            (showCardsMini
              ? 'w-48 top-24 -left-12 text-sm'
              : 'w-60 top-40 -left-16 text-md')
          }
        >
          {t('mock card')}
        </div>
      )}
    </div>
  );
};
