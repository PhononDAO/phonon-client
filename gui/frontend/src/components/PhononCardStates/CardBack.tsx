import { ButtonGroup } from '@chakra-ui/react';
import { CardSettings } from '../PhononCardActions/CardSettings';
import { ViewPhonons } from '../PhononCardActions/ViewPhonons';
import { CloseCard } from '../PhononCardActions/CloseCard';
import { LockCard } from '../PhononCardActions/LockCard';
import { useTranslation } from 'react-i18next';
import { useContext } from 'react';
import { CardManagementContext } from '../../assets/contexts/CardManagementContext';

export const CardBack: React.FC<{
  card;
}> = ({ card }) => {
  const { t } = useTranslation();
  const { isCardsMini } = useContext(CardManagementContext);

  return (
    <div className="absolute z-40 w-full h-full p-2">
      <div
        className={
          'flex space-x-2 font-bandeins-sans-bold uppercase ' +
          (isCardsMini && !card.InTray ? 'text-sm' : 'text-md')
        }
      >
        <img
          className={'inline ' + (isCardsMini && !card.InTray ? 'w-6' : 'w-10')}
          src="/assets/images/phonon-logo.png"
        />{' '}
        <span className="text-white">PHONON</span>
      </div>
      {card.IsMock && (
        <div
          className={
            'absolute rotate-30 font-bandeins-sans-bold text-center text-white bg-red-600 py-px ' +
            (isCardsMini && !card.InTray
              ? 'w-48 top-5 -right-12 text-sm'
              : 'w-60 top-5 -right-16 text-md')
          }
        >
          {t('MOCK CARD')}
        </div>
      )}

      <div className="absolute bottom-0 left-0 w-full">
        <div
          className={
            'text-right text-sm text-white pr-1 ' +
            (isCardsMini && !card.InTray ? 'py-px' : 'py-2')
          }
        >
          {t('Contains ' + String(card.Phonons.length) + ' Phonons.')}
        </div>
        <div
          className={
            'bg-white z-50 pt-px px-2 ' +
            (isCardsMini && !card.InTray ? 'pb-px' : 'pb-2')
          }
        >
          <div
            className={
              'font-noto-sans-mono text-black ' +
              (isCardsMini && !card.InTray ? 'pb-px' : 'pb-2')
            }
          >
            <div
              className={isCardsMini && !card.InTray ? 'text-md' : 'text-base'}
            >
              {card.VanityName ? card.VanityName : card.CardId}
            </div>
            {card.VanityName && (
              <div className="text-xxs text-gray-400">{card.CardId}</div>
            )}
          </div>
          {card.ShowActions && (
            <ButtonGroup className="text-white" spacing={2}>
              {card.InTray ? (
                <CloseCard card={card} />
              ) : (
                <ViewPhonons card={card} />
              )}
              <CardSettings card={card} />
              <LockCard card={card} />
            </ButtonGroup>
          )}
        </div>
      </div>
    </div>
  );
};
