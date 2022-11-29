import { useContext } from 'react';
import { useDrag } from 'react-dnd';
import { useDisclosure } from '@chakra-ui/react';
import { CardManagementContext } from '../contexts/CardManagementContext';
import { PhononCard } from '../interfaces/interfaces';
import { ModalUnlockCard } from './ModalUnlockCard';
import { CardBack } from './PhononCardStates/CardBack';
import { CardFront } from './PhononCardStates/CardFront';
import { CardRemote } from './PhononCardStates/CardRemote';

interface DropResult {
  name: string;
  type: string;
}

export const Card: React.FC<{
  card: PhononCard;
  isMini?: boolean;
  showActions?: boolean;
}> = ({ card, isMini = false, showActions = true }) => {
  const { onClose } = useDisclosure();
  const { isCardsMini } = useContext(CardManagementContext);

  const [{ isDragging }, drag] = useDrag(() => ({
    type: 'PhononCard',
    name: card.CardId,
    item: card,
    end: (item, monitor) => {
      const dropResult = monitor.getDropResult<DropResult>();
      if (item && dropResult) {
        // item.TrayId = true;
      }
    },
    collect: (monitor) => ({
      isDragging: monitor.isDragging(),
    }),
  }));

  // only show card if not a mock card or if mock cards are enabled
  return (
    <>
      <div
        ref={drag}
        className={
          'opacity-100 absolute transition-all flip-card duration-150 bg-transparent ' +
          ((isCardsMini && !card.InTray) || isMini
            ? 'w-56 h-36 '
            : 'w-80 h-52') +
          (card.IsLocked ? ' flip-card-locked ' : '') +
          (card.InTray ? '' : ' flip-card-tilt')
        }
      >
        {isDragging ? (
          <div className="flip-card-inner relative w-full h-full">
            {!card.IsLocked ? (
              <div
                className={
                  'flip-card-front w-full h-full absolute rounded-lg shadow-sm shadow-zinc-600 bg-phonon-card bg-cover bg-no-repeat overflow-hidden' +
                  (showActions
                    ? ' hover:shadow-md hover:shadow-zinc-500/60'
                    : '')
                }
              >
                <CardBack
                  card={card}
                  isMini={isMini}
                  showActions={showActions}
                />
              </div>
            ) : (
              <div className="flip-card-back w-full h-full absolute rounded-lg shadow-sm shadow-zinc-600 bg-phonon-card bg-cover bg-no-repeat overflow-hidden">
                <CardFront card={card} isMini={isMini} />
              </div>
            )}
          </div>
        ) : (
          <div className="flip-card-inner relative w-full h-full">
            <div
              className={
                'flip-card-front w-full h-full absolute rounded-lg shadow-sm shadow-zinc-600 bg-cover bg-no-repeat overflow-hidden' +
                (card.IsRemote ? ' bg-phonon-card-blue' : ' bg-phonon-card') +
                (showActions ? ' hover:shadow-md hover:shadow-zinc-500/60' : '')
              }
            >
              {card.IsRemote ? (
                <CardRemote isMini={isMini} showActions={showActions} />
              ) : (
                <CardBack
                  card={card}
                  isMini={isMini}
                  showActions={showActions}
                />
              )}
            </div>
            <div className="flip-card-back w-full h-full absolute rounded-lg shadow-sm shadow-zinc-600 bg-phonon-card bg-cover bg-no-repeat overflow-hidden">
              <CardFront card={card} isMini={isMini} />
            </div>
          </div>
        )}
      </div>
      <ModalUnlockCard
        isOpen={card.AttemptUnlock}
        onClose={onClose}
        card={card}
      />
    </>
  );
};
