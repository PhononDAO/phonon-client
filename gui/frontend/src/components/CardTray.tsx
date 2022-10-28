import { useTranslation } from 'react-i18next';
import { useDrop } from 'react-dnd';

import { PhononCard } from './PhononCard';
import { PhononCard as Card } from '../classes/PhononCard';
import { useState } from 'react';

export const CardTray: React.FC<{
  card?: Card;
}> = ({ card = null }) => {
  const { t } = useTranslation();
  const [loadedCard, setLoadedCard] = useState<Card | null>(null);

  const [{ canDrop, isOver }, drop] = useDrop(() => ({
    accept: 'PhononCard',
    drop: (item: Card, monitor) => {
      console.log(monitor.getItem());

      item.IsInTray = true;
      item.IsLocked = false;
      setLoadedCard(item);
    },
    collect: (monitor) => ({
      isOver: monitor.isOver(),
      canDrop: monitor.canDrop(),
    }),
  }));

  // only show card if not a mock card or if mock cards are enabled
  return loadedCard?.IsInTray ? (
    <div className="w-80 h-52">
      <PhononCard card={loadedCard} />
    </div>
  ) : (
    <div
      ref={drop}
      className={
        'w-80 h-52 rounded-lg border border-4 overflow-hidden flex items-center justify-center text-xl text-white transition-all ' +
        (isOver
          ? 'border-green-500 bg-green-200'
          : 'border-dashed border-white bg-phonon-card bg-cover bg-no-repeat')
      }
    >
      Drop a card here
    </div>
  );
};
