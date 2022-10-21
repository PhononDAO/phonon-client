import { ButtonGroup } from '@chakra-ui/react';
import { useTranslation } from 'react-i18next';
import { useDrag } from 'react-dnd';
import { PhononCard as Card } from '../classes/PhononCard';
import { HelpTooltip } from './HelpTooltip';
import { useFeature } from '../hooks/useFeature';
import { LockCard } from './PhononCardActions/LockCard';
import { useState } from 'react';
import { CardSettings } from './PhononCardActions/CardSettings';
import { ViewPhonons } from './PhononCardActions/ViewPhonons';
import { CloseCard } from './PhononCardActions/CloseCard';
import { UnlockCard } from './PhononCardActions/UnlockCard';

interface DropResult {
  name: string;
}

export const PhononCard: React.FC<{
  card: Card;
  showActions?: boolean;
  isMini?: boolean;
}> = ({ card, showActions = true, isMini = false }) => {
  const { t } = useTranslation();
  const { ENABLE_MOCK_CARDS } = useFeature();

  const [{ isDragging }, drag] = useDrag(() => ({
    type: 'PhononCard',
    item: { card },
    end: (item, monitor) => {
      const dropResult = monitor.getDropResult<DropResult>();
      if (item && dropResult) {
        alert(`You dropped ${String(item.card.CardId)}!`);
      }
    },
    collect: (monitor) => ({
      isDragging: monitor.isDragging(),
      handlerId: monitor.getHandlerId(),
    }),
  }));

  const [thisCard, setThisCard] = useState(card);

  // only show card if not a mock card or if mock cards are enabled
  return (thisCard.IsMock && ENABLE_MOCK_CARDS) || !thisCard.IsMock ? (
    <div ref={drag} data-testid={`PhononCard`}>
      <div
        className={
          'flip-card duration-150 bg-transparent ' +
          (isMini ? 'w-56 h-36 ' : 'w-80 h-52 ') +
          (thisCard.IsLocked ? 'flip-card-locked' : '')
        }
      >
        <div className="flip-card-inner relative w-full h-full">
          <div className="flip-card-front w-full h-full absolute rounded-lg shadow-sm shadow-zinc-600 hover:shadow-md hover:shadow-zinc-500/60 bg-phonon-card bg-cover bg-no-repeat overflow-hidden">
            {/* BEGIN BACK OF CARD */}
            <div className="absolute z-40 w-full h-full p-2">
              <div
                className={
                  'flex space-x-2 font-bandeins-sans-bold uppercase ' +
                  (isMini ? 'text-sm' : 'text-md')
                }
              >
                <img
                  className={'inline ' + (isMini ? 'w-6' : 'w-10')}
                  src="/assets/images/phonon-logo.png"
                />{' '}
                <span className="text-white">PHONON</span>
              </div>
              {thisCard.IsMock && (
                <div
                  className={
                    'absolute rotate-30 font-bandeins-sans-bold text-center bg-red-600 py-px ' +
                    (isMini
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
                    (isMini ? 'py-px' : 'py-2')
                  }
                >
                  {t(
                    'Contains ' + String(thisCard.Phonons.length) + ' Phonons.'
                  )}
                </div>
                <div
                  className={
                    'bg-white z-50 pt-px px-2 ' + (isMini ? 'pb-px' : 'pb-2')
                  }
                >
                  <div
                    className={
                      'font-noto-sans-mono text-black ' +
                      (isMini ? 'pb-px' : 'pb-2')
                    }
                  >
                    <div className={isMini ? 'text-md' : 'text-base'}>
                      {thisCard.VanityName
                        ? thisCard.VanityName
                        : thisCard.CardId}
                    </div>
                    {thisCard.VanityName && (
                      <div className="text-xxs text-gray-400">
                        {thisCard.CardId}
                      </div>
                    )}
                  </div>
                  {showActions && (
                    <ButtonGroup className="text-white" spacing={2}>
                      {thisCard.IsActive ? (
                        <CloseCard setThisCard={setThisCard} isMini={isMini} />
                      ) : (
                        <ViewPhonons
                          setThisCard={setThisCard}
                          isMini={isMini}
                        />
                      )}
                      <CardSettings isMini={isMini} />
                      <LockCard setThisCard={setThisCard} isMini={isMini} />
                    </ButtonGroup>
                  )}
                </div>
              </div>
            </div>
            {/* END BACK OF CARD */}
          </div>
          <div className="flip-card-back w-full h-full absolute rounded-lg shadow-sm shadow-zinc-600 hover:shadow-md hover:shadow-zinc-500/60 bg-phonon-card bg-cover bg-no-repeat overflow-hidden">
            {/* BEGIN FRONT OF CARD */}
            <div className="absolute w-full h-full">
              <div className="absolute w-full h-full p-2 font-noto-sans-mono">
                <div
                  className={'text-white ' + (isMini ? 'text-md' : 'text-lg')}
                >
                  {thisCard.VanityName ? thisCard.VanityName : thisCard.CardId}
                </div>
                {thisCard.VanityName && (
                  <div className="text-xxs text-gray-400">
                    {thisCard.CardId}
                  </div>
                )}
              </div>
              <img
                className={
                  'absolute ' +
                  (isMini ? 'w-16 right-10 bottom-3' : 'w-24 right-12 bottom-4')
                }
                src="/assets/images/phonon-logo.png"
              />
              <div className="absolute h-full relative flex items-center">
                <div
                  className={
                    'absolute text-center font-bandeins-sans-bold text-white uppercase rotate-90 ' +
                    (isMini
                      ? 'text-3xl -right-[46px]'
                      : 'text-phonon-card -right-[74px]')
                  }
                >
                  PHONON
                </div>
                {thisCard.IsLocked && (
                  <UnlockCard
                    card={thisCard}
                    setThisCard={setThisCard}
                    isMini={isMini}
                  />
                )}
              </div>
              {thisCard.IsMock && (
                <div
                  className={
                    'absolute rotate-30 font-bandeins-sans-bold text-center bg-red-600 py-px ' +
                    (isMini
                      ? 'w-48 top-24 -left-12 text-sm'
                      : 'w-60 top-40 -left-16 text-md')
                  }
                >
                  {t('MOCK CARD')}
                </div>
              )}
            </div>
            {/* END FRONT OF CARD */}
          </div>
        </div>
      </div>

      {thisCard.IsMock && (
        <div className={'pt-px flex justify-end ' + (isMini ? 'w-56' : 'w-80')}>
          <HelpTooltip text={t('What is a mock card?')} theme="error">
            {t(
              'A mock card is a temporary card to test the platform. Mock cards are deleted, including all phonons, when this app is closed and have a different certificate than alpha and testnet phonon cards and therefore cannot communicate with them.'
            )}
          </HelpTooltip>
        </div>
      )}
    </div>
  ) : (
    <></>
  );
};
