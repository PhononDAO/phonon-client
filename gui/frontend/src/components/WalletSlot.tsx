import { useTranslation } from 'react-i18next';
import { useDrag } from 'react-dnd';
import { PhononCard } from '../classes/PhononCard';
import { HelpTooltip } from './HelpTooltip';
import { useFeature } from '../hooks/useFeature';
import { CardShadow } from './CardShadow';
import { Card } from './Card';
import { useContext } from 'react';
import { CardManagementContext } from '../assets/contexts/CardManagementContext';

interface DropResult {
  name: string;
}

export const WalletSlot: React.FC<{
  card: PhononCard;
}> = ({ card }) => {
  const { t } = useTranslation();
  const { ENABLE_MOCK_CARDS } = useFeature();
  const { isCardsMini } = useContext(CardManagementContext);

  const [{ isDragging }, drag] = useDrag(() => ({
    type: 'PhononCard',
    item: card,
    end: (item, monitor) => {
      const dropResult = monitor.getDropResult<DropResult>();
      if (item && dropResult) {
        // item.TrayId = true;
      }
    },
    collect: (monitor) => ({
      isDragging: monitor.isDragging(),
      handlerId: monitor.getHandlerId(),
    }),
  }));

  // only show card if not a mock card or if mock cards are enabled
  return (card.IsMock && ENABLE_MOCK_CARDS) || !card.IsMock ? (
    <div
      ref={drag}
      data-testid={`PhononCard`}
      className={
        'transition-all  ' +
        (card.IsLocked && !isDragging ? 'card-selected' : '')
      }
    >
      {isDragging || card.TrayId ? <CardShadow /> : <Card card={card} />}

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
