import { useTranslation } from 'react-i18next';
import { useDrop } from 'react-dnd';
import { PhononCard } from '../interfaces/interfaces';
import { HelpTooltip } from './HelpTooltip';
import { useFeature } from '../hooks/useFeature';
import { CardShadow } from './CardShadow';
import { Card } from './Card';
import { useContext } from 'react';
import { CardManagementContext } from '../contexts/CardManagementContext';

export const WalletSlot: React.FC<{
  card: PhononCard;
}> = ({ card }) => {
  const { t } = useTranslation();
  const { ENABLE_MOCK_CARDS } = useFeature();
  const { addCardsToState, isCardsMini, removeCardsFromState, phononCards } =
    useContext(CardManagementContext);

  const [{ isOver }, drop] = useDrop(() => ({
    accept: ['PhononCard'],
    drop: (item: PhononCard, monitor) => {
      monitor.getItem().InTray = false;
      addCardsToState([item]);

      // remove all remote cards
      removeCardsFromState(
        phononCards.filter((card: PhononCard) => card.IsRemote)
      );
    },
    collect: (monitor) => ({
      isOver: monitor.isOver(),
      canDrop: monitor.canDrop(),
    }),
  }));

  // only show card if not a mock card or if mock cards are enabled
  return (card.IsMock && ENABLE_MOCK_CARDS) || !card.IsMock ? (
    <div>
      <div
        ref={drop}
        className={'relative ' + (isCardsMini ? 'w-56 h-36 ' : 'w-80 h-52')}
      >
        {card.InTray ? <CardShadow isOver={isOver} /> : <Card card={card} />}
      </div>

      {card.IsMock && (
        <div
          className={
            'pt-px flex justify-end ' + (isCardsMini ? 'w-56' : 'w-80')
          }
        >
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
