import { useTranslation } from 'react-i18next';
import { useDrag } from 'react-dnd';
import { PhononCard as Card } from '../classes/PhononCard';
import { HelpTooltip } from './HelpTooltip';
import { useFeature } from '../hooks/useFeature';
import { useState } from 'react';
import { CardShadow } from './CardShadow';
import { PhononCard } from './PhononCard';

interface DropResult {
  name: string;
}

export const WalletSlot: React.FC<{
  card: Card;
  isMini?: boolean;
}> = ({ card, isMini = false }) => {
  const { t } = useTranslation();
  const { ENABLE_MOCK_CARDS } = useFeature();
  const [thisCard, setThisCard] = useState(card);
  const props = { card, isMini, setThisCard };

  const [{ isDragging }, drag] = useDrag(() => ({
    type: 'PhononCard',
    item: thisCard,
    end: (item, monitor) => {
      const dropResult = monitor.getDropResult<DropResult>();
      if (item && dropResult) {
        // item.IsInTray = true;
      }
    },
    collect: (monitor) => ({
      isDragging: monitor.isDragging(),
      handlerId: monitor.getHandlerId(),
    }),
  }));

  // only show card if not a mock card or if mock cards are enabled
  return (thisCard.IsMock && ENABLE_MOCK_CARDS) || !thisCard.IsMock ? (
    <div
      ref={drag}
      data-testid={`PhononCard`}
      className={
        'transition-all  ' +
        (thisCard.IsLocked && !isDragging ? 'card-selected' : '')
      }
    >
      {isDragging || card.IsInTray ? <CardShadow /> : <PhononCard {...props} />}

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
