import { useContext } from 'react';
import { useDrag } from 'react-dnd';
import { CardManagementContext } from '../assets/contexts/CardManagementContext';
import { PhononCard } from '../classes/PhononCard';

import { CardBack } from './PhononCardStates/CardBack';
import { CardFront } from './PhononCardStates/CardFront';

interface DropResult {
  name: string;
}

export const Card: React.FC<{
  card: PhononCard;
}> = ({ card }) => {
  const { isCardsMini } = useContext(CardManagementContext);

  const [{ isDragging }] = useDrag(() => ({
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

  console.log(card);

  // only show card if not a mock card or if mock cards are enabled
  return (
    <div
      className={
        'transition-all flip-card duration-150 bg-transparent ' +
        (isCardsMini && !card.TrayId ? 'w-56 h-36 ' : 'w-80 h-52 ') +
        (card.IsLocked ? 'flip-card-locked' : '')
      }
    >
      <div className="flip-card-inner relative w-full h-full">
        <div className="flip-card-front w-full h-full absolute rounded-lg shadow-sm shadow-zinc-600 hover:shadow-md hover:shadow-zinc-500/60 bg-phonon-card bg-cover bg-no-repeat overflow-hidden">
          {isDragging ? (
            !card.IsLocked && <CardBack card={card} />
          ) : (
            <CardBack card={card} />
          )}
        </div>
        <div className="flip-card-back w-full h-full absolute rounded-lg shadow-sm shadow-zinc-600 bg-phonon-card bg-cover bg-no-repeat overflow-hidden">
          {isDragging ? (
            card.IsLocked && <CardFront card={card} />
          ) : (
            <CardFront card={card} />
          )}
        </div>
      </div>
    </div>
  );
};
